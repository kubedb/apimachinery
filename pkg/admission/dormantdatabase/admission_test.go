/*
Copyright The KubeDB Authors.

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

package dormantdatabase

import (
	"net/http"
	"testing"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	fake_ext "kubedb.dev/apimachinery/client/clientset/versioned/fake"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	admission "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	fake_dynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/meta"
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientsetscheme.Scheme))
}

var requestKind = metav1.GroupVersionKind{
	Group:   api.SchemeGroupVersion.Group,
	Version: api.SchemeGroupVersion.Version,
	Kind:    api.ResourceKindDormantDatabase,
}

func TestDormantDatabaseValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := DormantDatabaseValidator{}
			validator.initialized = true
			validator.client = fake.NewSimpleClientset()
			validator.dc = fake_dynamic.NewSimpleDynamicClient(clientsetscheme.Scheme)
			validator.extClient = fake_ext.NewSimpleClientset()

			objJS, err := meta.MarshalToJson(&c.object, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}
			oldObjJS, err := meta.MarshalToJson(&c.oldObject, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}

			req := new(admission.AdmissionRequest)

			req.Kind = c.kind
			req.Name = c.objectName
			req.Namespace = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationv1.UserInfo{}
			req.Object.Raw = objJS
			req.OldObject.Raw = oldObjJS

			if c.heatUp {
				if _, err := validator.extClient.KubedbV1alpha1().DormantDatabases(c.namespace).Create(&c.object); err != nil && !kerr.IsAlreadyExists(err) {
					t.Errorf(err.Error())
				}
			}
			if c.operation == admission.Delete {
				req.Object = runtime.RawExtension{}
			}
			if c.operation != admission.Update {
				req.OldObject = runtime.RawExtension{}
			}

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
	testName   string
	kind       metav1.GroupVersionKind
	objectName string
	namespace  string
	operation  admission.Operation
	object     api.DormantDatabase
	oldObject  api.DormantDatabase
	heatUp     bool
	result     bool
}{
	{"Create Dormant Database",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleDormantDatabase(),
		api.DormantDatabase{},
		false,
		true,
	},
	{"Edit Status",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editStatus(sampleDormantDatabase()),
		sampleDormantDatabase(),
		false,
		true,
	},
	{"Edit Spec.Origin ",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecOrigin(sampleDormantDatabase()),
		sampleDormantDatabase(),
		false,
		false,
	},
	{"Edit Spec.WipeOut",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecWipeOut(sampleDormantDatabase()),
		sampleDormantDatabase(),
		false,
		true,
	},
	{"Delete Without Wiping Out",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		sampleDormantDatabase(),
		api.DormantDatabase{},
		true,
		true,
	},
	{"Delete With Wiping Out ",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		editStatusWipedOut(sampleDormantDatabase()),
		api.DormantDatabase{},
		true,
		true,
	},
	{"Delete Non Existing Dormant",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		api.DormantDatabase{},
		api.DormantDatabase{},
		false,
		true,
	},
}

func sampleDormantDatabase() api.DormantDatabase {
	return api.DormantDatabase{
		TypeMeta: metav1.TypeMeta{
			Kind:       api.ResourceKindDormantDatabase,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMongoDB,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						api.LabelDatabaseKind: api.ResourceKindMongoDB,
					},
					Annotations: map[string]string{
						api.AnnotationInitialized: "",
					},
				},
				Spec: api.OriginSpec{
					MongoDB: &api.MongoDBSpec{},
				},
			},
		},
	}
}

func editSpecOrigin(old api.DormantDatabase) api.DormantDatabase {
	old.Spec.Origin.Annotations = nil
	return old
}

func editStatus(old api.DormantDatabase) api.DormantDatabase {
	old.Status = api.DormantDatabaseStatus{
		Phase: api.DormantDatabasePhasePaused,
	}
	return old
}

func editSpecWipeOut(old api.DormantDatabase) api.DormantDatabase {
	old.Spec.WipeOut = true
	return old
}

func editStatusWipedOut(old api.DormantDatabase) api.DormantDatabase {
	old.Spec.WipeOut = true
	old.Status.Phase = api.DormantDatabasePhaseWipedOut
	return old
}
