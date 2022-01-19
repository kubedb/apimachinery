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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/crds"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	clientutil "kmodules.xyz/client-go/client"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (_ ElasticsearchDashboard) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralElasticsearchDashboard))
}

var _ apis.ResourceInfo = &ElasticsearchDashboard{}

func (ed ElasticsearchDashboard) OffshootName() string {
	return ed.Name
}
func (ed ElasticsearchDashboard) ServiceName() string {
	return ed.OffshootName()
}
func (ed ElasticsearchDashboard) DeploymentName() string {
	return ed.OffshootName()
}
func (ed ElasticsearchDashboard) ContainerName() string {
	return meta_util.NameWithSuffix(ed.Name, "dashboard")
}

// returns owner reference to resources

func (ed ElasticsearchDashboard) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(&ed, SchemeGroupVersion.WithKind(ResourceKindElasticsearchDashboard))
}

func (ed ElasticsearchDashboard) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralElasticsearchDashboard, kubedb.GroupName)
}

func (ed ElasticsearchDashboard) ResourceShortCode() string {
	return ResourceCodeElasticsearchDashboard
}

func (ed ElasticsearchDashboard) ResourceKind() string {
	return ResourceKindElasticsearchDashboard
}

func (ed ElasticsearchDashboard) ResourceSingular() string {
	return ResourceSingularElasticsearchDashboard
}

func (ed ElasticsearchDashboard) ResourcePlural() string {
	return ResourcePluralElasticsearchDashboard
}

func (ed ElasticsearchDashboard) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      ed.ResourceFQN(),
		meta_util.InstanceLabelKey:  ed.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

// returns a labelselector by offshooting extra selector (if any)

func (ed ElasticsearchDashboard) Selectors() *meta.LabelSelector {
	extraLabels := map[string]string{
		meta_util.InstanceLabelKey: ed.Name,
	}
	return &meta.LabelSelector{
		MatchLabels: ed.OffshootSelectors(extraLabels),
	}
}

func (ed ElasticsearchDashboard) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentDashboard
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, ed.Labels(), override))
}

func (ed ElasticsearchDashboard) OffshootLabels() map[string]string {
	return ed.offshootLabels(ed.OffshootSelectors(), nil)
}

func (ed ElasticsearchDashboard) Labels(extraLabels ...map[string]string) map[string]string {
	return meta_util.OverwriteKeys(ed.OffshootSelectors(), extraLabels...)
}

func (ed ElasticsearchDashboard) PodLabels(extraLabels ...map[string]string) map[string]string {
	return meta_util.OverwriteKeys(ed.OffshootSelectors(), extraLabels...)
}
func (ed ElasticsearchDashboard) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return ed.offshootLabels(meta_util.OverwriteKeys(ed.OffshootSelectors(), extraLabels...), ed.Spec.PodTemplate.Controller.Labels)
}

func (ed ElasticsearchDashboard) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	return meta_util.OverwriteKeys(ed.OffshootSelectors(), extraLabels...)
}
func (ed ElasticsearchDashboard) GetServiceSelectors() map[string]string {
	extraSelectors := map[string]string{
		"app.kubernetes.io/instance": ed.Name,
	}
	return ed.OffshootSelectors(extraSelectors)
}

// CertificateName returns the default certificate name and/or certificate secret name for a certificate alias
func (ed *ElasticsearchDashboard) CertificateName(alias ElasticsearchDashboardCertificateAlias) string {
	return meta_util.NameWithSuffix(ed.Name, fmt.Sprintf("%s-cert", string(alias)))
}

// returns the mountPath for certificate secrets.
// if configDir is "/usr/share/kibana/config",
// mountPath will be, "/usr/share/kibana/config/certs/<alias>/filename".

func (ed ElasticsearchDashboard) CertSecretVolumeMountPath(configDir string, alias ElasticsearchDashboardCertificateAlias) string {
	return filepath.Join(configDir, "certs", string(alias))
}

// returns a certificate file path  for a specific file using the certificate alias

func (ed ElasticsearchDashboard) CertificateFilePath(configDir string, alias ElasticsearchDashboardCertificateAlias, filename string) string {
	return filepath.Join(ed.CertSecretVolumeMountPath(configDir, alias), filename)
}

func (ed ElasticsearchDashboard) DashboardConfigSecretName() string {
	return meta_util.NameWithSuffix(ed.Name, "dashboard-config")
}

func (ed ElasticsearchDashboard) GetServicePort(alias ServiceAlias) int32 {
	reqAlias := v1alpha2.ServiceAlias(alias)
	svcTemplate := v1alpha2.GetServiceTemplate(ed.Spec.ServiceTemplates, reqAlias)
	return svcTemplate.Spec.Ports[0].Port
}

func (ed ElasticsearchDashboard) DatabaseConnectionURL(servicePort int32) (string, error) {
	if ed.Spec.DatabaseRef != nil {
		if &ed.Spec.DatabaseRef.Name == nil || &ed.Spec.DatabaseRef.Namespace == nil {
			return "", errors.New("required database fields not found")
		}
		return fmt.Sprintf("%s://%s.%s.svc:%d", ed.GetConnectionScheme(), ed.Spec.DatabaseRef.Name, ed.Spec.DatabaseRef.Namespace, servicePort), nil
	}
	return fmt.Sprintf("%s://%s.%s.svc:%d", ed.GetConnectionScheme(), ed.Spec.DatabaseRef.Name, ed.Spec.DatabaseRef.Namespace, servicePort), nil
}

