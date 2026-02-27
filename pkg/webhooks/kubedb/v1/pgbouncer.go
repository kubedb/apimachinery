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
	"strings"

	catalogapi "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	amv "kubedb.dev/apimachinery/pkg/validator"

	cm_api "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	meta_util "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupPgBouncerWebhookWithManager registers the webhook for PgBouncer in the manager.
func SetupPgBouncerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&dbapi.PgBouncer{}).
		WithValidator(&PgBouncerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&PgBouncerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PgBouncerCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &PgBouncerCustomWebhook{}

// log is for logging in this package.
var pgBouncerLog = logf.Log.WithName("pqbouncer-resource")

func (pw PgBouncerCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	db, ok := obj.(*dbapi.PgBouncer)
	if !ok {
		return fmt.Errorf("expected an PgBouncer object but got %T", obj)
	}

	pgBouncerLog.Info("defaulting", "name", db.GetName())

	if db.Spec.Replicas == nil {
		db.Spec.Replicas = pointer.Int32P(1)
	}
	if db.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}

	var pgBouncerVersion catalogapi.PgBouncerVersion
	err := pw.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.Spec.Version,
	}, &pgBouncerVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to get PgBouncer Version: %s", db.Spec.Version)
	}

	usesAcme := false
	if db.Spec.TLS != nil {
		usesAcme, err = dbapi.UsesAcmeIssuer(pw.DefaultClient, db.Namespace, *db.Spec.TLS.IssuerRef)
		if err != nil {
			return err
		}
	}

	db.SetDefaults(&pgBouncerVersion, usesAcme)

	return nil
}

var pbForbiddenEnvVars = []string{
	"PGBOUNCER_PASSWORD",
	"PGBOUNCER_USER",
	"POSTGRES_PASSWORD",
	"POSTGRES_USER",
	"PGPASSWORD",
}

var pbReservedVolumes = []string{
	kubedb.PgBouncerAuthSecretVolume,
	kubedb.GitSecretVolume,
}

var pbReservedMountPaths = []string{
	kubedb.PgBouncerConfigMountPath,
	kubedb.PgBouncerSecretMountPath,
	kubedb.PgBouncerServingCertMountPath,
	kubedb.GitSecretMountPath,
}

func (pw PgBouncerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pgBouncer, ok := obj.(*dbapi.PgBouncer)
	if !ok {
		return nil, fmt.Errorf("expected a PgBouncer but got a %T", pgBouncer)
	}

	var finalConfigSecret v1.Secret
	err := pw.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: pgBouncer.Namespace, Name: pgBouncer.PgBouncerFinalConfigSecretName()}, &finalConfigSecret)
	if err == nil {
		return nil, err
	}
	return pw.Validate(ctx, pgBouncer)
}

func (pw PgBouncerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	// TODO : user can't update anything related to config
	pgBouncer, ok := newObj.(*dbapi.PgBouncer)
	if !ok {
		return nil, fmt.Errorf("expected a PgBouncer but got a %T", pgBouncer)
	}
	oldPgBouncer, ok := oldObj.(*dbapi.PgBouncer)
	if !ok {
		return nil, fmt.Errorf("expected a PgBouncer but got a %T", oldPgBouncer)
	}

	var pgBouncerVersion catalogapi.PgBouncerVersion
	err := pw.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      oldPgBouncer.Spec.Version,
		Namespace: oldPgBouncer.Namespace,
	}, &pgBouncerVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get PgBouncerVersion: %s", oldPgBouncer.Spec.Version)
	}

	usesAcme := false
	if oldPgBouncer.Spec.TLS != nil {
		usesAcme, err = dbapi.UsesAcmeIssuer(pw.DefaultClient, oldPgBouncer.Namespace, *oldPgBouncer.Spec.TLS.IssuerRef)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get issuer: %s", oldPgBouncer.Spec.TLS.IssuerRef.Name)
		}
	}

	oldPgBouncer.SetDefaults(&pgBouncerVersion, usesAcme)

	if oldPgBouncer.Spec.AuthSecret == nil {
		oldPgBouncer.Spec.AuthSecret = pgBouncer.Spec.AuthSecret
	}

	if err := validatePgBouncerUpdate(pgBouncer, oldPgBouncer); err != nil {
		return nil, err
	}
	return pw.Validate(ctx, pgBouncer)
}

