/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"context"
	"strings"
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kmapi "kmodules.xyz/client-go/api/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

func natsRule(cluster string, role psapi.DCRole, idx ...int32) psapi.DistributionRule {
	return psapi.DistributionRule{ClusterName: cluster, ReplicaIndices: idx, Role: role}
}

func natsPolicy(name string, rules ...psapi.DistributionRule) *psapi.PlacementPolicy {
	return &psapi.PlacementPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: psapi.PlacementPolicySpec{ClusterSpreadConstraint: &psapi.ClusterSpreadConstraint{
			DistributionRules: rules,
		}},
	}
}

func distributedNATS(policy string) *olddbapi.NATS {
	db := newNATS()
	db.Spec.Distributed = true
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: policy}
	db.SetDefaults(nil)
	return db
}

func validate(t *testing.T, db *olddbapi.NATS, objs ...runtime.Object) ([]string, string) {
	t.Helper()
	w := natsTestWebhook(t, objs...)
	warnings, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
	if len(errs) == 0 {
		return warnings, ""
	}
	return warnings, errs.ToAggregate().Error()
}

func wantErr(t *testing.T, got, want string) {
	t.Helper()
	if want == "" {
		if got != "" {
			t.Fatalf("unexpected error: %s", got)
		}
		return
	}
	if !strings.Contains(got, want) {
		t.Fatalf("want error containing %q, got %q", want, got)
	}
}

func hasWarning(warnings []string, sub string) bool {
	for _, w := range warnings {
		if strings.Contains(w, sub) {
			return true
		}
	}
	return false
}

func TestNATSGatewayNameEqualsClusterName(t *testing.T) {
	db := newNATS()
	db.Spec.Gateway = &olddbapi.NATSGatewaySpec{Name: "other"}
	_, got := validate(t, db)
	wantErr(t, got, "gateway name must equal the cluster name")

	db.Spec.Gateway.Name = db.Name
	_, got = validate(t, db)
	wantErr(t, got, "")

	// defaulting fills the name
	db = newNATS()
	db.Spec.Gateway = &olddbapi.NATSGatewaySpec{}
	db.SetDefaults(nil)
	if db.Spec.Gateway.Name != db.Name {
		t.Fatalf("gateway name should default to %q, got %q", db.Name, db.Spec.Gateway.Name)
	}
}

func TestNATSGatewayRemotes(t *testing.T) {
	cases := []struct {
		name    string
		remote  olddbapi.NATSRemoteGateway
		wantErr string
	}{
		{"urls", olddbapi.NATSRemoteGateway{Name: "dc-b", URLs: []string{"nats://b.example:7222"}}, ""},
		{"natsRef", olddbapi.NATSRemoteGateway{Name: "dc-b", NATSRef: &kmapi.ObjectReference{Name: "b"}}, ""},
		{"neither", olddbapi.NATSRemoteGateway{Name: "dc-b"}, "one of urls and natsRef is required"},
		{"both", olddbapi.NATSRemoteGateway{Name: "dc-b", URLs: []string{"nats://b:7222"}, NATSRef: &kmapi.ObjectReference{Name: "b"}}, "mutually exclusive"},
		{"no name", olddbapi.NATSRemoteGateway{URLs: []string{"nats://b:7222"}}, "remote gateway name is required"},
		{"self", olddbapi.NATSRemoteGateway{Name: "nats", URLs: []string{"nats://b:7222"}}, "can not list itself"},
		{"bad scheme", olddbapi.NATSRemoteGateway{Name: "dc-b", URLs: []string{"http://b:7222"}}, "one of the schemes"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := newNATS()
			db.Spec.Gateway = &olddbapi.NATSGatewaySpec{Remotes: []olddbapi.NATSRemoteGateway{c.remote, {Name: "dc-c", URLs: []string{"nats://c:7222"}}}}
			db.SetDefaults(nil)
			_, got := validate(t, db)
			wantErr(t, got, c.wantErr)
		})
	}
}

