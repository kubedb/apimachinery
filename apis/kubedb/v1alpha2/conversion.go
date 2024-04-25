// nolint:goconst
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/conversion"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha3"
)

func Convert_v1alpha3_PgBouncerSpec_To_v1alpha2_PgBouncerSpec(in *v1alpha3.PgBouncerSpec, out *PgBouncerSpec, s conversion.Scope) error {
	return nil
}
