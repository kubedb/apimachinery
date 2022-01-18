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

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmodulesv1 "kmodules.xyz/client-go/api/v1"
	kmodulesv1alpha1 "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make" to regenerate code after modifying this file

// MySQLDatabaseSpec defines the desired state of MySQLDatabase
type MySQLDatabaseSpec struct {

	//MySQLRef refers to kubedb MySQL instance
	MySQLRef kmodulesv1alpha1.AppReference `json:"mySqlRef"`

	//VaultRef refers to kubevault Vault server
	VaultRef kmodulesv1.ObjectReference `json:"vaultRef"`

	//DatabaseConfig contains the expected/target database properties
	DatabaseConfig MySQLDatabaseSchema `json:"databaseConfig"`

	//todo add specification
	Subjects []rbac.Subject `json:"subjects"`

	//InitSpec contains info about the init script volume source
	//+optional
	InitSpec *InitSpec `json:"initSpec,omitempty"`

	//Restore contains info for stash restore
	//+optional
	Restore *RestoreConf `json:"restore,omitempty"`

	//ValidationTimeLimit contains the time duration for which the schema user
	//would have access to the MySQL server
	//+optional
	ValidationTimeLimit *TTL `json:"validationTimeLimit,omitempty"`

	//TerminationPolicy contains the deletion policy
	TerminationPolicy TerminationPolicy `json:"terminationPolicy"`
}

// MySQLDatabaseStatus defines the observed state of MySQLDatabase
type MySQLDatabaseStatus struct {
	Phase              MySQLDatabasePhase          `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=MySQLDatabasePhase"`
	ObservedGeneration int64                       `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
	Conditions         []kmodulesv1.Condition      `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	LoginCreds         *kmodulesv1.ObjectReference `json:"loginCreds,omitempty" protobuf:"bytes,5,opt,name=secret"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Login Credential",type="string",JSONPath=".status.loginCreds.name"

// MySQLDatabase is the Schema for the mysqldatabases API
type MySQLDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLDatabaseSpec   `json:"spec,omitempty"`
	Status MySQLDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MySQLDatabaseList contains a list of MySQLDatabase
type MySQLDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLDatabase{}, &MySQLDatabaseList{})
}

type MySQLDatabaseSchema struct {
	// Name is target database name
	Name string `json:"name"`

	//CharacterSet is the target database character set
	//+optional
	CharacterSet string `json:"characterSet,omitempty"`

	//Collation is the target database collation
	//+optional
	Collation string `json:"collation,omitempty"`

	//Encryption is the target databae encryption mode
	//+optional
	Encryption string `json:"encryption,omitempty"`

	//ReadOnly is the target database read only mode
	//+optional
	ReadOnly int32 `json:"readOnly,omitempty"`
}

// type InitSpec struct {
// 	//Script is the volume source which contains script.sql data field
// 	Script core.VolumeSource `json:"script,omitempty"`

// 	//PodTemplate contains user preferred pod configuration
// 	//+optional
// 	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`
// }

type RestoreConf struct {
	//Repository is the stash repository reference
	Repository core.ObjectReference `json:"repository,omitempty"`

	//Snapshot is the rules for restore
	//+optional
	Snapshot string `json:"snapshot,omitempty"`
}

type TTL struct {
	//DefaultTTL refers to the actual validation time limit
	DefaultTTL string `json:"defaultTTL"`

	//MaxTTL refers to the maximum validation time limit
	//+optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

type MySQLDatabasePhase string
type TerminationPolicy string
type MySQLDatabaseCondition string
type MySQLDatabaseVerbs string

const (
	AddCondition    MySQLDatabaseVerbs = "AddCondition"
	RemoveCondition MySQLDatabaseVerbs = "RemoveCondition"

	TerminationPolicyDelete      TerminationPolicy = "Delete"
	TerminationPolicyDoNotDelete TerminationPolicy = "DoNotDelete"

	MySQLRoleCreated           MySQLDatabaseCondition = "MySQLRoleCreated"
	VaultSecretEngineCreated   MySQLDatabaseCondition = "VaultSecretEngineCreated"
	SecretAccessRequestCreated MySQLDatabaseCondition = "SecretAccessRequestCreated"

	MySQLNotReady               MySQLDatabaseCondition = "MySQLNotReady"
	VaultNotReady               MySQLDatabaseCondition = "VaultNotReady"
	SecretAccessRequestApproved MySQLDatabaseCondition = "SecretAccessRequestApproved"
	SecretAccessRequestDenied   MySQLDatabaseCondition = "SecretAccessRequestDenied"
	SecretAccessRequestExpired  MySQLDatabaseCondition = "SecretAccessRequestExpired"

	SchemaIgnored          MySQLDatabaseCondition = "SchemaIgnored"
	DatabaseCreated        MySQLDatabaseCondition = "DatabaseCreated"
	DatabaseDeleted        MySQLDatabaseCondition = "DatabaseDeleted"
	DatabaseAltered        MySQLDatabaseCondition = "DatabaseAltered"
	ScriptApplied          MySQLDatabaseCondition = "ScriptApplied"
	RestoredFromRepository MySQLDatabaseCondition = "RestoredFromRepository"
	FailedInitializing     MySQLDatabaseCondition = "FailedInitializing"
	FailedRestoring        MySQLDatabaseCondition = "FailedRestoring"
	TerminationHalted      MySQLDatabaseCondition = "TerminationHalted"
	UserDisconnected       MySQLDatabaseCondition = "UserDisconnected"

	Success     MySQLDatabasePhase = "Success"
	Running     MySQLDatabasePhase = "Running"
	Waiting     MySQLDatabasePhase = "Waiting"
	Ignored     MySQLDatabasePhase = "Ignored"
	Failed      MySQLDatabasePhase = "Failed"
	Expired     MySQLDatabasePhase = "Expired"
	Terminating MySQLDatabasePhase = "Terminating"
)