func TestNATSGatewayTLSSwitchesNeedTLS(t *testing.T) {
	db := newNATS()
	f := false
	db.Spec.GatewayTLS = &f
	_, got := validate(t, db)
	wantErr(t, got, "spec.gatewayTLS can only be set together with spec.tls")

	db = newNATS()
	db.Spec.LeafNodeTLS = &f
	_, got = validate(t, db)
	wantErr(t, got, "spec.leafNodeTLS can only be set together with spec.tls")
}

func TestNATSLeafNodeJetStreamNeedsDomain(t *testing.T) {
	remote := olddbapi.NATSLeafRemote{URLs: []string{"nats-leaf://hub:7422"}, Account: "APP"}
	db := newNATS()
	db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{Remotes: []olddbapi.NATSLeafRemote{remote}}
	_, got := validate(t, db)
	wantErr(t, got, "jetStreamDomain is required")

	db.Spec.LeafNodes.JetStreamDomain = "edge"
	_, got = validate(t, db)
	wantErr(t, got, "")

	// JetStream off: no domain needed
	db = newNATS()
	off := false
	db.Spec.JetStream.Enabled = &off
	db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{Remotes: []olddbapi.NATSLeafRemote{remote}}
	_, got = validate(t, db)
	wantErr(t, got, "")

	// hub side only: no domain needed
	db = newNATS()
	db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{Accept: true}
	_, got = validate(t, db)
	wantErr(t, got, "")
}

func TestNATSLeafRemoteRules(t *testing.T) {
	mk := func(r olddbapi.NATSLeafRemote) *olddbapi.NATS {
		db := newNATS()
		db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{JetStreamDomain: "edge", Remotes: []olddbapi.NATSLeafRemote{r}}
		return db
	}
	_, got := validate(t, mk(olddbapi.NATSLeafRemote{Account: "APP"}))
	wantErr(t, got, "one of urls and natsRef is required")
	_, got = validate(t, mk(olddbapi.NATSLeafRemote{Account: "APP", NATSRef: &kmapi.ObjectReference{Name: "hub"}, URLs: []string{"nats://hub:7422"}}))
	wantErr(t, got, "mutually exclusive")
	_, got = validate(t, mk(olddbapi.NATSLeafRemote{NATSRef: &kmapi.ObjectReference{Name: "hub"}, CredentialsSecret: &core.LocalObjectReference{}}))
	wantErr(t, got, "credentialsSecret.name can not be empty")
}

func TestNATSLeafUnscopedDefaultAccountWarns(t *testing.T) {
	mk := func(r olddbapi.NATSLeafRemote) *olddbapi.NATS {
		db := newNATS()
		db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{JetStreamDomain: "edge", Remotes: []olddbapi.NATSLeafRemote{r}}
		return db
	}
	warnings, got := validate(t, mk(olddbapi.NATSLeafRemote{NATSRef: &kmapi.ObjectReference{Name: "hub"}}))
	wantErr(t, got, "")
	if !hasWarning(warnings, "default account") {
		t.Fatalf("unscoped leaf must warn, got %v", warnings)
	}
	warnings, _ = validate(t, mk(olddbapi.NATSLeafRemote{NATSRef: &kmapi.ObjectReference{Name: "hub"}, Account: "APP"}))
	if hasWarning(warnings, "default account") {
		t.Fatalf("a leaf with an account must not warn, got %v", warnings)
	}
	warnings, _ = validate(t, mk(olddbapi.NATSLeafRemote{NATSRef: &kmapi.ObjectReference{Name: "hub"}, CredentialsSecret: &core.LocalObjectReference{Name: "creds"}}))
	if hasWarning(warnings, "default account") {
		t.Fatalf("a leaf with credentials must not warn, got %v", warnings)
	}
}

