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

package archiver

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	archiverapi "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/pkg/double_optin"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	client_util "kmodules.xyz/client-go/client"
	"kmodules.xyz/client-go/cluster"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCorrespondingArchiver(kbClient client.Client, dbMeta metav1.ObjectMeta, archiverList []archiverapi.Accessor) (*metav1.ObjectMeta, error) {
	projectNSList, err := GetSameProjectNamespaces(kbClient, dbMeta.Namespace)
	if err != nil {
		return nil, err
	}

	var priorityList []priority
	for _, archiver := range archiverList {
		var archiverNs core.Namespace
		err := kbClient.Get(context.TODO(), types.NamespacedName{
			Name: archiver.GetObjectMeta().Namespace,
		}, &archiverNs)
		if err != nil {
			return nil, err
		}

		var dbNs core.Namespace
		err = kbClient.Get(context.TODO(), types.NamespacedName{
			Name: dbMeta.Namespace,
		}, &dbNs)
		if err != nil {
			return nil, err
		}

		possible, err := double_optin.CheckIfDoubleOptInPossible(dbMeta, dbNs.ObjectMeta, archiverNs.ObjectMeta, archiver.GetConsumers())
		if err != nil {
			return nil, err
		}
		if possible {
			priorityList = append(priorityList, getPriority(archiver.GetObjectMeta(), projectNSList, dbMeta.Namespace))
		}
	}
	if priorityList == nil {
		return nil, err
	}
	sort.Slice(priorityList, func(i, j int) bool {
		return priorityList[i].index < priorityList[j].index
	})
	return &priorityList[0].archiver, nil
}

func GetSameProjectNamespaces(kbClient client.Client, dbNs string) ([]string, error) {
	if cluster.IsRancherManaged(kbClient.RESTMapper()) {
		namespaces, err := cluster.ListSiblingNamespaces(kbClient, dbNs)
		if err != nil {
			return nil, err
		}
		ret := make([]string, len(namespaces))
		for i, namespace := range namespaces {
			ret[i] = namespace.Name
		}
		return ret, nil
	}
	return nil, nil
}

type priority struct {
	archiver metav1.ObjectMeta
	index    int
}

func getPriority(archiver metav1.ObjectMeta, projectNSList []string, dbNs string) priority {
	idx := 2
	if archiver.Namespace == dbNs {
		idx = 0
	} else if slices.Contains(projectNSList, archiver.Namespace) {
		idx = 1
	}
	return priority{archiver, idx}
}

func SetAnnotationForStorageCredSecret(annotations map[string]string, dbName string) map[string]string {
	databases := annotations[kubedb.OwnerDatabasesAnnotation]
	if databases == "" {
		annotations[kubedb.OwnerDatabasesAnnotation] = dbName
		return annotations
	}
	parts := strings.Split(databases, ",")
	if slices.Contains(parts, dbName) {
		return annotations
	}
	databases = fmt.Sprintf("%s,%s", databases, dbName)
	annotations[kubedb.OwnerDatabasesAnnotation] = databases
	return annotations
}

func RemoveAnnotationFromStorageCredSecret(annotations map[string]string, dbName string) map[string]string {
	databases := annotations[kubedb.OwnerDatabasesAnnotation]
	if databases == "" {
		return annotations
	}
	parts := strings.Split(databases, ",")
	parts = slices.DeleteFunc(parts, func(s string) bool {
		return s == dbName
	})
	if len(parts) == 0 {
		delete(annotations, kubedb.OwnerDatabasesAnnotation)
	} else {
		annotations[kubedb.OwnerDatabasesAnnotation] = strings.Join(parts, ",")
	}
	return annotations
}

func UpdateOrDeleteCopiedStorageCredSecretD(kc client.Client, gvk schema.GroupVersionKind, dbMeta metav1.ObjectMeta) error {
	var db unstructured.Unstructured
	db.SetGroupVersionKind(gvk)
	err := kc.Get(context.Background(), client.ObjectKey{Name: dbMeta.Name, Namespace: dbMeta.Namespace}, &db)
	if err != nil {
		return err
	}
	refName, _, err := unstructured.NestedString(db.Object, "spec", "archiver", "ref", "name")
	if err != nil {
		return err
	}

	refNamespace, _, err := unstructured.NestedString(db.Object, "spec", "archiver", "ref", "namespace")
	if err != nil {
		return err
	}
	var archiver unstructured.Unstructured
	archiver.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   archiverapi.SchemeGroupVersion.Group,
		Version: archiverapi.SchemeGroupVersion.Version,
		Kind:    fmt.Sprintf("%vArchiver", gvk.Kind),
	})

	err = kc.Get(context.Background(), client.ObjectKey{Name: refName, Namespace: refNamespace}, &archiver)
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	if kerr.IsNotFound(err) {
		return nil
	}
	bsName, _, err := unstructured.NestedString(archiver.Object, "spec", "backupStorage", "ref", "name")
	if err != nil {
		return err
	}
	bsNamespace, _, err := unstructured.NestedString(archiver.Object, "spec", "backupStorage", "ref", "namespace")
	if err != nil {
		return err
	}

	bs := storageapi.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bsName,
			Namespace: bsNamespace,
		},
	}
	err = kc.Get(context.Background(), client.ObjectKey{Name: bs.Name, Namespace: bs.Namespace}, &bs)
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	if kerr.IsNotFound(err) {
		return nil
	}

	if bs.GetNamespace() == db.GetNamespace() {
		return nil
	}
	secretName, err := GetStorageSecretName(&bs)
	if err != nil {
		return err
	}

	secret := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: db.GetNamespace(),
		},
	}
	err = kc.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, &secret)
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	if kerr.IsNotFound(err) {
		return nil
	}
	annotations := secret.GetAnnotations()
	annotations = RemoveAnnotationFromStorageCredSecret(annotations, db.GetName())
	if annotations == nil || len(annotations[kubedb.OwnerDatabasesAnnotation]) == 0 {
		err = kc.Delete(context.Background(), &secret)
		if err != nil && !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}
	_, err = client_util.CreateOrPatch(context.TODO(), kc, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: db.GetNamespace(),
		},
	}, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.Secret)
		in.Annotations = annotations
		return in
	})
	return err
}

func GetStorageSecretName(backupStorage *storageapi.BackupStorage) (string, error) {
	if backupStorage.Spec.Storage.Provider == storageapi.ProviderS3 {
		return backupStorage.Spec.Storage.S3.SecretName, nil
	}
	if backupStorage.Spec.Storage.Provider == storageapi.ProviderGCS {
		return backupStorage.Spec.Storage.GCS.SecretName, nil
	}
	if backupStorage.Spec.Storage.Provider == storageapi.ProviderAzure {
		return backupStorage.Spec.Storage.Azure.SecretName, nil
	}
	return "", fmt.Errorf("failed to get storage secret")
}
