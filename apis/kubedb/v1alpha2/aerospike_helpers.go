package v1alpha2

import (
	"kubedb.dev/apimachinery/apis/kubedb"

	meta_util "kmodules.xyz/client-go/meta"
)

func (a *Aerospike) ConfigSecretName() string {
	if a.Spec.Configuration != nil && a.Spec.Configuration.SecretName != "" {
		return a.Spec.Configuration.SecretName
	}
	return a.Name + "-config"
}

func (a *Aerospike) OffshootName() string {
	return a.Name
}

func (a *Aerospike) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.InstanceLabelKey:  a.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (a *Aerospike) GoverningServiceName() string {
	return meta_util.NameWithSuffix(a.ServiceName(), "pods")
}

func (a *Aerospike) ServiceName() string {
	return a.OffshootName()
}

func (a Aerospike) OffshootLabels() map[string]string {
	return a.offshootLabels(a.OffshootSelectors(), nil)
}

func (a Aerospike) offshootLabels(selector, overrides map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, a.Labels, overrides))
}

func (a Aerospike) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(a.Spec.ServiceTemplates, alias)
	return a.offshootLabels(meta_util.OverwriteKeys(a.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}
