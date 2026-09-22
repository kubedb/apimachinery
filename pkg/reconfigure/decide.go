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

package reconfigure

import (
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

type Decision struct {
	Restart   bool
	LiveApply []Change
	Rejected  []Change
	Blocked   []Change
	Classes   map[string]Class
}

func (d Decision) Denied() bool {
	return len(d.Rejected) > 0 || len(d.Blocked) > 0
}

// Decide maps a classified change set onto what the operator should do. Anything the classifier
// did not place is treated as Static.
func Decide(restart opsapi.ReconfigureRestartType, changes []Change, classes map[string]Class) Decision {
	d := Decision{Classes: make(map[string]Class, len(changes))}

	var static []Change
	for _, c := range changes {
		class := classes[c.Key()].OrStatic()
		if c.New == nil && class == ClassDynamic {
			// Removing a parameter means "go back to the default", and the default is only
			// re-derived when the server reads the file again. There is nothing to apply live.
			class = ClassStatic
		}
		d.Classes[c.Key()] = class
		switch class {
		case ClassReadOnly, ClassLocked:
			d.Rejected = append(d.Rejected, c)
		case ClassDynamic:
			d.LiveApply = append(d.LiveApply, c)
		default:
			static = append(static, c)
		}
	}
	if len(d.Rejected) > 0 {
		return Decision{Rejected: d.Rejected, Classes: d.Classes}
	}

	switch restart {
	case opsapi.ReconfigureRestartTrue:
		d.LiveApply = nil
		d.Restart = len(changes) > 0
	case opsapi.ReconfigureRestartFalse:
		if len(static) > 0 {
			return Decision{Blocked: static, Classes: d.Classes}
		}
	default: // auto
		if len(static) > 0 {
			d.LiveApply = nil
			d.Restart = true
		}
	}
	return d
}
