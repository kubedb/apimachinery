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

package logforwarder

import "testing"

func TestParseComponents(t *testing.T) {
	for _, raw := range []string{"batch: {}", "batch/custom: {}", "resource/kubedb: {}", "resource/custom: null", "resource/custom: {}\nresource/custom: {}", "resource/custom: &x {}\nresource/other: *x", "resource/custom: {}\n---\nresource/other: {}", "- wrong"} {
		if _, err := ParseComponents(raw, map[string]bool{"batch/*": true, "resource/kubedb": true}); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	components, err := ParseComponents("resource/custom: {}\nattributes/custom: {}", map[string]bool{"attributes": true})
	if err != nil || len(components) != 2 {
		t.Fatal(components, err)
	}
	if _, err := OrderedStages(components, []string{"resource/custom", "attributes/custom"}); err != nil {
		t.Fatal(err)
	}
}
