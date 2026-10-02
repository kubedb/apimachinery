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

package mongodb

import (
	"fmt"

	"kubedb.dev/apimachinery/pkg/reconfigure"
)

type Explainer struct{}

func (Explainer) Explain(c reconfigure.Change, class reconfigure.Class) string {
	path := Canonical(c.Path)
	switch class {
	case reconfigure.ClassLocked:
		msg := "this parameter is managed by KubeDB and is rewritten on every reconcile, so setting it here is either ignored or breaks the cluster"
		if reason, ok := lockedReason[path]; ok {
			msg = fmt.Sprintf("%s: %s", msg, reason)
		}
		if field := locked[path]; field != "" {
			msg = fmt.Sprintf("%s. Set it via the MongoDB CRD field %s instead", msg, field)
		}
		return msg + "."
	case reconfigure.ClassReadOnly:
		return "this is a runtime property of the server, not a setting, and cannot be configured."
	default:
		return fmt.Sprintf("unexpected class %q.", class)
	}
}
