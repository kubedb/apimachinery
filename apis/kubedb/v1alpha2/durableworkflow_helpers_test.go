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

package v1alpha2

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPersistenceMarshalOmitsAnEmptyDefault(t *testing.T) {
	b, err := json.Marshal(DurableWorkflowPersistence{Managed: &DurableWorkflowManagedStore{}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"default"`) {
		t.Errorf("an empty default must not be sent: %s", b)
	}
	b, _ = json.Marshal(DurableWorkflowPersistence{Default: DurableWorkflowStore{Type: DurableWorkflowStorePostgres}})
	if !strings.Contains(string(b), `"default"`) {
		t.Errorf("a default that is set must be sent: %s", b)
	}
}