func (pw PgBouncerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// req.Object.Raw = nil, so read from kubernetes
	pgBouncer, ok := obj.(*dbapi.PgBouncer)
	if !ok {
		return nil, fmt.Errorf("expected a PgBouncer but got a %T", pgBouncer)
	}

	err := pw.DefaultClient.Get(ctx, types.NamespacedName{Name: pgBouncer.Name, Namespace: pgBouncer.Namespace}, &dbapi.PgBouncer{})
	if kerr.IsNotFound(err) {
		klog.Infoln("obj ", pgBouncer.Name, " already deleted")
		return nil, errors.New(fmt.Sprintf("obj %s/%s already deleted", pgBouncer.Namespace, pgBouncer.Name))
	}

	if pgBouncer.Spec.DeletionPolicy == dbapi.DeletionPolicyDoNotTerminate {
		return nil, fmt.Errorf(`PgBouncer "%v/%v" can't be terminated. To delete, change spec.DeletionPolicy`, pgBouncer.Namespace, pgBouncer.Name)
	}
	return nil, nil
}

func RemoveEmptyString(strListadr *[]string) *[]string {
	strList := *strListadr
	for i := 0; i < len(strList); {
		if strList[i] == "" {
			strList = append(strList[:i], strList[i+1:]...)
		} else {
			i++
		}
	}
	return &strList
}

func DivideIniConfDataInSections(dataReceive, sectionsReceive *[]string) *map[string][]string {
	data := *dataReceive
	sections := *sectionsReceive

	sectionData := make([]string, 0)
	sectionName := ""
	result := make(map[string][]string)

	for i := range data {
		newSectionName := ""
		for j := range sections {
			sectionTitle := fmt.Sprintf("%s%s%s", "[", sections[j], "]")
			if data[i] == sectionTitle {
				newSectionName = sections[j]
			}
		}
		if newSectionName != "" {
			if sectionName != "" {
				result[sectionName] = sectionData
			}
			sectionName = newSectionName
			sectionData = make([]string, 0)
			continue
		}
		sectionData = append(sectionData, data[i])
	}
	if sectionName != "" {
		result[sectionName] = sectionData
	}
	return &result
}

