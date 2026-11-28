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

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/catalog"
	"kubedb.dev/apimachinery/crds"

	"github.com/Masterminds/semver/v3"
	"kmodules.xyz/client-go/apiextensions"
)

func (YugabyteDBVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralYugabyteDBVersion))
}

var _ apis.ResourceInfo = &YugabyteDBVersion{}

func (r YugabyteDBVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralYugabyteDBVersion, catalog.GroupName)
}

func (r YugabyteDBVersion) ResourceShortCode() string {
	return ResourceCodeYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourceKind() string {
	return ResourceKindYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourceSingular() string {
	return ResourceSingularYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourcePlural() string {
	return ResourcePluralYugabyteDBVersion
}

func (r YugabyteDBVersion) ValidateSpecs() error {
	if r.Spec.Version == "" ||
		r.Spec.DB.Image == "" ||
		r.Spec.YSQLMajorVersion == 0 {
		fields := []string{
			"spec.version",
			"spec.db.image",
			"spec.ysqlMajorVersion",
		}
		return fmt.Errorf("atleast one of the following specs is not set for YugabyteDBVersion %q: %s", r.Name, strings.Join(fields, ", "))
	}
	return nil
}

// CheckUpdateTo reports whether a database that runs this version may be updated to target. It is the one
// check that both the webhook and the ops-manager use:
//   - the target must be newer: a downgrade is refused whatever the catalog says (an unfinalized upgrade is
//     undone with a RollbackUpgrade request instead);
//   - the YSQL major version must stay the same;
//   - as in every KubeDB catalog, the updateConstraints of the current version decide: a denylist entry that
//     matches refuses the update, and a non-empty allowlist must match. An entry is a version name, a
//     version, or a semver constraint such as "> 2025.1.4".
func (v *YugabyteDBVersion) CheckUpdateTo(target *YugabyteDBVersion) error {
	if v.Name == target.Name {
		return fmt.Errorf("version %s is the current version", target.Name)
	}
	cv, err := semver.NewVersion(v.Spec.Version)
	if err != nil {
		return fmt.Errorf("version %s: %q is not a semantic version: %w", v.Name, v.Spec.Version, err)
	}
	tv, err := semver.NewVersion(target.Spec.Version)
	if err != nil {
		return fmt.Errorf("version %s: %q is not a semantic version: %w", target.Name, target.Spec.Version, err)
	}
	if !tv.GreaterThan(cv) {
		return fmt.Errorf("version %s (%s) is not newer than the current version %s (%s): downgrades are not supported; an upgrade that is not finalized is undone with a RollbackUpgrade request",
			target.Name, target.Spec.Version, v.Name, v.Spec.Version)
	}
	if v.Spec.YSQLMajorVersion != target.Spec.YSQLMajorVersion {
		return fmt.Errorf("version %s runs YSQL on PostgreSQL %d but %s on %d: an update between YSQL major versions is not supported",
			v.Name, v.Spec.YSQLMajorVersion, target.Name, target.Spec.YSQLMajorVersion)
	}
	matches := func(list string, entries []string) (bool, error) {
		for _, e := range entries {
			if e == target.Name || e == target.Spec.Version {
				return true, nil
			}
			c, err := semver.NewConstraint(e)
			if err != nil {
				return false, fmt.Errorf("invalid %s entry %q in the updateConstraints of %s: %w", list, e, v.Name, err)
			}
			if c.Check(tv) {
				return true, nil
			}
		}
		return false, nil
	}
	uc := v.Spec.UpdateConstraints
	if denied, err := matches("denylist", uc.Denylist); err != nil {
		return err
	} else if denied {
		return fmt.Errorf("updating from %s to %s is denied by the updateConstraints of %s (denylist %v)", v.Name, target.Name, v.Name, uc.Denylist)
	}
	if len(uc.Allowlist) == 0 {
		return nil
	}
	if allowed, err := matches("allowlist", uc.Allowlist); err != nil {
		return err
	} else if !allowed {
		return fmt.Errorf("updating from %s to %s is not allowed by the updateConstraints of %s (allowlist %v)", v.Name, target.Name, v.Name, uc.Allowlist)
	}
	return nil
}