func (ed *ElasticsearchDashboard) GetConnectionScheme() string {
	scheme := "http"
	if ed.Spec.EnableSSL {
		scheme = "https"
	}
	return scheme
}

// returns the volume name for certificate secret.
// Values will be like: es-toplogy-<alias>-certs
func (ed *ElasticsearchDashboard) CertSecretVolumeName(alias ElasticsearchDashboardCertificateAlias) string {
	return strings.Join([]string{ed.Name, string(alias), "certs"}, "-")
}

// GetCertSecretName returns the secret name for a certificate alias if any,
// otherwise returns default certificate secret name for the given alias.

func (ed *ElasticsearchDashboard) GetCertSecretName(alias ElasticsearchDashboardCertificateAlias) string {
	if ed.Spec.TLS != nil {
		name, ok := kmapi.GetCertificateSecretName(ed.Spec.TLS.Certificates, string(alias))
		if ok {
			return name
		}
	}
	return ed.CertificateName(alias)
}

func (ed *ElasticsearchDashboard) CertSecretExists(alias ElasticsearchDashboardCertificateAlias) bool {
	if ed.Spec.TLS != nil {
		_, ok := kmapi.GetCertificateSecretName(ed.Spec.TLS.Certificates, string(alias))
		if ok {
			return true
		}
	}
	return false
}

// returns the volume name for config Secret.

func (ed *ElasticsearchDashboard) GetVolumeName(alias ElasticsearchDashboardCertificateAlias) string {
	return meta_util.NameWithSuffix(string(alias), "volume")
}

func (ed *ElasticsearchDashboard) ConfigSecretVolumeName() string {
	return meta_util.NameWithSuffix(ed.Name, "dashboard-config")
}

func (ed *ElasticsearchDashboard) AuthSecretName() string {
	if ed.Spec.AuthSecret != nil {
		return ed.Spec.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(ed.Name, "database-cred")
}

func (ed *ElasticsearchDashboard) GetSecretName(alias ElasticsearchDashboardCertificateAlias) string {
	return meta_util.NameWithSuffix(ed.Name, string(alias))
}

func (ed *ElasticsearchDashboard) DatabaseClientSecretName() string {
	return meta_util.NameWithSuffix(ed.Name, "database-client")
}

func (ed *ElasticsearchDashboard) ClientCertificateCN(alias ElasticsearchDashboardCertificateAlias) string {
	return fmt.Sprintf("%s-%s", ed.Name, string(alias))
}

func (ed *ElasticsearchDashboard) GetDatabaseClientCertName(databaseName string) string {
	return fmt.Sprintf("%s-%s", databaseName, DefaultDatabaseClientCertSuffix)
}

// ......................................................status condition and phase helpers...........................................................

// check if deployment status is available
func (ed *ElasticsearchDashboard) IsDeploymentAvailable(depl *apps.Deployment) bool {
	cond := true
	cond = cond && ed.IsConditionTrue(depl.Status.Conditions, string(Available))
	return cond
}

// Check if status condition type is true
func (ed *ElasticsearchDashboard) IsConditionTrue(conditions []apps.DeploymentCondition, condType string) bool {
	for i := range conditions {
		if string(conditions[i].Type) == condType && conditions[i].Status == core.ConditionTrue {
			return true
		}
	}
	return false
}

func isOfficialTypes(group string) bool {
	return !strings.ContainsRune(group, '.')
}

func (ed *ElasticsearchDashboard) PatchStatus(ctx context.Context, r client.Client, obj client.Object, transform clientutil.TransformFunc, opts ...client.PatchOption) (client.Object, kutil.VerbType, error) {

	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := r.Get(ctx, key, obj)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	var patch client.Patch
	if isOfficialTypes(obj.GetObjectKind().GroupVersionKind().Group) {
		patch = client.StrategicMergeFrom(obj)
	} else {
		patch = client.MergeFrom(obj)
	}

	obj = transform(obj.DeepCopyObject().(client.Object), false)
	err = r.Status().Patch(ctx, obj, patch, opts...)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	return obj, kutil.VerbPatched, nil
}

func (ed *ElasticsearchDashboard) GetPhaseFromCondition(conditions []kmapi.Condition) DashboardPhase {

	if kmapi.IsConditionTrue(conditions, string(DashboardConditionStateGreenOrYellow)) &&
		kmapi.IsConditionTrue(conditions, string(DashboardConditionAcceptingConnection)) &&
		kmapi.IsConditionTrue(conditions, string(DashboardConditionDeploymentAvailable)) &&
		kmapi.IsConditionTrue(conditions, string(DashboardConditionServiceReady)) {
		return DashboardPhaseReady
	}

	if kmapi.IsConditionTrue(conditions, string(DashboardConditionStateGreenOrYellow)) ||
		kmapi.IsConditionTrue(conditions, string(DashboardConditionAcceptingConnection)) ||
		kmapi.IsConditionTrue(conditions, string(DashboardConditionDeploymentAvailable)) ||
		kmapi.IsConditionTrue(conditions, string(DashboardConditionServiceReady)) {
		return DashboardPhaseProvisioning
	}

	return DashboardPhaseNotReady
}

// ..................................................................................................................................................
