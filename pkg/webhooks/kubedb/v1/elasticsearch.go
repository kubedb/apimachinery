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

package v1

import (
	"context"
	"fmt"

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	amv "kubedb.dev/apimachinery/pkg/validator"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupElasticsearchWebhookWithManager registers the webhook for Elasticsearch in the manager.
func SetupElasticsearchWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.Elasticsearch{}).
		WithValidator(&ElasticsearchCustomWebhook{DefaultClient: mgr.GetClient()}).
		WithDefaulter(&ElasticsearchCustomWebhook{DefaultClient: mgr.GetClient()}).
		Complete()
}

type ElasticsearchCustomWebhook struct {
	DefaultClient    client.Client
	StrictValidation bool
}

var eslog = logf.Log.WithName("elasticsearch-resource")

var _ webhook.CustomDefaulter = &ElasticsearchCustomWebhook{}

func (w *ElasticsearchCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	db := obj.(*dbapi.Elasticsearch)
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	eslog.Info("default", "name", db.GetName())

	if db.Spec.Halted {
		if db.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
			return errors.New(`Can't halt, since termination policy is 'DoNotTerminate'`)
		}
		db.Spec.DeletionPolicy = dbapi.DeletionPolicyHalt
	}

	if db.Spec.Replicas == nil && db.Spec.Topology == nil {
		db.Spec.Replicas = pointer.Int32P(1)
	}

	var esVersion catalogapi.ElasticsearchVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &esVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get ElasticsearchVersion: %s", db.Spec.Version)
	}

	db.SetDefaults(&esVersion)
	db.SetHealthCheckerDefaults()

	return nil
}

var forbiddenElasticsearchEnvVars = []string{
	"node.name",
	"node.ingest",
	"node.master",
	"node.data",
	"node.ml",
	"node.data_hot",
	"node.data_warm",
	"node.data_cold",
	"node.data_frozen",
	"node.data_content",
}

// Allow only built in users to be synced
// custom users are not supported to be created via es client
var allowedInternalUsers = []string{
	string(dbapi.ElasticsearchInternalUserElastic),
	string(dbapi.ElasticsearchInternalUserLogstashSystem),
	string(dbapi.ElasticsearchInternalUserRemoteMonitoringUser),
	string(dbapi.ElasticsearchInternalUserKibanaSystem),
	string(dbapi.ElasticsearchInternalUserApmSystem),
	string(dbapi.ElasticsearchInternalUserBeatsSystem),
}

var reservedElasticsearchVolumes = []string{
	kubedb.ElasticsearchVolumeData,
	kubedb.ElasticsearchVolumeConfig,
	kubedb.ElasticsearchVolumeSecurityConfig,
	kubedb.ElasticsearchVolumeSecureSettings,
	kubedb.ElasticsearchVolumeCustomConfig,
	kubedb.ElasticsearchVolumeTempConfig,
}

var reservedElasticsearchMountPaths = []string{
	kubedb.ElasticsearchVolumeData,
	kubedb.ElasticsearchConfigDir,
	kubedb.ElasticsearchCustomConfigDir,
	kubedb.ElasticsearchOpenSearchConfigDir,
	kubedb.ElasticsearchOpenSearchSecurityConfigDir,
	kubedb.ElasticsearchOpendistroSecurityConfigDir,
}

var _ webhook.CustomValidator = &ElasticsearchCustomWebhook{}

