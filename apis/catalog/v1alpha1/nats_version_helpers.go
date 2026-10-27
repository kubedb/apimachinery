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
	"fmt"
	"strings"
	"time"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/catalog"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

// NATSMinLameDuckDuration is the server minimum of lame_duck_duration.
const NATSMinLameDuckDuration = 30 * time.Second

func (NATSVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralNATSVersion))
}

var _ apis.ResourceInfo = &NATSVersion{}

func (n NATSVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralNATSVersion, catalog.GroupName)
}

func (n NATSVersion) ResourceShortCode() string {
	return ResourceCodeNATSVersion
}

func (n NATSVersion) ResourceKind() string {
	return ResourceKindNATSVersion
}

func (n NATSVersion) ResourceSingular() string {
	return ResourceSingularNATSVersion
}

func (n NATSVersion) ResourcePlural() string {
	return ResourcePluralNATSVersion
}

// ValidateSpecs checks that the server, reloader and exporter images are all set, and that the
// default lame-duck timing, if given, is one the server accepts.
func (n NATSVersion) ValidateSpecs() error {
	if n.Spec.Version == "" ||
		n.Spec.DB.Image == "" ||
		n.Spec.Reloader.Image == "" ||
		n.Spec.Exporter.Image == "" {
		fields := []string{
			"spec.version",
			"spec.db.image",
			"spec.reloader.image",
			"spec.exporter.image",
		}
		return fmt.Errorf("atleast one of the following specs is not set for NATSVersion %q: %s", n.Name, strings.Join(fields, ", "))
	}
	ld := n.Spec.LameDuck
	if ld.Duration.Duration != 0 && ld.Duration.Duration < NATSMinLameDuckDuration {
		return fmt.Errorf("spec.lameDuck.duration of NATSVersion %q must be at least %s", n.Name, NATSMinLameDuckDuration)
	}
	if ld.Duration.Duration != 0 && ld.GracePeriod.Duration >= ld.Duration.Duration {
		return fmt.Errorf("spec.lameDuck.gracePeriod of NATSVersion %q must be less than spec.lameDuck.duration", n.Name)
	}
	return nil
}
