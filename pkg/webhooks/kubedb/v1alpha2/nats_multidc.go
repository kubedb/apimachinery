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
	"fmt"
	"net/url"
	"strings"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kmapi "kmodules.xyz/client-go/api/v1"
	psapi "kubeops.dev/petset/apis/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// natsDCs is the DC layout read from a PlacementPolicy.
type natsDCs struct {
	members  []string
	arbiters []string
}

// total counts every DC that holds a JetStream meta peer, Arbiter DCs included.
func (d natsDCs) total() int { return len(d.members) + len(d.arbiters) }

// validateNATSMultiDC checks the gateway, leafnodes, distributed and arbiter rules of design 15.1 to 15.4.
func (w *NATSCustomWebhook) validateNATSMultiDC(ctx context.Context, db *olddbapi.NATS) (admission.Warnings, field.ErrorList) {
	var warnings admission.Warnings
	var allErr field.ErrorList
	specPath := field.NewPath("spec")

	if db.Spec.Gateway != nil {
		allErr = append(allErr, validateNATSGateway(db)...)
	}
	if db.Spec.LeafNodes != nil {
		errs, warns := validateNATSLeafNodes(db)
		allErr = append(allErr, errs...)
		warnings = append(warnings, warns...)
	}

	if !db.Spec.Distributed {
		if db.Spec.Arbiter != nil {
			allErr = append(allErr, field.Forbidden(specPath.Child("arbiter"), "spec.arbiter is only allowed with spec.distributed: true"))
		}
		// Two data DCs joined by a gateway merge their JetStream meta groups, and one lost DC then
		// leaves 3 of 6 peers. Only the seed list is visible here, so this is a warning.
		if db.Spec.Gateway != nil && db.IsJetStreamEnabled() && len(db.Spec.Gateway.Remotes) == 1 {
			warnings = append(warnings, "a gateway to exactly one remote cluster with JetStream joins two JetStream meta groups into one with no majority after losing a DC: add a third (Arbiter) DC, or connect the two DCs with spec.leafNodes and distinct jetStreamDomain values instead")
		}
		return warnings, allErr
	}

	// distributed
	if db.Spec.Gateway != nil && db.Spec.Gateway.Name != "" {
		allErr = append(allErr, field.Forbidden(specPath.Child("gateway", "name"),
			"spec.gateway.name must be empty when spec.distributed is true: each DC's gateway is named after the DC"))
	}
	dcs, errs, warns := w.validateNATSPlacementPolicy(ctx, db)
	allErr = append(allErr, errs...)
	warnings = append(warnings, warns...)
	if errs == nil && dcs != nil {
		allErr = append(allErr, validateNATSDCCount(db, dcs)...)
		if db.Spec.Arbiter != nil {
			if r := db.Spec.Arbiter.Replicas; r != nil && *r < 1 {
				allErr = append(allErr, field.Invalid(specPath.Child("arbiter", "replicas"), *r, "arbiter replicas must be at least 1"))
			}
			if len(dcs.arbiters) == 0 && db.Spec.Arbiter.Replicas != nil && *db.Spec.Arbiter.Replicas != 1 {
				warnings = append(warnings, "spec.arbiter is ignored: the PlacementPolicy has no Arbiter rule")
			}
		}
	}
	return warnings, allErr
}