func (w *ElasticsearchCustomWebhook) ValidateElasticsearch(db *dbapi.Elasticsearch) error {
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	if db.Spec.Topology == nil {
		if db.Spec.Replicas == nil || ptr.Deref(db.Spec.Replicas, 0) < 1 {
			return fmt.Errorf(`spec.replicas "%d" invalid. Value must be greater than zero`, ptr.Deref(db.Spec.Replicas, 0))
		}
	}

	var esVersion catalogapi.ElasticsearchVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &esVersion)
	if err != nil {
		return fmt.Errorf(`expected version "%v" "%v"`, db.Spec.Version, err.Error())
	}

	if db.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if db.Spec.StorageType != dbapi.StorageTypeDurable && db.Spec.StorageType != dbapi.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, db.Spec.StorageType)
	}

	err = w.validateVolumes(db)
	if err != nil {
		return err
	}

	err = w.validateVolumeMountPaths(db)
	if err != nil {
		return err
	}

	err = w.validateSecureConfig(db, &esVersion)
	if err != nil {
		return err
	}

	topology := db.Spec.Topology
	if topology != nil {
		if db.Spec.Replicas != nil {
			return errors.New("doesn't support spec.replicas when spec.topology is set")
		}
		if db.Spec.Storage != nil {
			return errors.New("doesn't support spec.storage when spec.topology is set")
		}
		err = w.validateNodeRoles(topology, &esVersion)
		if err != nil {
			return err
		}
		// Check node name suffix
		err = w.validateNodeSuffix(topology)
		if err != nil {
			return err
		}
		err = w.validateNodeReplicas(topology)
		if err != nil {
			return err
		}
		err = w.validateNodeSpecs(w.DefaultClient, db, &esVersion)
		if err != nil {
			return err
		}

	} else {
		if db.Spec.Replicas == nil || ptr.Deref(db.Spec.Replicas, 0) < 1 {
			return fmt.Errorf(`spec.replicas "%d" invalid. Must be greater than zero`, ptr.Deref(db.Spec.Replicas, 0))
		}

		if err := amv.ValidateStorage(w.DefaultClient, olddbapi.StorageType(db.Spec.StorageType), db.Spec.Storage); err != nil {
			return err
		}

		if db.Spec.MaxUnavailable != nil {
			if int32(db.Spec.MaxUnavailable.IntValue()) > ptr.Deref(db.Spec.Replicas, 0) {
				return fmt.Errorf("MaxUnavailable replicas can't be greater that number of replicas")
			}
		}

		dbContainer := core_util.GetContainerByName(db.Spec.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
		// Resources validation
		// Heap size is the 50% of memory & it cannot be less than 128Mi(some say 97Mi)
		// So, minimum memory request should be twice of 128Mi, i.e. 256Mi.
		if value, ok := dbContainer.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*kubedb.ElasticsearchMinHeapSize {
			return fmt.Errorf("PodTemplate.Spec.Resources.Requests.memory cannot be less than %dMi, given %dMi", (2*kubedb.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024))
		}

		if err = w.validateContainerSecurityContext(dbContainer.SecurityContext, &esVersion); err != nil {
			return err
		}

		if err := amv.ValidateEnvVar(dbContainer.Env, forbiddenElasticsearchEnvVars, dbapi.ResourceKindElasticsearch); err != nil {
			return err
		}
	}

	if db.Spec.InternalUsers != nil {
		// Allow support for only builtin internal users
		// custom users can not be created via client api request
		// The allowed builtin users are supported for xpack authplugin version >= 7.8
		// ref: https://github.com/elastic/go-elasticsearch/find/main
		if esVersion.Spec.AuthPlugin == catalogapi.ElasticsearchAuthPluginXpack {
			version, err := semver.NewVersion(esVersion.Spec.Version)
			if err != nil {
				return err
			}
			if version.Major() >= 8 || (version.Major() == 7 && version.Minor() >= 8) {
				if err := amv.ValidateInternalUsersV1(db.Spec.InternalUsers, allowedInternalUsers, dbapi.ResourceKindElasticsearch); err != nil {
					return err
				}
			} else {
				return fmt.Errorf(`'spec.internalUsers' is not unsupported for versions < 7.8`)
			}
		}
	}

	if db.Spec.AuthSecret != nil &&
		db.Spec.AuthSecret.ExternallyManaged &&
		db.Spec.AuthSecret.Name == "" {
		return fmt.Errorf(`for externallyManaged auth secret, user must configure "spec.authSecret.name"`)
	}

	if esVersion.Spec.Deprecated {
		return fmt.Errorf("elasticsearch %s/%s is using deprecated version %v. Skipped processing", db.Namespace,
			db.Name, esVersion.Name)
	}

	if err := esVersion.ValidateSpecs(); err != nil {
		return fmt.Errorf("elasticsearch %s/%s is using invalid elasticsearchVersion %v. Skipped processing. reason: %v", db.Namespace,
			db.Name, esVersion.Name, err)
	}

	if db.Spec.DeletionPolicy == "" {
		return fmt.Errorf(`'spec.deletionPolicy' is missing`)
	}

	if db.Spec.StorageType == dbapi.StorageTypeEphemeral && db.Spec.DeletionPolicy == dbapi.DeletionPolicyHalt {
		return fmt.Errorf(`'spec.deletionPolicy: Halt' can not be set for 'Ephemeral' storage`)
	}

	if db.Spec.DisableSecurity && db.Spec.EnableSSL {
		return fmt.Errorf(`to enable 'spec.enableSSL', 'spec.disableSecurity' needs to be set to false`)
	}

	// TODO:
	//		- OpenSearch provision fails with security plugin disabled.
	//		- Remove the validation, once the issue is fixed.
	//		- Issue Ref: https://github.com/opensearch-project/security/issues/1481
	if db.Spec.DisableSecurity && esVersion.Spec.AuthPlugin == catalogapi.ElasticsearchAuthPluginOpenSearch {
		return fmt.Errorf(`'spec.disableSecurity' cannot be 'true' for opensearch`)
	}

	monitorSpec := db.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	err = amv.ValidateHealth(&db.Spec.HealthChecker)
	if err != nil {
		return err
	}

	return nil
}