// ValidatePgBouncer checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func (pw PgBouncerCustomWebhook) Validate(ctx context.Context, db *dbapi.PgBouncer) (admission.Warnings, error) {
	if db.Spec.Replicas == nil || *db.Spec.Replicas < 1 {
		return nil, fmt.Errorf(`spec.replicas "%v" invalid. Value must be greater than zero`, db.Spec.Replicas)
	}

	// TODO : there should be only one db

	if err := validateGivenPbConfigSecret(db, pw.DefaultClient); err != nil {
		return nil, err
	}

	if db.Spec.Version == "" {
		return nil, fmt.Errorf(`spec.Version can't be empty`)
	}

	if db.Spec.TLS != nil {
		if *db.Spec.TLS.IssuerRef.APIGroup != cm_api.SchemeGroupVersion.Group {
			return nil, fmt.Errorf(`spec.tls.client.issuerRef.apiGroup must be %s`, cm_api.SchemeGroupVersion.Group)
		}
		if (db.Spec.TLS.IssuerRef.Kind != cm_api.IssuerKind) && (db.Spec.TLS.IssuerRef.Kind != cm_api.ClusterIssuerKind) {
			return nil, fmt.Errorf(`spec.tls.client.issuerRef.issuerKind must be either %s or %s`, cm_api.IssuerKind, cm_api.ClusterIssuerKind)
		}
	}

	// Check if pgbouncerVersion is absent or deprecated.
	// If deprecated, return error
	var pgBouncerVersion catalogapi.PgBouncerVersion
	err := pw.DefaultClient.Get(ctx, types.NamespacedName{
		Name: db.Spec.Version,
	}, &pgBouncerVersion)
	if err != nil {
		return nil, err
	}
	if pgBouncerVersion.Spec.Deprecated {
		return nil, fmt.Errorf("pgbouncer %s/%s is using deprecated version %v. Skipped processing",
			db.Namespace, db.Name, pgBouncerVersion.Name)
	}

	if db.Spec.Database.DatabaseName == "" || db.Spec.Database.DatabaseName == "pgbouncer" {
		return nil, fmt.Errorf("pgbouncer %s/%s databaseName: %s already exist", db.Namespace, db.Name, db.Spec.Database.DatabaseName)
	}

	if err := amv.ValidateHealth(&db.Spec.HealthChecker); err != nil {
		return nil, err
	}

	pbReservedVolumes = append(pbReservedVolumes, db.GetCertSecretName(dbapi.PgBouncerServerCert))
	pbReservedVolumes = append(pbReservedVolumes, db.GetCertSecretName(dbapi.PgBouncerClientCert))
	pbReservedVolumes = append(pbReservedVolumes, db.GetCertSecretName(dbapi.PgBouncerMetricsExporterCert))
	pbReservedVolumes = append(pbReservedVolumes, db.PgBouncerFinalConfigSecretName())

	err = validatePgBouncerVolumes(db)
	if err != nil {
		return nil, err
	}

	err = validateVolumeMountsForAllPbContainers(db)
	if err != nil {
		return nil, err
	}

	err = validateEnvsForAllPbContainers(db)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func validatePgBouncerUpdate(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{
		Set: sets.New[string](
			"spec.authSecret",
		),
	}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func validateEnvsForAllPbContainers(pgbouncer *dbapi.PgBouncer) error {
	var err error
	for _, container := range pgbouncer.Spec.PodTemplate.Spec.Containers {
		if container.Env != nil {
			if errC := amv.ValidateEnvVar(container.Env, pbForbiddenEnvVars, dbapi.ResourceKindPgBouncer); errC != nil {
				if err == nil {
					err = errC
				} else {
					err = errors.Wrap(err, errC.Error())
				}
			}
		}
	}
	return err
}

func validateVolumeMountsForAllPbContainers(pgbouncer *dbapi.PgBouncer) error {
	var err error
	for _, container := range pgbouncer.Spec.PodTemplate.Spec.Containers {
		if container.VolumeMounts != nil {
			if errC := amv.ValidateMountPaths(container.VolumeMounts, pbReservedMountPaths); errC != nil {
				if err == nil {
					err = errC
				} else {
					err = errors.Wrap(err, errC.Error())
				}
			}
		}
	}
	if errC := amv.ValidateGitInitRootPath(pgbouncer.Spec.Init, pbReservedMountPaths); errC != nil {
		if err == nil {
			err = errC
		} else {
			err = errors.Wrap(err, errC.Error())
		}
	}
	return err
}

func validatePgBouncerVolumes(db *dbapi.PgBouncer) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	return amv.ValidateVolumes(ofst.ConvertVolumes(db.Spec.PodTemplate.Spec.Volumes), pbReservedVolumes)
}

func validateGivenPbConfigSecret(db *dbapi.PgBouncer, client client.Client) error {
	if db.Spec.ConfigSecret != nil {
		var configSecret v1.Secret
		err := client.Get(context.TODO(), types.NamespacedName{Name: db.Spec.ConfigSecret.Name, Namespace: db.Namespace}, &configSecret)
		if err != nil {
			return fmt.Errorf("spec.configSecret %v not found", db.Spec.ConfigSecret.Name)
		}
		if _, present := configSecret.Data[kubedb.PgBouncerConfigFile]; !present {
			return fmt.Errorf("spec.configSecret %v doesn't contain %s config file", db.Spec.ConfigSecret.Name, kubedb.PgBouncerConfigFile)
		}
		// can't change database section
		configFile := configSecret.Data[kubedb.PgBouncerConfigFile]
		DataList := strings.Split(string(configFile), "\n")
		DataList = *RemoveEmptyString(&DataList)
		dividedDataInSections := *DivideIniConfDataInSections(&DataList, dbapi.PgBouncerConfigSections())
		if _, present := dividedDataInSections[kubedb.PgBouncerConfigSectionDatabases]; present {
			sectionData := dividedDataInSections[kubedb.PgBouncerConfigSectionDatabases]
			if len(sectionData) != 0 {
				return fmt.Errorf("database reference can not be added via config secret")
			}
		}
	}
	return nil
}