func validateNATSGateway(db *olddbapi.NATS) field.ErrorList {
	var allErr field.ErrorList
	path := field.NewPath("spec", "gateway")
	gw := db.Spec.Gateway

	// the server refuses to start when the gateway name differs from cluster.name (ErrClusterNameConfigConflict)
	if !db.Spec.Distributed && gw.Name != "" && gw.Name != db.Name {
		allErr = append(allErr, field.Invalid(path.Child("name"), gw.Name,
			fmt.Sprintf("gateway name must equal the cluster name %q", db.Name)))
	}

	seen := map[string]struct{}{}
	for i, r := range gw.Remotes {
		rp := path.Child("remotes").Index(i)
		switch {
		case r.Name == "":
			allErr = append(allErr, field.Required(rp.Child("name"), "remote gateway name is required"))
		case !db.Spec.Distributed && r.Name == db.GatewayName():
			allErr = append(allErr, field.Invalid(rp.Child("name"), r.Name, "a gateway can not list itself as a remote"))
		default:
			if _, dup := seen[r.Name]; dup {
				allErr = append(allErr, field.Duplicate(rp.Child("name"), r.Name))
			}
			seen[r.Name] = struct{}{}
		}
		allErr = append(allErr, validateNATSURLsOrRef(rp, r.URLs, r.NATSRef, "nats", "tls")...)
	}
	return allErr
}

func validateNATSLeafNodes(db *olddbapi.NATS) (field.ErrorList, admission.Warnings) {
	var allErr field.ErrorList
	var warnings admission.Warnings
	path := field.NewPath("spec", "leafNodes")
	ln := db.Spec.LeafNodes

	// a leaf with its own JetStream needs a domain of its own (leaf-nodes.md, "JetStream over a leaf")
	if db.IsJetStreamEnabled() && len(ln.Remotes) > 0 && ln.JetStreamDomain == "" {
		allErr = append(allErr, field.Required(path.Child("jetStreamDomain"),
			"spec.leafNodes.jetStreamDomain is required when JetStream is enabled and spec.leafNodes.remotes is set"))
	}

	for i, r := range ln.Remotes {
		rp := path.Child("remotes").Index(i)
		allErr = append(allErr, validateNATSURLsOrRef(rp, r.URLs, r.NATSRef, "nats", "tls", "nats-leaf", "tls-leaf", "ws", "wss")...)
		if r.CredentialsSecret != nil && r.CredentialsSecret.Name == "" {
			allErr = append(allErr, field.Required(rp.Child("credentialsSecret", "name"), "credentialsSecret.name can not be empty"))
		}
		if r.Account == "" && r.CredentialsSecret == nil {
			warnings = append(warnings, fmt.Sprintf("spec.leafNodes.remotes[%d] binds the default account: it has no subject boundary, set account or credentialsSecret", i))
		}
	}
	return allErr, warnings
}