func (w *ElasticsearchCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	es := obj.(*dbapi.Elasticsearch)
	err := w.ValidateElasticsearch(es)
	mysqlLog.Info("validating", "name", es.Name)
	return admission.Warnings{}, err
}

func (w *ElasticsearchCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldElasticsearch, ok := oldObj.(*dbapi.Elasticsearch)
	if !ok {
		return nil, fmt.Errorf("expected a Elasticsearch but got a %T", oldElasticsearch)
	}
	elasticsearch, ok := newObj.(*dbapi.Elasticsearch)
	if !ok {
		return nil, fmt.Errorf("expected a Elasticsearch but got a %T", elasticsearch)
	}

	var esVersion catalogapi.ElasticsearchVersion
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: oldElasticsearch.Spec.Version,
	}, &esVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ElasticsearchVersion: %s", oldElasticsearch.Spec.Version)
	}
	oldElasticsearch.SetDefaults(&esVersion)
	if oldElasticsearch.Spec.AuthSecret == nil {
		oldElasticsearch.Spec.AuthSecret = elasticsearch.Spec.AuthSecret
	}
	if err := w.validatePreconditions(elasticsearch, oldElasticsearch); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	err = w.ValidateElasticsearch(elasticsearch)
	return admission.Warnings{}, err
}

func (w *ElasticsearchCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	elasticsearch, ok := obj.(*dbapi.Elasticsearch)
	if !ok {
		return nil, fmt.Errorf("expected a Elasticsearch but got a %T", obj)
	}

	var es dbapi.Elasticsearch
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      elasticsearch.Name,
		Namespace: elasticsearch.Namespace,
	}, &es)
	if err != nil && !kerr.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to get Elasticsearch: %s", elasticsearch.Name)
	} else if err == nil && es.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`ElasticSearch "%v/%v" can't be terminated. To delete, change spec.deletionPolicy`, es.Namespace, es.Name)
	}
	return nil, nil
}

