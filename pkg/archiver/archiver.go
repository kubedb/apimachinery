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

func SyncStorageCredSecret(kc client.Client, gvk schema.GroupVersionKind, dbMeta metav1.ObjectMeta) error {
	db, err := func() (*unstructured.Unstructured, error) {
		var db unstructured.Unstructured
		db.SetGroupVersionKind(gvk)
		err := kc.Get(context.Background(), client.ObjectKey{Name: dbMeta.Name, Namespace: dbMeta.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return &db, nil
	}()
	if err != nil {
		return err
	}

	archiver, err := func(db *unstructured.Unstructured) (*unstructured.Unstructured, error) {
		_, found, err := unstructured.NestedFieldNoCopy(db.Object, "spec", "archiver")
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, nil // archiver not configured, not an error
		}

		refName, found, err := unstructured.NestedString(db.Object, "spec", "archiver", "ref", "name")
		if err != nil {
			return nil, err
		}
		if !found || refName == "" {
			return nil, fmt.Errorf("spec.archiver.ref.name is required but missing")
		}

		refNamespace, found, err := unstructured.NestedString(db.Object, "spec", "archiver", "ref", "namespace")
		if err != nil {
			return nil, err
		}
		if !found || refNamespace == "" {
			// ref.namespace — fall back to db's namespace
			refNamespace = db.GetNamespace()
		}

		var archiver unstructured.Unstructured
		archiver.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   archiverapi.SchemeGroupVersion.Group,
			Version: archiverapi.SchemeGroupVersion.Version,
			Kind:    fmt.Sprintf("%vArchiver", gvk.Kind),
		})

		err = kc.Get(context.Background(), client.ObjectKey{Name: refName, Namespace: refNamespace}, &archiver)
		if kerr.IsNotFound(err) {
			return nil, nil // referenced archiver doesn't exist yet
		}
		if err != nil {
			return nil, err
		}
		return &archiver, nil
	}(db)
	if err != nil {
		return err
	} else if archiver == nil {
		return nil
	}

	bs, err := func(archiver *unstructured.Unstructured) (*storageapi.BackupStorage, error) {
		_, found, err := unstructured.NestedFieldNoCopy(archiver.Object, "spec", "backupStorage")
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("spec.backupStorage is required but missing")
		}

		bsName, found, err := unstructured.NestedString(archiver.Object, "spec", "backupStorage", "ref", "name")
		if err != nil {
			return nil, err
		}
		if !found || bsName == "" {
			return nil, fmt.Errorf("spec.backupStorage.ref.name is required but missing")
		}

		bsNamespace, found, err := unstructured.NestedString(archiver.Object, "spec", "backupStorage", "ref", "namespace")
		if err != nil {
			return nil, err
		}
		if !found || bsNamespace == "" {
			bsNamespace = archiver.GetNamespace()
		}

		var bs storageapi.BackupStorage
		bs.Name = bsName
		bs.Namespace = bsNamespace

		err = kc.Get(context.Background(), client.ObjectKey{Name: bsName, Namespace: bsNamespace}, &bs)
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &bs, nil
	}(archiver)
	if err != nil {
		return err
	} else if bs == nil {
		return nil
	}

	if bs.GetNamespace() == db.GetNamespace() {
		return nil
	}
	storageSecretName, err := GetStorageSecretName(bs)
	if err != nil {
		return err
	}

	var secret core.Secret
	err = kc.Get(context.Background(), client.ObjectKey{Name: storageSecretName, Namespace: db.GetNamespace()}, &secret)
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}

	// Secret exists + db is being deleted → cleanup
	if err == nil && dbMeta.DeletionTimestamp != nil {
		return removeAnnotationsOrDeleteCopiedStorageCredSecret(kc, &secret, db.GetName())
	}

	// Fetch source secret data (needed for create, harmless for update)
	var storageSecret core.Secret
	err = kc.Get(context.TODO(), types.NamespacedName{
		Name:      storageSecretName,
		Namespace: bs.Namespace,
	}, &storageSecret)
	if err != nil {
		return err
	}

	_, err = client_util.CreateOrPatch(context.TODO(), kc, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storageSecretName,
			Namespace: db.GetNamespace(),
		},
	}, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.Secret)
		annotations := in.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		in.Annotations = SetAnnotationForStorageCredSecret(annotations, db.GetName())
		in.Data = storageSecret.Data
		in.StringData = storageSecret.StringData
		in.Type = storageSecret.Type
		return in
	})
	return err
}

func SetAnnotationForStorageCredSecret(annotations map[string]string, dbName string) map[string]string {
	if annotations == nil {
		annotations = make(map[string]string)
	}
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
	if annotations == nil {
		return nil
	}
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

func removeAnnotationsOrDeleteCopiedStorageCredSecret(kc client.Client, secret *core.Secret, dbName string) error {
	annotations := secret.GetAnnotations()
	annotations = RemoveAnnotationFromStorageCredSecret(annotations, dbName)

	if annotations == nil || annotations[kubedb.OwnerDatabasesAnnotation] == "" {
		err := kc.Delete(context.Background(), secret)
		if err != nil && !kerr.IsNotFound(err) {
			return err
		}
		return nil
	}

	_, err := client_util.CreateOrPatch(context.TODO(), kc, secret, func(obj client.Object, createOp bool) client.Object {
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
