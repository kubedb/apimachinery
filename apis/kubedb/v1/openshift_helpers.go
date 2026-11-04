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

package v1

import (
	"fmt"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
)

// OpenShift assigns every namespace a block of UIDs and supplemental groups and records them
// in these annotations. Its restricted-v2 SCC only admits pods whose runAsUser lies in the
// uid-range block and whose fsGroup lies in the supplemental-groups block, so a catalog UID
// such as postgres's 70 is rejected outright.
const (
	OpenShiftSCCUIDRangeAnnotation           = "openshift.io/sa.scc.uid-range"
	OpenShiftSCCSupplementalGroupsAnnotation = "openshift.io/sa.scc.supplemental-groups"
)

// NamespaceSecurityIDs are the IDs a database pod should run with in a namespace that carries
// OpenShift's SCC annotations. They are defaults only: a value the user set on the pod or
// container securityContext always wins.
// +k8s:deepcopy-gen=false
// +k8s:openapi-gen=false
type NamespaceSecurityIDs struct {
	RunAsUser  int64
	RunAsGroup int64
	FSGroup    int64
}

// GetNamespaceSecurityIDs reads the OpenShift SCC annotations from a namespace.
//
// It returns nil, nil when the namespace has no uid-range annotation, which is the case on
// every non-OpenShift cluster; callers then keep the catalog defaults.
//
//   - RunAsUser and RunAsGroup are the first ID of the uid-range block.
//   - FSGroup is the first ID of the supplemental-groups block, or of the uid-range block when
//     the namespace has no supplemental-groups annotation.
func GetNamespaceSecurityIDs(ns *core.Namespace) (*NamespaceSecurityIDs, error) {
	if ns == nil {
		return nil, nil
	}
	uidRange, ok := ns.Annotations[OpenShiftSCCUIDRangeAnnotation]
	if !ok {
		return nil, nil
	}
	uid, err := firstIDOfSCCRange(uidRange)
	if err != nil {
		return nil, fmt.Errorf("namespace %s: invalid %s annotation %q: %w", ns.Name, OpenShiftSCCUIDRangeAnnotation, uidRange, err)
	}
	ids := &NamespaceSecurityIDs{RunAsUser: uid, RunAsGroup: uid, FSGroup: uid}

	if groups, ok := ns.Annotations[OpenShiftSCCSupplementalGroupsAnnotation]; ok {
		gid, err := firstIDOfSCCRange(groups)
		if err != nil {
			return nil, fmt.Errorf("namespace %s: invalid %s annotation %q: %w", ns.Name, OpenShiftSCCSupplementalGroupsAnnotation, groups, err)
		}
		ids.FSGroup = gid
	}
	return ids, nil
}

// firstIDOfSCCRange returns the first ID of the first block in an OpenShift SCC range
// annotation. A block is written either as "start/size" or as "start-end", and the
// supplemental-groups annotation may hold several comma separated blocks.
func firstIDOfSCCRange(value string) (int64, error) {
	block := strings.TrimSpace(strings.Split(value, ",")[0])

	var startStr, restStr string
	var isSize bool
	if s, r, found := strings.Cut(block, "/"); found {
		startStr, restStr, isSize = s, r, true
	} else if s, r, found := strings.Cut(block, "-"); found {
		startStr, restStr = s, r
	} else {
		return 0, fmt.Errorf("expected <start>/<size> or <start>-<end>")
	}

	start, err := strconv.ParseInt(strings.TrimSpace(startStr), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid start: %w", err)
	}
	rest, err := strconv.ParseInt(strings.TrimSpace(restStr), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid range: %w", err)
	}
	if start <= 0 {
		return 0, fmt.Errorf("start must be a non-root ID, got %d", start)
	}
	if isSize && rest <= 0 {
		return 0, fmt.Errorf("size must be positive, got %d", rest)
	}
	if !isSize && rest < start {
		return 0, fmt.Errorf("end %d is before start %d", rest, start)
	}
	return start, nil
}