// validateNATSURLsOrRef requires exactly one of urls and natsRef; every url must use one of the schemes.
func validateNATSURLsOrRef(path *field.Path, urls []string, ref *kmapi.ObjectReference, schemes ...string) field.ErrorList {
	var allErr field.ErrorList
	switch {
	case len(urls) == 0 && ref == nil:
		allErr = append(allErr, field.Required(path, "one of urls and natsRef is required"))
	case len(urls) > 0 && ref != nil:
		allErr = append(allErr, field.Forbidden(path.Child("natsRef"), "urls and natsRef are mutually exclusive"))
	}
	if ref != nil && ref.Name == "" {
		allErr = append(allErr, field.Required(path.Child("natsRef", "name"), "natsRef.name can not be empty"))
	}
	for j, raw := range urls {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || !contains(schemes, u.Scheme) {
			allErr = append(allErr, field.Invalid(path.Child("urls").Index(j), raw,
				fmt.Sprintf("must be a URL with host and one of the schemes %s", strings.Join(schemes, ", "))))
		}
	}
	return allErr
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// validateNATSPlacementPolicy reads the PlacementPolicy of a distributed NATS and checks its rules.
func (w *NATSCustomWebhook) validateNATSPlacementPolicy(ctx context.Context, db *olddbapi.NATS) (*natsDCs, field.ErrorList, admission.Warnings) {
	var allErr field.ErrorList
	var warnings admission.Warnings
	refPath := field.NewPath("spec", "podTemplate", "spec", "podPlacementPolicy")

	ref := db.Spec.PodTemplate.Spec.PodPlacementPolicy
	if ref == nil || ref.Name == "" {
		return nil, field.ErrorList{field.Required(refPath, "spec.podTemplate.spec.podPlacementPolicy is required when spec.distributed is true")}, nil
	}
	if w.DefaultClient == nil {
		return nil, nil, nil
	}
	var pp psapi.PlacementPolicy
	if err := w.DefaultClient.Get(ctx, client.ObjectKey{Name: ref.Name}, &pp); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, field.ErrorList{field.NotFound(refPath, ref.Name)}, nil
		}
		return nil, field.ErrorList{field.InternalError(refPath, err)}, nil
	}
	csc := pp.Spec.ClusterSpreadConstraint
	if csc == nil {
		return nil, field.ErrorList{field.Invalid(refPath, ref.Name,
			"PlacementPolicy needs a clusterSpreadConstraint for a distributed NATS")}, nil
	}
	// failoverPolicy switches on the DC/DR machinery (primary-DC Lease, active/standby), which NATS does not have
	if csc.FailoverPolicy != nil {
		allErr = append(allErr, field.Forbidden(refPath,
			fmt.Sprintf("PlacementPolicy %q must not set clusterSpreadConstraint.failoverPolicy: every NATS DC is active", ref.Name)))
	}

	dcs := &natsDCs{}
	seen := map[string]struct{}{}
	for i, rule := range csc.DistributionRules {
		rp := field.NewPath("placementPolicy", "clusterSpreadConstraint", "distributionRules").Index(i)
		if _, dup := seen[rule.ClusterName]; dup {
			allErr = append(allErr, field.Duplicate(rp.Child("clusterName"), rule.ClusterName))
			continue
		}
		seen[rule.ClusterName] = struct{}{}
		switch rule.Role {
		case "", psapi.DCRoleMember:
			if len(rule.ReplicaIndices) == 0 {
				allErr = append(allErr, field.Required(rp.Child("replicaIndices"), "a Member rule needs replicaIndices"))
			}
			dcs.members = append(dcs.members, rule.ClusterName)
		case psapi.DCRoleArbiter:
			// the voter is not one of the policy's ordinals: the operator creates it as its own PetSet
			if len(rule.ReplicaIndices) != 0 {
				allErr = append(allErr, field.Invalid(rp.Child("replicaIndices"), rule.ReplicaIndices, "an Arbiter holds no data and must have empty replicaIndices"))
			}
			dcs.arbiters = append(dcs.arbiters, rule.ClusterName)
		default:
			allErr = append(allErr, field.NotSupported(rp.Child("role"), rule.Role, []string{string(psapi.DCRoleMember), string(psapi.DCRoleArbiter)}))
		}
	}
	if len(dcs.members) < 2 {
		allErr = append(allErr, field.Invalid(refPath, ref.Name, "a distributed NATS needs at least two Member DCs"))
	}
	return dcs, allErr, warnings
}

// validateNATSDCCount: one JetStream meta group spans every DC joined by gateways, so the DC count
// including the Arbiter DC must be odd and at least 3 (design 15.4 items 1 and 2).
func validateNATSDCCount(db *olddbapi.NATS, dcs *natsDCs) field.ErrorList {
	if !db.IsJetStreamEnabled() || len(dcs.members) < 2 {
		return nil
	}
	n := dcs.total()
	if n >= 3 && n%2 == 1 {
		return nil
	}
	msg := fmt.Sprintf("with JetStream the number of DCs including the Arbiter DC must be odd and at least 3, got %d", n)
	if len(dcs.members) == 2 && len(dcs.arbiters) == 0 {
		msg = "two data DCs without an Arbiter rule can not be joined by gateways with JetStream (the meta group loses its majority with either DC): add an Arbiter rule to the PlacementPolicy, or run two separate NATS objects connected by spec.leafNodes with distinct jetStreamDomain values"
	}
	return field.ErrorList{field.Invalid(field.NewPath("spec", "podTemplate", "spec", "podPlacementPolicy"), db.Spec.PodTemplate.Spec.PodPlacementPolicy.Name, msg)}
}
