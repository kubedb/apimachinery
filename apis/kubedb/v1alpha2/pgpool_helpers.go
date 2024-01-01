package v1alpha2

import (
	"fmt"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
)

func (p *Pgpool) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", p.ResourcePlural(), kubedb.GroupName)
}

func (p *Pgpool) ResourceShortCode() string {
	return ResourceCodePgpool
}

func (p *Pgpool) ResourceKind() string {
	return ResourceKindPgpool
}

func (p *Pgpool) ResourceSingular() string {
	return ResourceSingularPgpool
}

func (p *Pgpool) ResourcePlural() string {
	return ResourcePluralPgpool
}

func (p *Pgpool) ConfigSecretName() string {
	return meta_util.NameWithSuffix(p.OffshootName(), "config")
}

func (p *Pgpool) ServiceAccountName() string {
	return p.OffshootName()
}

func (p *Pgpool) GoverningServiceName() string {
	return meta_util.NameWithSuffix(p.ServiceName(), "pods")
}

func (p *Pgpool) ServiceName() string {
	return p.OffshootName()
}

// Owner returns owner reference to resources
func (p *Pgpool) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(p, SchemeGroupVersion.WithKind(p.ResourceKind()))
}

func (p *Pgpool) PodLabels(extraLabels ...map[string]string) map[string]string {
	var labels map[string]string
	if p.Spec.PodTemplate != nil {
		return p.offshootLabels(meta_util.OverwriteKeys(p.OffshootSelectors(), extraLabels...), p.Spec.PodTemplate.Labels)
	} else {
		return p.offshootLabels(meta_util.OverwriteKeys(p.OffshootSelectors(), extraLabels...), labels)
	}
}

func (p *Pgpool) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	var labels map[string]string
	if p.Spec.PodTemplate != nil {
		return p.offshootLabels(meta_util.OverwriteKeys(p.OffshootSelectors(), extraLabels...), p.Spec.PodTemplate.Controller.Labels)
	} else {
		return p.offshootLabels(meta_util.OverwriteKeys(p.OffshootSelectors(), extraLabels...), labels)
	}
}

func (p *Pgpool) OffshootLabels() map[string]string {
	return p.offshootLabels(p.OffshootSelectors(), nil)
}

func (p *Pgpool) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentConnectionPooler
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, p.Labels, override))
}

func (p *Pgpool) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      p.ResourceFQN(),
		meta_util.InstanceLabelKey:  p.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (p *Pgpool) StatefulSetName() string {
	return p.OffshootName()
}

func (p *Pgpool) OffshootName() string {
	return p.Name
}

func (p *Pgpool) GetAuthSecretName() string {
	if p.Spec.AuthSecret != nil && p.Spec.AuthSecret.Name != "" {
		return p.Spec.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(p.OffshootName(), "auth")
}

func (p *Pgpool) SetHealthCheckerDefaults() {
	if p.Spec.HealthChecker.PeriodSeconds == nil {
		p.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if p.Spec.HealthChecker.TimeoutSeconds == nil {
		p.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if p.Spec.HealthChecker.FailureThreshold == nil {
		p.Spec.HealthChecker.FailureThreshold = pointer.Int32P(1)
	}
}

// PrimaryServiceDNS make primary host dns with require template
func (p *Pgpool) PrimaryServiceDNS() string {
	return fmt.Sprintf("%v.%v.svc", p.ServiceName(), p.Namespace)
}

func (p *Pgpool) GetNameSpacedName() string {
	return p.Namespace + "/" + p.Name
}

func (p *Pgpool) SetSecurityContext(ppVersion *catalog.PgpoolVersion) {
	if p.Spec.PodTemplate.Spec.SecurityContext == nil {
		p.Spec.PodTemplate.Spec.SecurityContext = &core.PodSecurityContext{
			RunAsUser:    ppVersion.Spec.SecurityContext.RunAsUser,
			RunAsGroup:   ppVersion.Spec.SecurityContext.RunAsUser,
			RunAsNonRoot: pointer.BoolP(true),
		}
	} else {
		if p.Spec.PodTemplate.Spec.SecurityContext.RunAsUser == nil {
			p.Spec.PodTemplate.Spec.SecurityContext.RunAsUser = ppVersion.Spec.SecurityContext.RunAsUser
		}
		if p.Spec.PodTemplate.Spec.SecurityContext.RunAsGroup == nil {
			p.Spec.PodTemplate.Spec.SecurityContext.RunAsGroup = p.Spec.PodTemplate.Spec.SecurityContext.RunAsUser
		}
	}

	// Need to set FSGroup equal to  p.Spec.PodTemplate.Spec.SecurityContext.RunAsGroup.
	// So that /var/pv directory have the group permission for the RunAsGroup user GID.
	// Otherwise, We will get write permission denied.
	p.Spec.PodTemplate.Spec.SecurityContext.FSGroup = p.Spec.PodTemplate.Spec.SecurityContext.RunAsGroup
}

func (p *Pgpool) SetDefaults(ppVersion *catalog.PgpoolVersion) {
	if p == nil {
		return
	}
	if p.Spec.Replicas == nil {
		p.Spec.Replicas = pointer.Int32P(1)
	}
	if p.Spec.TerminationPolicy == "" {
		p.Spec.TerminationPolicy = TerminationPolicyDelete
	}
	if p.Spec.PodTemplate != nil {
		p.SetSecurityContext(ppVersion)
	}
}