func (w *ElasticsearchCustomWebhook) validatePreconditions(obj, oldObj *dbapi.Elasticsearch) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.topology.*.suffix",
			"spec.storageType",
		),
	}
	// Once the database has been initialized, don't let update the "spec.init" section
	if oldObj.Spec.Init != nil && oldObj.Spec.Init.Initialized {
		preconditions.Insert("spec.init")
	}
	_, err := meta_util.CreateJSONMergePatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *ElasticsearchCustomWebhook) validateContainerSecurityContext(sc *core.SecurityContext, esVersion *catalogapi.ElasticsearchVersion) error {
	if sc == nil {
		return nil
	}
	if esVersion.Spec.AuthPlugin == catalogapi.ElasticsearchAuthPluginOpenSearch {
		return nil
	}

	// if RunAsAnyNonRoot == false
	//		only allow default UID (runAsUser)
	// else
	//		allow any UID but root (0)
	if !esVersion.Spec.SecurityContext.RunAsAnyNonRoot {
		// if default RunAsUser is missing, user isn't allowed to user RunAsUser.
		if esVersion.Spec.SecurityContext.RunAsUser == nil && sc.RunAsUser != nil {
			return fmt.Errorf("not allowed to set containerSecurityContext.runAsUser for ElasticsearchVersion: %s", esVersion.Name)
		}
		// if default RunAsUser is set, validate it.
		if sc.RunAsUser != nil && esVersion.Spec.SecurityContext.RunAsUser != nil &&
			*sc.RunAsUser != *esVersion.Spec.SecurityContext.RunAsUser {
			return fmt.Errorf("containerSecurityContext.runAsUser must be %d for ElasticsearchVersion: %s", *esVersion.Spec.SecurityContext.RunAsUser, esVersion.Name)
		}
	} else {
		if sc.RunAsUser != nil && *sc.RunAsUser == 0 {
			return fmt.Errorf("not allowed to set containerSecurityContext.runAsUser to root (0) for ElasticsearchVersion: %s", esVersion.Name)
		}
	}

	return nil
}

func (w *ElasticsearchCustomWebhook) validateNodeSuffix(topology *dbapi.ElasticsearchClusterTopology) error {
	tMap := topology.ToMap()
	names := make(map[string]bool)
	for _, value := range tMap {
		names[value.Suffix] = true
	}
	if len(tMap) != len(names) {
		return errors.New("two or more node cannot have same suffix")
	}
	return nil
}

func (w *ElasticsearchCustomWebhook) validateNodeReplicas(topology *dbapi.ElasticsearchClusterTopology) error {
	tMap := topology.ToMap()
	var err error
	for key, node := range tMap {
		if ptr.Deref(node.Replicas, 0) <= 0 {
			err = appendError(err, errors.Errorf("replicas for node role %s must be alteast 1", string(key)))
		}
	}
	return err
}

func appendError(err error, newError error) error {
	if err == nil {
		return newError
	} else {
		return errors.Wrap(err, newError.Error())
	}
}

