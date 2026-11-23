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
	"testing"

	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

func TestStrongerRestart(t *testing.T) {
	f, a, tr := opsapi.ReconfigureRestartFalse, opsapi.ReconfigureRestartAuto, opsapi.ReconfigureRestartTrue
	tests := []struct{ a, b, want opsapi.ReconfigureRestartType }{
		{f, f, f}, {f, a, a}, {f, "", ""}, {f, tr, tr}, {a, f, a}, {"", f, ""}, {a, tr, tr}, {tr, f, tr}, {tr, a, tr},
	}
	for _, tt := range tests {
		if got := StrongerRestart(tt.a, tt.b); got != tt.want {
			t.Errorf("StrongerRestart(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}