func TestNATSDistributedNeedsPlacementPolicy(t *testing.T) {
	// no reference
	db := newNATS()
	db.Spec.Distributed = true
	_, got := validate(t, db)
	wantErr(t, got, "podPlacementPolicy is required")

	// policy missing
	_, got = validate(t, distributedNATS("absent"))
	wantErr(t, got, "Not found")

	// policy without clusterSpreadConstraint
	pp := &psapi.PlacementPolicy{ObjectMeta: metav1.ObjectMeta{Name: "zonal"}}
	_, got = validate(t, distributedNATS("zonal"), pp)
	wantErr(t, got, "needs a clusterSpreadConstraint")

	// good three DCs
	pp = natsPolicy("ok", natsRule("dc-a", "", 0, 1, 2), natsRule("dc-b", psapi.DCRoleMember, 3, 4, 5), natsRule("dc-c", psapi.DCRoleArbiter))
	_, got = validate(t, distributedNATS("ok"), pp)
	wantErr(t, got, "")
}

func TestNATSDistributedForbidsFailoverPolicy(t *testing.T) {
	pp := natsPolicy("fo", natsRule("dc-a", "", 0, 1, 2), natsRule("dc-b", "", 3, 4, 5), natsRule("dc-c", psapi.DCRoleArbiter))
	pp.Spec.ClusterSpreadConstraint.FailoverPolicy = &psapi.FailoverPolicy{}
	_, got := validate(t, distributedNATS("fo"), pp)
	wantErr(t, got, "must not set clusterSpreadConstraint.failoverPolicy")
}

func TestNATSDistributedArbiterRuleHasEmptyReplicaIndices(t *testing.T) {
	pp := natsPolicy("arb", natsRule("dc-a", "", 0, 1, 2), natsRule("dc-b", "", 3, 4, 5), natsRule("dc-c", psapi.DCRoleArbiter, 6))
	_, got := validate(t, distributedNATS("arb"), pp)
	wantErr(t, got, "must have empty replicaIndices")

	// a Member needs indices and Witness is not used by NATS
	pp = natsPolicy("mem", natsRule("dc-a", "", 0, 1, 2), natsRule("dc-b", ""), natsRule("dc-c", psapi.DCRoleArbiter))
	_, got = validate(t, distributedNATS("mem"), pp)
	wantErr(t, got, "a Member rule needs replicaIndices")
	pp = natsPolicy("wit", natsRule("dc-a", "", 0, 1, 2), natsRule("dc-b", "", 3, 4, 5), natsRule("dc-c", psapi.DCRoleWitness, 6))
	_, got = validate(t, distributedNATS("wit"), pp)
	wantErr(t, got, "Unsupported value")
}

func TestNATSDistributedDCCountIsOddAndAtLeastThree(t *testing.T) {
	cases := []struct {
		name    string
		rules   []psapi.DistributionRule
		wantErr string
	}{
		{"two data DCs, no arbiter", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5)}, "two data DCs without an Arbiter rule"},
		{"two data DCs plus arbiter", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", psapi.DCRoleArbiter)}, ""},
		{"three data DCs", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", "", 6, 7, 8)}, ""},
		{"three data DCs plus arbiter is even", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", "", 6, 7, 8), natsRule("d", psapi.DCRoleArbiter)}, "odd and at least 3, got 4"},
		{"one data DC", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2)}, "at least two Member DCs"},
		{"duplicate DC", []psapi.DistributionRule{natsRule("a", "", 0, 1, 2), natsRule("a", "", 3, 4, 5), natsRule("c", psapi.DCRoleArbiter)}, "Duplicate value"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, got := validate(t, distributedNATS("p"), natsPolicy("p", c.rules...))
			wantErr(t, got, c.wantErr)
		})
	}

	// without JetStream there is no meta group, so two DCs are fine
	db := distributedNATS("p")
	off := false
	db.Spec.JetStream.Enabled = &off
	_, got := validate(t, db, natsPolicy("p", natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5)))
	wantErr(t, got, "")
}