func (w *ElasticsearchCustomWebhook) validateNodeSpecs(kc client.Client, db *dbapi.Elasticsearch, esVersion *catalogapi.ElasticsearchVersion) error {
	topology := db.Spec.Topology
	tMap := topology.ToMap()
	var errForNodeContainer error

	for nodeRole, node := range tMap {
		if err := amv.ValidateStorage(kc, olddbapi.StorageType(db.Spec.StorageType), node.Storage); err != nil {
			errForNodeContainer = appendError(errForNodeContainer, err)
		}
		// Resources validation
		// Heap size is the 50% of memory & it cannot be less than 128Mi(some say 97Mi)
		// So, minimum memory request should be twice of 128Mi, i.e. 256Mi.
		dbContainer := core_util.GetContainerByName(node.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
		if value, ok := dbContainer.Resources.Requests[core.ResourceMemory]; ok && value.Value() < 2*kubedb.ElasticsearchMinHeapSize {
			errForNodeContainer = appendError(errForNodeContainer, fmt.Errorf("%s.resources.reqeusts.memory cannot be less than %dMi, given %dMi", string(nodeRole), (2*kubedb.ElasticsearchMinHeapSize)/(1024*1024), value.Value()/(1024*1024)))
		}

		if err := w.validateContainerSecurityContext(dbContainer.SecurityContext, esVersion); err != nil {
			errForNodeContainer = appendError(errForNodeContainer, err)
		}

		if node.MaxUnavailable != nil {
			if int32(node.MaxUnavailable.IntValue()) > *node.Replicas {
				errForNodeContainer = appendError(errForNodeContainer, fmt.Errorf("MaxUnavailable replicas can't be greater that number of replicas in %s node", nodeRole))
			}
		}

		if err := amv.ValidateEnvVar(dbContainer.Env, forbiddenElasticsearchEnvVars, dbapi.ResourceKindElasticsearch); err != nil {
			errForNodeContainer = appendError(errForNodeContainer, err)
		}
	}
	return errForNodeContainer
}

func (w *ElasticsearchCustomWebhook) validateNodeRoles(topology *dbapi.ElasticsearchClusterTopology, esVersion *catalogapi.ElasticsearchVersion) error {
	switch esVersion.Spec.Distribution {
	case catalogapi.ElasticsearchDistroOpenDistro:
		if topology.ML != nil || topology.DataContent != nil || topology.DataCold != nil || topology.DataFrozen != nil ||
			topology.Coordinating != nil || topology.Transform != nil {
			return errors.Errorf("node role: ml, data_cold, data_frozen, data_content, transform, coordinating are not supported for ElasticsearchVersion %s", esVersion.Name)
		}
	case catalogapi.ElasticsearchDistroSearchGuard:
		if topology.Data == nil {
			return errors.New("topology.data cannot be empty")
		}
		if topology.ML != nil || topology.DataHot != nil || topology.DataContent != nil ||
			topology.DataCold != nil || topology.DataWarm != nil || topology.DataFrozen != nil ||
			topology.Coordinating != nil || topology.Transform != nil {
			return errors.Errorf("node role: ml, data_hot, data_cold, data_warm, data_frozen, data_content, transform, coordinating are not supported for ElasticsearchVersion %s", esVersion.Name)
		}
	}

	// Every cluster requires the following node roles:
	//	- (data_content and data_hot) OR (data)
	//	- ref: https://www.elastic.co/guide/en/elasticsearch/reference/7.14/modules-node.html#node-roles
	if esVersion.Spec.Distribution == catalogapi.ElasticsearchDistroElasticStack &&
		topology.Data == nil &&
		(topology.DataHot == nil || topology.DataContent == nil) {
		return errors.New("when data node is empty, you need to have both dataHot and dataContent nodes")
	}

	return nil
}

func (w *ElasticsearchCustomWebhook) validateSecureConfig(db *dbapi.Elasticsearch, esVersion *catalogapi.ElasticsearchVersion) error {
	dbVersion, err := semver.NewVersion(esVersion.Spec.Version)
	if err != nil {
		return err
	}
	if db.Spec.SecureConfigSecret != nil {
		// Elasticsearch keystore is not supported for OpenSearch
		if esVersion.Spec.Distribution == catalogapi.ElasticsearchDistroOpenSearch {
			return errors.New("secureConfigSecret is not supported for OpenSearch")
		}
		// KEYSTORE_PASSWORD is supported since ES version 7.9
		if dbVersion.Major() < 7 || (dbVersion.Major() == 7 && dbVersion.Minor() < 9) {
			return errors.Errorf("secureConfigSecret is not supported for ElasticsearchVersion %s, try with latest versions", esVersion.Name)
		}
	}
	return nil
}

func (w *ElasticsearchCustomWebhook) validateVolumes(db *dbapi.Elasticsearch) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := reservedElasticsearchVolumes
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rsv = append(rsv, db.CertSecretVolumeName(dbapi.ElasticsearchCertificateAlias(c.Alias)))
		}
	}
	return amv.ValidateVolumes(ofst.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), rsv)
}

func (w *ElasticsearchCustomWebhook) validateVolumeMountPaths(db *dbapi.Elasticsearch) error {
	rPaths := reservedElasticsearchMountPaths
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rPaths = append(rPaths, db.CertSecretVolumeMountPath(kubedb.ElasticsearchConfigDir, dbapi.ElasticsearchCertificateAlias(c.Alias)))
			rPaths = append(rPaths, db.CertSecretVolumeMountPath(kubedb.ElasticsearchOpenSearchConfigDir, dbapi.ElasticsearchCertificateAlias(c.Alias)))
		}
	}

	var err error
	if db.Spec.Topology != nil {
		tMap := db.Spec.Topology.ToMap()
		for _, node := range tMap {
			dbContainer := core_util.GetContainerByName(node.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
			if dbContainer.VolumeMounts != nil {
				errForNodeContainer := amv.ValidateMountPaths(dbContainer.VolumeMounts, rPaths)
				err = appendError(err, errForNodeContainer)
			}
		}

	} else {
		dbContainer := core_util.GetContainerByName(db.Spec.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
		if dbContainer.VolumeMounts != nil {
			err = amv.ValidateMountPaths(dbContainer.VolumeMounts, rPaths)
		}
	}

	return err
}
