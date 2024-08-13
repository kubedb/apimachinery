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
	"fmt"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	"kubedb.dev/apimachinery/apis/ops"
	"kubedb.dev/apimachinery/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/client-go/apiextensions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (p PgBouncerOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralPgBouncerOpsRequest))
}

var _ apis.ResourceInfo = &PgBouncerOpsRequest{}

func (p PgBouncerOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralPgBouncerOpsRequest, ops.GroupName)
}

func (p PgBouncerOpsRequest) ResourceShortCode() string {
	return ResourceCodePgBouncerOpsRequest
}

func (p PgBouncerOpsRequest) ResourceKind() string {
	return ResourceKindPgBouncerOpsRequest
}

func (p PgBouncerOpsRequest) ResourceSingular() string {
	return ResourceSingularPgBouncerOpsRequest
}

func (p PgBouncerOpsRequest) ResourcePlural() string {
	return ResourcePluralPgBouncerOpsRequest
}

func (p PgBouncerOpsRequest) ValidateSpecs() error {
	return nil
}

var _ Accessor = &PgBouncerOpsRequest{}

func (p *PgBouncerOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return p.ObjectMeta
}

func (p *PgBouncerOpsRequest) GetDBRefName() string {
	return p.Spec.ServerRef.Name
}

func (p *PgBouncerOpsRequest) GetRequestType() any {
	return p.Spec.Type
}

func (p *PgBouncerOpsRequest) GetStatus() OpsRequestStatus {
	return p.Status
}

func (p *PgBouncerOpsRequest) SetStatus(s OpsRequestStatus) {
	p.Status = s
}

func (p *PgBouncerOpsRequest) GetCurrentVersionName(client client.Client) (string, error) {
	version, err := p.GetCurrentVersion(client)
	if err != nil {
		return "", err
	}
	return version.Name, nil
}

func (p *PgBouncerOpsRequest) GetCurrentVersion(client client.Client) (*v1alpha1.PgBouncerVersion, error) {
	if bouncer, err := p.GetPgBouncer(client); err != nil {
		return nil, err
	} else {
		version := v1alpha1.PgBouncerVersion{}
		err = client.Get(context.TODO(), types.NamespacedName{Namespace: bouncer.Namespace, Name: bouncer.Spec.Version}, &version)
		if err != nil {
			return nil, err
		}
		return &version, nil
	}
}

func (p *PgBouncerOpsRequest) GetTargetVersion(client client.Client) (*v1alpha1.PgBouncerVersion, error) {
	if p.Spec.Type != PgBouncerOpsRequestTypeUpdateVersion {
		return nil, fmt.Errorf("pgbouncer version will not be updated with this ops-request")
	}
	if p.Spec.UpdateVersion.TargetVersion == "" {
		return nil, fmt.Errorf("targeted pgbouncer version name is invalid")
	}
	version := v1alpha1.PgBouncerVersion{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Spec.UpdateVersion.TargetVersion}, &version)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (p *PgBouncerOpsRequest) GetPgBouncer(client client.Client) (*dbapi.PgBouncer, error) {
	bouncer := &dbapi.PgBouncer{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: p.Spec.ServerRef.Name, Namespace: p.Namespace}, bouncer); err != nil {
		return nil, err
	}
	return bouncer, nil
}
