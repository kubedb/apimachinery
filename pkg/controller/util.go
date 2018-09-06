package controller

//
//import (
//	core_util "github.com/appscode/kutil/core/v1"
//	meta_util "github.com/appscode/kutil/meta"
//	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
//	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
//		coreV1 "k8s.io/api/core/v1"
//	kerr "k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/labels"
//	"k8s.io/client-go/kubernetes"
//)
//
//// Remove Owner Reference from Snapshots
//
//
//
//// Remove Owner Reference from secrets
//func RemoveOwnerReferenceFromSecrets(
//	client kubernetes.Interface,
//	extClient cs.Interface,
//	meta metav1.ObjectMeta,
//	labelSelector labels.Selector,
//	ref *coreV1.ObjectReference,
//	secretVolList ...*coreV1.SecretVolumeSource,
//) error {
//	for _, secretVolSrc := range secretVolList {
//		if secretVolSrc == nil {
//			continue
//		}
//		secret, err := client.CoreV1().Secrets(meta.Namespace).Get(secretVolSrc.SecretName, metav1.GetOptions{})
//		if err != nil && kerr.IsNotFound(err) {
//			continue
//		} else if err != nil {
//			return err
//		}
//		if _, _, err := core_util.PatchSecret(client, secret, func(in *coreV1.Secret) *coreV1.Secret {
//			in.ObjectMeta = core_util.RemoveOwnerReference(in.ObjectMeta, ref)
//			return in
//		}); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// Set Owner Reference to snapshots
//
//
//// Set Owner Reference to PVCs
//func SetOwnerReferenceToPVCs(
//	client kubernetes.Interface,
//	extClient cs.Interface,
//	meta metav1.ObjectMeta,
//	labelSelector labels.Selector,
//	ref *coreV1.ObjectReference,
//) error {
//	pvcList, err := client.CoreV1().PersistentVolumeClaims(meta.Namespace).List(
//		metav1.ListOptions{
//			LabelSelector: labelSelector.String(),
//		},
//	)
//	if err != nil {
//		return err
//	}
//	for _, pvc := range pvcList.Items {
//		if _, _, err := core_util.PatchPVC(client, &pvc, func(in *coreV1.PersistentVolumeClaim) *coreV1.PersistentVolumeClaim {
//			in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
//			return in
//		}); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// Set Owner Reference to Secrets
//func SetOwnerReferenceToSecrets(
//	client kubernetes.Interface,
//	extClient cs.Interface,
//	meta metav1.ObjectMeta,
//	labelSelector labels.Selector,
//	dbKind string,
//	ref *coreV1.ObjectReference,
//	secretVolList ...*coreV1.SecretVolumeSource,
//) error {
//	for _, secretVolSrc := range secretVolList {
//		if secretVolSrc == nil {
//			continue
//		}
//		if err := SterilizeSecrets(client, extClient, meta, dbKind, ref, secretVolSrc); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// SterilizeSecrets cleans secret that is created  by KubeDB-Operator and
//// not used by any other dbKind or DormantDatabases objects.
//func SterilizeSecrets(
//	client kubernetes.Interface,
//	extClient cs.Interface,
//	meta metav1.ObjectMeta,
//	dbKind string,
//	ref *coreV1.ObjectReference,
//	secretVolume *coreV1.SecretVolumeSource,
//) error {
//	secretFound := false
//	if secretVolume == nil {
//		return nil
//	}
//
//	secret, err := client.CoreV1().Secrets(meta.Namespace).Get(secretVolume.SecretName, metav1.GetOptions{})
//	if err != nil && kerr.IsNotFound(err) {
//		return nil
//	} else if err != nil {
//		return err
//	}
//
//	// if api.LabelDatabaseKind not exists in secret, then the secret is not created by KubeDB-Operator
//	// otherwise, probably KubeDB-Operator created the secrets.
//	if _, err := meta_util.GetStringValue(secret.ObjectMeta.Labels, api.LabelDatabaseKind); err != nil {
//		return nil
//	}
//
//	secretFound, err = isSecretUsedInExistingDB(extClient, meta, dbKind, secretVolume)
//	if err != nil {
//		return err
//	}
//
//	if !secretFound {
//		labelMap := map[string]string{
//			api.LabelDatabaseKind: dbKind,
//		}
//		dormantDatabaseList, err := extClient.KubedbV1alpha1().DormantDatabases(meta.Namespace).List(
//			metav1.ListOptions{
//				LabelSelector: labels.SelectorFromSet(labelMap).String(),
//			},
//		)
//		if err != nil {
//			return err
//		}
//
//		for _, ddb := range dormantDatabaseList.Items {
//			if ddb.Name == meta.Name {
//				continue
//			}
//
//			databaseSecretList := GetDatabaseSecretName(&ddb, dbKind)
//			if databaseSecretList != nil {
//				for _, databaseSecret := range databaseSecretList {
//					if databaseSecret == nil {
//						continue
//					}
//					if databaseSecret.SecretName == secretVolume.SecretName {
//						secretFound = true
//						break
//					}
//				}
//			}
//			if secretFound {
//				break
//			}
//		}
//	}
//
//	if !secretFound {
//		if _, _, err := core_util.PatchSecret(client, secret, func(in *coreV1.Secret) *coreV1.Secret {
//			in.ObjectMeta = core_util.EnsureOwnerReference(in.ObjectMeta, ref)
//			return in
//		}); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func isSecretUsedInExistingDB(
//	extClient cs.Interface,
//	meta metav1.ObjectMeta,
//	dbKind string,
//	secretVolume *coreV1.SecretVolumeSource,
//) (bool, error) {
//	if dbKind == api.ResourceKindMemcached || dbKind == api.ResourceKindRedis {
//		return false, nil
//	}
//	switch dbKind {
//	case api.ResourceKindMongoDB:
//		mgList, err := extClient.KubedbV1alpha1().MongoDBs(meta.Namespace).List(metav1.ListOptions{})
//		if err != nil {
//			return true, err
//		}
//		for _, mg := range mgList.Items {
//			databaseSecret := mg.Spec.DatabaseSecret
//			if databaseSecret != nil {
//				if databaseSecret.SecretName == secretVolume.SecretName {
//					return true, nil
//				}
//			}
//		}
//	case api.ResourceKindMySQL:
//		msList, err := extClient.KubedbV1alpha1().MySQLs(meta.Namespace).List(metav1.ListOptions{})
//		if err != nil {
//			return true, err
//		}
//		for _, ms := range msList.Items {
//			databaseSecret := ms.Spec.DatabaseSecret
//			if databaseSecret != nil {
//				if databaseSecret.SecretName == secretVolume.SecretName {
//					return true, nil
//				}
//			}
//		}
//	case api.ResourceKindPostgres:
//		pgList, err := extClient.KubedbV1alpha1().Postgreses(meta.Namespace).List(metav1.ListOptions{})
//		if err != nil {
//			return true, err
//		}
//		for _, pg := range pgList.Items {
//			databaseSecret := pg.Spec.DatabaseSecret
//			if databaseSecret != nil {
//				if databaseSecret.SecretName == secretVolume.SecretName {
//					return true, nil
//				}
//			}
//		}
//	case api.ResourceKindElasticsearch:
//		esList, err := extClient.KubedbV1alpha1().Elasticsearches(meta.Namespace).List(metav1.ListOptions{})
//		if err != nil {
//			return true, err
//		}
//		for _, es := range esList.Items {
//			databaseSecret := es.Spec.DatabaseSecret
//			if databaseSecret != nil {
//				if databaseSecret.SecretName == secretVolume.SecretName {
//					return true, nil
//				}
//			}
//			certCertificate := es.Spec.CertificateSecret
//			if certCertificate != nil {
//				if certCertificate.SecretName == secretVolume.SecretName {
//					return true, nil
//				}
//			}
//		}
//	}
//	return false, nil
//}
//
//func GetDatabaseSecretName(dormantDatabase *api.DormantDatabase, dbKind string) []*coreV1.SecretVolumeSource {
//	if dbKind == api.ResourceKindMemcached || dbKind == api.ResourceKindRedis {
//		return nil
//	}
//	switch dbKind {
//	case api.ResourceKindMongoDB:
//		secretVol := []*coreV1.SecretVolumeSource{
//			dormantDatabase.Spec.Origin.Spec.MongoDB.DatabaseSecret,
//		}
//		if dormantDatabase.Spec.Origin.Spec.MongoDB.ReplicaSet != nil {
//			secretVol = append(secretVol, dormantDatabase.Spec.Origin.Spec.MongoDB.ReplicaSet.KeyFile)
//		}
//		return secretVol
//	case api.ResourceKindMySQL:
//		return []*coreV1.SecretVolumeSource{dormantDatabase.Spec.Origin.Spec.MySQL.DatabaseSecret}
//	case api.ResourceKindPostgres:
//		return []*coreV1.SecretVolumeSource{dormantDatabase.Spec.Origin.Spec.Postgres.DatabaseSecret}
//	case api.ResourceKindElasticsearch:
//		return []*coreV1.SecretVolumeSource{
//			dormantDatabase.Spec.Origin.Spec.Elasticsearch.DatabaseSecret,
//			dormantDatabase.Spec.Origin.Spec.Elasticsearch.CertificateSecret,
//		}
//	}
//	return nil
//}