func TestNATSTwoDCsWithJetStreamNeedLeafNodesNotGateway(t *testing.T) {
	// the recommended alternative: leafNodes with distinct JetStream domains, no gateway
	db := newNATS()
	db.Spec.LeafNodes = &olddbapi.NATSLeafNodesSpec{
		JetStreamDomain: "dc-b",
		Remotes:         []olddbapi.NATSLeafRemote{{NATSRef: &kmapi.ObjectReference{Name: "dc-a"}, Account: "APP"}},
	}
	warnings, got := validate(t, db)
	wantErr(t, got, "")
	if hasWarning(warnings, "gateway") {
		t.Fatalf("leafNodes between two DCs must not warn, got %v", warnings)
	}

	// a single gateway remote with JetStream warns
	db = newNATS()
	db.Spec.Gateway = &olddbapi.NATSGatewaySpec{Remotes: []olddbapi.NATSRemoteGateway{{Name: "dc-b", URLs: []string{"nats://b:7222"}}}}
	db.SetDefaults(nil)
	warnings, got = validate(t, db)
	wantErr(t, got, "")
	if !hasWarning(warnings, "spec.leafNodes") {
		t.Fatalf("a two-cluster gateway with JetStream must warn, got %v", warnings)
	}
}

func TestNATSDistributedGatewayNameMustBeEmpty(t *testing.T) {
	db := distributedNATS("p")
	db.Spec.Gateway = &olddbapi.NATSGatewaySpec{Name: "x"}
	pp := natsPolicy("p", natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", psapi.DCRoleArbiter))
	_, got := validate(t, db, pp)
	wantErr(t, got, "spec.gateway.name must be empty")
}

func TestNATSArbiterRules(t *testing.T) {
	// arbiter without distributed
	db := newNATS()
	db.Spec.Arbiter = &olddbapi.NATSArbiterSpec{}
	_, got := validate(t, db)
	wantErr(t, got, "spec.arbiter is only allowed with spec.distributed")

	// replicas below one
	pp := natsPolicy("p", natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", psapi.DCRoleArbiter))
	db = distributedNATS("p")
	zero := int32(0)
	db.Spec.Arbiter.Replicas = &zero
	_, got = validate(t, db, pp)
	wantErr(t, got, "arbiter replicas must be at least 1")

	// defaults: 1 replica, 1Gi
	db = distributedNATS("p")
	if db.ArbiterReplicas() != 1 || db.Spec.Arbiter.Storage == nil || db.Spec.Arbiter.Storage.Resources.Requests.Storage().String() != "1Gi" {
		t.Fatalf("unexpected arbiter defaults: %+v", db.Spec.Arbiter)
	}
	_, got = validate(t, db, pp)
	wantErr(t, got, "")
}

func TestNATSDistributedIsImmutable(t *testing.T) {
	pp := natsPolicy("p", natsRule("a", "", 0, 1, 2), natsRule("b", "", 3, 4, 5), natsRule("c", psapi.DCRoleArbiter))
	w := natsTestWebhook(t, pp)
	old := newNATS()
	_, err := w.ValidateUpdate(context.TODO(), old, distributedNATS("p"))
	if err == nil || !strings.Contains(err.Error(), "spec.distributed is immutable") {
		t.Fatalf("flipping distributed must be rejected, got %v", err)
	}
}

func TestNATSLeafNodesServiceAliasIsKnown(t *testing.T) {
	db := newNATS()
	db.Spec.ServiceTemplates = []olddbapi.NamedServiceTemplateSpec{{Alias: olddbapi.LeafNodesServiceAlias}}
	if string(olddbapi.LeafNodesServiceAlias) != "leafnodes" {
		t.Fatal("alias must be leafnodes")
	}
	_, got := validate(t, db)
	wantErr(t, got, "")
}
