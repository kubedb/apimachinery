/*
Copyright AppsCode Inc. and Contributors.

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

package v1

import (
	"fmt"
	"sort"
)

// Validate checks the DC/DR failover configuration of a ClusterSpreadConstraint.
// It is safe to call when FailoverPolicy is nil (returns nil). Wire it from the
// PlacementPolicy validating webhook.
func (c *ClusterSpreadConstraint) Validate() error {
	if c == nil || c.FailoverPolicy == nil {
		return nil
	}
	if err := c.validateTopology(); err != nil {
		return err
	}
	members, arbiters, witnesses := 0, 0, 0
	for i := range c.DistributionRules {
		r := &c.DistributionRules[i]
		switch r.Role {
		case "", DCRoleMember:
			members++
			if !r.IsDataBearing() {
				return fmt.Errorf("distributionRule for cluster %q has role Member but no replicaIndices or topology", r.ClusterName)
			}
		case DCRoleArbiter:
			arbiters++
			if r.IsDataBearing() {
				return fmt.Errorf("distributionRule for cluster %q has role Arbiter but carries replicaIndices or topology (an arbiter holds no data)", r.ClusterName)
			}
		case DCRoleWitness:
			witnesses++
			if !r.IsDataBearing() {
				return fmt.Errorf("distributionRule for cluster %q has role Witness but no replicaIndices or topology (a witness is data bearing)", r.ClusterName)
			}
		default:
			return fmt.Errorf("distributionRule for cluster %q has unknown role %q", r.ClusterName, r.Role)
		}
	}
	if members < 2 {
		return fmt.Errorf("DC/DR requires at least two Member data centers, found %d", members)
	}

	fp := c.FailoverPolicy
	if fp.FailoverGroupRef != nil && fp.FailoverGroupRef.Name == "" {
		return fmt.Errorf("failoverGroupRef.name must not be empty")
	}

	switch fp.Mode {
	case "":
		// derived, nothing to check
	case FailoverModeTwoDC:
		if members != 2 || (arbiters+witnesses) < 1 {
			return fmt.Errorf("mode TwoDC requires exactly two Members and at least one Arbiter or Witness, found members=%d arbiters=%d witnesses=%d", members, arbiters, witnesses)
		}
	case FailoverModeThreeDC:
		if members < 3 {
			return fmt.Errorf("mode ThreeDC requires at least three Members, found %d", members)
		}
	default:
		return fmt.Errorf("mode must be TwoDC or ThreeDC, got %q", fp.Mode)
	}
	return nil
}

// validateTopology checks the per role placement: a rule sets either
// ReplicaIndices or Topology, all data bearing rules use the same one, and an
// ordinal of a role is placed on one cluster only.
func (c *ClusterSpreadConstraint) validateTopology() error {
	topology := c.IsTopology()
	owner := map[string]map[int32]string{}
	for i := range c.DistributionRules {
		r := &c.DistributionRules[i]
		if len(r.ReplicaIndices) > 0 && len(r.Topology) > 0 {
			return fmt.Errorf("distributionRule for cluster %q sets both replicaIndices and topology; use topology for a database with node roles, replicaIndices otherwise", r.ClusterName)
		}
		if topology && len(r.ReplicaIndices) > 0 {
			return fmt.Errorf("distributionRule for cluster %q sets replicaIndices while other rules set topology; leave replicaIndices empty in topology mode", r.ClusterName)
		}
		for _, role := range sortedRoles(r.Topology) {
			if role == "" {
				return fmt.Errorf("distributionRule for cluster %q has a topology entry with an empty role name", r.ClusterName)
			}
			if owner[role] == nil {
				owner[role] = map[int32]string{}
			}
			for _, idx := range r.Topology[role] {
				if idx < 0 {
					return fmt.Errorf("distributionRule for cluster %q has a negative index %d for role %q", r.ClusterName, idx, role)
				}
				if prev, ok := owner[role][idx]; ok {
					return fmt.Errorf("index %d of role %q is placed on both cluster %q and cluster %q", idx, role, prev, r.ClusterName)
				}
				owner[role][idx] = r.ClusterName
			}
		}
	}
	return nil
}

// IsTopology reports whether the placement places node roles, that is whether
// any rule sets Topology.
func (c *ClusterSpreadConstraint) IsTopology() bool {
	if c == nil {
		return false
	}
	for i := range c.DistributionRules {
		if len(c.DistributionRules[i].Topology) > 0 {
			return true
		}
	}
	return false
}

// RoleTotal returns the number of nodes of a role across all rules.
func (c *ClusterSpreadConstraint) RoleTotal(role string) int32 {
	if c == nil {
		return 0
	}
	var n int32
	for i := range c.DistributionRules {
		n += int32(len(c.DistributionRules[i].Topology[role]))
	}
	return n
}

// IsDataBearing reports whether the rule places any pod on its cluster, through
// ReplicaIndices or Topology. A rule with neither holds no data (an Arbiter).
func (r DistributionRule) IsDataBearing() bool {
	if len(r.ReplicaIndices) > 0 {
		return true
	}
	for _, indices := range r.Topology {
		if len(indices) > 0 {
			return true
		}
	}
	return false
}

// RoleReplicas returns how many nodes of a role the rule places on its cluster,
// and whether the rule names the role at all.
func (r DistributionRule) RoleReplicas(role string) (int32, bool) {
	indices, ok := r.Topology[role]
	return int32(len(indices)), ok
}

func sortedRoles(t map[string][]int32) []string {
	roles := make([]string, 0, len(t))
	for role := range t {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}
