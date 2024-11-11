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

package namespace

import (
	"net/http"
	"testing"

	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	admission "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	fake_dynamic "k8s.io/client-go/dynamic/fake"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	meta_util "kmodules.xyz/client-go/meta"
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientsetscheme.Scheme))
}

var requestKind = metav1.GroupVersionKind{
	Group:   core.SchemeGroupVersion.Group,
	Version: core.SchemeGroupVersion.Version,
	Kind:    "Namespace",
}

func TestNamespaceValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := NamespaceValidator{
				Resources: []string{olddbapi.ResourcePluralPostgres},
			}
			validator.initialized = true

			objJS, err := meta_util.MarshalToJson(c.object, core.SchemeGroupVersion)
			if err != nil {
				t.Fatalf("failed create marshal for input %s: %s", c.testName, err)
			}

			req := new(admission.AdmissionRequest)
			req.Kind = c.kind
			req.Name = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationv1.UserInfo{}
			req.Object.Raw = objJS

			var storeObjects []runtime.Object

			if c.operation == admission.Delete {
				storeObjects = append(storeObjects, c.object)
			}
			storeObjects = append(storeObjects, c.heatUp...)
			validator.dc = fake_dynamic.NewSimpleDynamicClient(clientsetscheme.Scheme, storeObjects...)

			response := validator.Admit(req)
			if c.result == true {
				if response.Allowed != true {
					t.Errorf("expected: 'Allowed=true'. but got response: %v", response)
				}
			} else if c.result == false {
				if response.Allowed == true || response.Result.Code == http.StatusInternalServerError {
					t.Errorf("expected: 'Allowed=false', but got response: %v", response)
				}
			}
		})
	}
}

var cases = []struct {
	testName  string
	kind      metav1.GroupVersionKind
	namespace string
	operation admission.Operation
	object    runtime.Object
	heatUp    []runtime.Object
	result    bool
}{
	{
		"Create Namespace",
		requestKind,
		"demo",
		admission.Create,
		sampleNamespace(),
		nil,
		true,
	},
	{
		"Delete Namespace containing db with terminationPolicy DoNotTerminate",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]runtime.Object{setTerminationPolicy(sampleDatabase(), olddbapi.DeletionPolicyDoNotTerminate)},
		false,
	},
	{
		"Delete Namespace containing db with terminationPolicy Pause",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]runtime.Object{setTerminationPolicy(sampleDatabase(), olddbapi.DeletionPolicyHalt)},
		false,
	},
	{
		"Delete Namespace containing db with terminationPolicy Delete",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]runtime.Object{setTerminationPolicy(sampleDatabase(), olddbapi.DeletionPolicyDelete)},
		true,
	},
	{
		"Delete Namespace containing db with terminationPolicy WipeOut",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]runtime.Object{setTerminationPolicy(sampleDatabase(), olddbapi.DeletionPolicyWipeOut)},
		true,
	},
	{
		"Delete Namespace containing db with NO terminationPolicy",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]runtime.Object{deleteTerminationPolicy(sampleDatabase())},
		true,
	},
}

func sampleNamespace() *core.Namespace {
	return &core.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: core.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
		},
	}
}

func sampleDatabase() *olddbapi.Postgres {
	return &olddbapi.Postgres{
		TypeMeta: metav1.TypeMeta{
			APIVersion: olddbapi.SchemeGroupVersion.String(),
			Kind:       "Postgres",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "demo",
			Labels: map[string]string{
				meta_util.ManagedByLabelKey: kubedb.GroupName,
			},
		},
		Spec: olddbapi.PostgresSpec{
			TerminationPolicy: olddbapi.DeletionPolicyDelete,
		},
	}
}

func setTerminationPolicy(obj runtime.Object, terminationPolicy olddbapi.DeletionPolicy) runtime.Object {
	db := obj.(*olddbapi.Postgres)
	db.Spec.TerminationPolicy = terminationPolicy
	return obj
}

func deleteTerminationPolicy(obj runtime.Object) runtime.Object {
	db := obj.(*olddbapi.Postgres)
	db.Spec.TerminationPolicy = ""
	return obj
}
