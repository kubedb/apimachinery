package ops

import (
	"fmt"

	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ValidateStorageExpansion(curStorage *core.PersistentVolumeClaimSpec, newStorage *resource.Quantity, phase opsapi.OpsRequestPhase, nodeType string) error {
	if curStorage == nil {
		return fmt.Errorf("storage not configured for %s", nodeType)
	}
	cur, ok := curStorage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return fmt.Errorf("failed to parse current %s storage size", nodeType)
	}
	if (phase == opsapi.OpsRequestPhasePending || phase == "") && cur.Cmp(*newStorage) >= 0 {
		return fmt.Errorf("desired %s storage size must be greater than current storage. Current storage: %v", nodeType, cur.String())
	}
	return nil
}
