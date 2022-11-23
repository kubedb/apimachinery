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
	"errors"

	errors2 "github.com/pkg/errors"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kafkalog = logf.Log.WithName("kafka-resource")

func (k *Kafka) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-kafka-kubedb-com-v1alpha1-kafka,mutating=true,failurePolicy=fail,sideEffects=None,groups=kafka.kubedb.com,resources=kafkas,verbs=create,versions=v1alpha1,name=mkafka.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Kafka{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *Kafka) Default() {
	if k == nil {
		return
	}
	kafkalog.Info("default", "name", k.Name)
	// TODO(user): fill in your defaulting logic.
	k.SetDefaults()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kafka-kubedb-com-v1alpha1-kafka,mutating=false,failurePolicy=fail,sideEffects=None,groups=kafka.kubedb.com,resources=kafkas,verbs=create;update,versions=v1alpha1,name=vkafka.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Kafka{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateCreate() error {
	kafkalog.Info("validate create", "name", k.Name)
	var allErr field.ErrorList
	// TODO(user): fill in your validation logic upon object creation.
	if k.Spec.Topology != nil {
		if k.Spec.Topology.Controller == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("controller"),
				k.Name,
				".spec.topology.controller can't be empty in topology cluster"))
		}
		if k.Spec.Topology.Broker == nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("broker"),
				k.Name,
				".spec.topology.broker can't be empty in topology cluster"))
		}

		if k.Spec.Replicas != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicas"),
				k.Name,
				"doesn't support spec.replicas when spec.topology is set"))
		}
		if k.Spec.Storage != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology").Child("broker"),
				k.Name,
				"doesn't support spec.storage when spec.topology is set"))
		}
		if k.Spec.PodTemplate.Spec.Resources.Size() != 0 {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("resources"),
				k.Name,
				"doesn't support spec.podTemplate.spec.resources when spec.topology is set"))
		}
		// validate that multiple nodes don't have same suffixes
		err := validateNodeSuffix(k.Spec.Topology)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology"),
				k.Name,
				err.Error()))
		}

		// validate that node replicas are not 0 or negative
		err = validateNodeReplicas(k.Spec.Topology)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("topology"),
				k.Name,
				err.Error()))
		}

		err = validateVolumes(k)
		if err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("podTemplate").Child("spec").Child("volumes"),
				k.Name,
				err.Error()))
		}

	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "kafka.kubedb.com", Kind: "Kafka"}, k.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateUpdate(old runtime.Object) error {
	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateDelete() error {
	kafkalog.Info("validate delete", "name", k.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validateNodeSuffix(topology *KafkaClusterTopology) error {
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

func validateNodeReplicas(topology *KafkaClusterTopology) error {
	tMap := topology.ToMap()
	for key, node := range tMap {
		if pointer.Int32(node.Replicas) <= 0 {
			return errors2.Errorf("replicas for node role %s must be alteast 1", string(key))
		}
	}
	return nil
}

var reservedVolumes = []string{
	KafkaVolumeData,
	KafkaVolumeConfig,
	KafkaVolumeTempConfig,
}

func validateVolumes(db *Kafka) error {
	if db.Spec.PodTemplate.Spec.Volumes == nil {
		return nil
	}
	rsv := reservedVolumes
	if db.Spec.TLS != nil && db.Spec.TLS.Certificates != nil {
		for _, c := range db.Spec.TLS.Certificates {
			rsv = append(rsv, db.CertSecretVolumeName(KafkaCertificateAlias(c.Alias)))
		}
	}
	return ValidateVolumes(db.Spec.PodTemplate.Spec.Volumes, rsv)
}

func ValidateVolumes(volumes []core.Volume, reservedVolumeNames []string) error {
	for _, rv := range reservedVolumeNames {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}
