/*
Copyright 2023.

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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeSolr          = "sl"
	ResourceKindSolr          = "Solr"
	ResourceSingularSolr      = "solr"
	ResourcePluralSolr        = "solrs"
	SolrPortName              = "http"
	SolrRestPort              = 8983
	SolrSecretName            = "solr-secret"
	SolrSecretKey             = "solr.xml"
	SolrContainerName         = "solr"
	SolrInitContainerName     = "init-solr"
	SolrInitAuthContainerName = "init-auth"
	SolrSecret                = "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n<solr>\n <str name=\"coreRootDirectory\">/var/solr/data</str>\n <str name=\"sharedLib\">${solr.sharedLib:}</str>\n  <solrcloud>\n    <str name=\"host\">${host:}</str>\n    <int name=\"hostPort\">${solr.port.advertise:80}</int>\n    <str name=\"hostContext\">${hostContext:solr}</str>\n    <bool name=\"genericCoreNodeNames\">${genericCoreNodeNames:true}</bool>\n    <int name=\"zkClientTimeout\">${zkClientTimeout:30000}</int>\n    <int name=\"distribUpdateSoTimeout\">${distribUpdateSoTimeout:600000}</int>\n    <int name=\"distribUpdateConnTimeout\">${distribUpdateConnTimeout:60000}</int>\n    <str name=\"zkCredentialsProvider\">${zkCredentialsProvider:org.apache.solr.common.cloud.DefaultZkCredentialsProvider}</str>\n    <str name=\"zkACLProvider\">${zkACLProvider:org.apache.solr.common.cloud.DefaultZkACLProvider}</str>\n  </solrcloud>\n  <shardHandlerFactory name=\"shardHandlerFactory\"\n    class=\"HttpShardHandlerFactory\">\n    <int name=\"socketTimeout\">${socketTimeout:600000}</int>\n    <int name=\"connTimeout\">${connTimeout:60000}</int>\n  </shardHandlerFactory>\n  <int name=\"maxBooleanClauses\">${solr.max.booleanClauses:1024}</int>\n  <str name=\"allowPaths\">${solr.allowPaths:}</str>\n  <metrics enabled=\"${metricsEnabled:true}\"/>\n  \n</solr>\n"
	SolrAdmin                 = "admin"
	SolrUser                  = "solr"
	SecurityJSON              = "security.json"
	SolrACL                   = "acl"
	SolrACLReadOnly           = "acl-read-only"
	TempSecret                = "<?xml version=\"1.0\" encoding=\"UTF-8\" ?><solr>\n    <str name=\"coreRootDirectory\">var/solr/data</str>\n    <int name=\"maxBooleanClauses\">${solr.max.booleanClauses:1024}</int>\n    <str name=\"allowPaths\">${allowPaths:}</str>\n    <solrCloud><int name=\"hostPort\">${solr.port.advertise:80}</int>\n        <str name=\"host\">${host:}</str>\n        <int name=\"distribUpdateSoTimeout\">${distribUpdateSoTimeout:600000}</int>\n        <int name=\"distribUpdateConnTimeout\">${distribUpdateConnTimeout:60000}</int>\n        <str name=\"zkACLProvider\">${zkACLProvider:org.apache.solr.common.cloud.DefaultZkACLProvider}</str>\n        <bool name=\"genericCoreNodeNames\">${genericCoreNodeNames:true}</bool>\n        <str name=\"zkCredentialsProvider\">${zkCredentialsProvider:org.apache.solr.common.cloud.DefaultZkCredentialsProvider}</str>\n    </solrCloud>\n    <metrics enabled=\"${metricsEnabled:true}\"/>\n</solr>\n"
)

const (
	SolrDefaultConfig       = "solr-config"
	SolrVolumeDefaultConfig = "solr-config"
	SolrVolumeCustomConfig  = "custom-config"
	SolrVolumeAuthConfig    = "auth-config"
	SolrVolumeData          = "data"
	SolrVolumeConfig        = "slconfig"
)

const (
	DistLibs              = "/opt/solr/dist"
	ContribLibs           = "/opt/solr/contrib/%s/lib"
	SysPropLibPlaceholder = "${solr.sharedLib:}"
)

const (
	SolrHomeDir           = "/var/solr"
	SolrDataDir           = "/var/solr/data"
	SolrTempConfigDir     = "/temp-config"
	SolrCustomConfigDir   = "/custom-config"
	SolrSecurityConfigDir = "/var/security"
)

var Keys = map[string]string{
	"maxBooleanClauses": "solr.max.booleanClauses",
	"sharedLib":         "solr.sharedLib",
	"hostPort":          "solr.port.advertise",
	"allowPaths":        "solr.allowPaths",
}

var ShardHandlerFactory = map[string]interface{}{
	"socketTimeout": 600000,
	"connTimeout":   60000,
}

var SolrCloud = map[string]interface{}{
	"host":                     "",
	"hostPort":                 80,
	"hostContext":              "solr",
	"genericCoreNodeNames":     true,
	"zkClientTimeout":          30000,
	"distribUpdateSoTimeout":   600000,
	"distribUpdateConnTimeout": 60000,
	"zkCredentialsProvider":    "org.apache.solr.common.cloud.DefaultZkCredentialsProvider",
	"zkACLProvider":            "org.apache.solr.common.cloud.DefaultZkACLProvider",
}

var SolrConf = map[string]interface{}{
	"maxBooleanClauses": 1024,
	//"sharedLib":           "",
	"allowPaths":          "",
	"solrcloud":           SolrCloud,
	"shardHandlerFactory": ShardHandlerFactory,
}

// Solr is the schema for the Sole API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sl,scope=Namespaced
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".apiVersion"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Solr struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SolrSpec   `json:"spec,omitempty"`
	Status SolrStatus `json:"status,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SolrSpec defines the desired state of Solr c
type SolrSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version of Solr to be deployed
	Version string `json:"version"`

	// Number of instances to deploy for a Solr database
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Solr topology for node specification
	// +optional
	Topology *SolrClusterTopology `json:"topology,omitempty"`

	// StorageType van be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage to specify how storage shall be used
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	ZookeeperRef *core.LocalObjectReference `json:"zookeeperRef"`

	// +optional
	SolrModules []string `json:"solrModules,omitempty"`

	// +optional
	SolrOpts []string `json:"solrOpts,omitempty"`

	// To enable ssl for http layer
	EnableSSL bool `json:"enableSSL,omitempty"`

	// TLS contains tls configurations for client and server.
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// Disable security. It disables authentication security of users.
	// If unset, default is false
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// +optional
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty"`

	// +optional
	AuthConfigSecret *core.LocalObjectReference `json:"authConfigSecret,omitempty"`

	// +optional
	ZookeeperACLSecret *core.LocalObjectReference `json:"zookeeperACLSecret,omitempty"`

	// +optional
	ZookeeperACLReadOnlySecret *core.LocalObjectReference `json:"zookeeperACLReadOnlySecret,omitempty"`

	//***********keystoresecret will be added later

	//***********tls will be added later

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// TerminationPolicy controls the delete operation for database
	// +optional
	TerminationPolicy TerminationPolicy `json:"terminationPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 20, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`

	//*********monitor will be added later
}

type SolrClusterTopology struct {
	Overseer    *SolrNode `json:"overseer,omitempty"`
	Data        *SolrNode `json:"data,omitempty"`
	Coordinator *SolrNode `json:"coordinator,omitempty"`
}

type SolrNode struct {
	// Replica represents number of replica for this specific type of nodes
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// suffix to append with node name
	// +optional
	Suffix string `json:"suffix,omitempty"`

	// Storage to specify how storage shall be used.
	// +optional
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// +kubebuilder:validation:Enum=Provisioning;Ready;NotReady;Critical
type SolrPhase string

const (
	SolrPhaseProvisioning SolrPhase = "Provisioning"
	SolrPhaseReady        SolrPhase = "Ready"
	SolrPhaseNotReady     SolrPhase = "NotReady"
	SolrPhaseCritical     SolrPhase = "Critical"
)

// SolrStatus defines the observed state of Solr
type SolrStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the current phase of the database
	// +optional
	Phase SolrPhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:validation:Enum=overseer;data;coordinator;combined
type SolrNodeRoleType string

const (
	SolrNodeRoleOverseer    SolrNodeRoleType = "overseer"
	SolrNodeRoleData        SolrNodeRoleType = "data"
	SolrNodeRoleCoordinator SolrNodeRoleType = "coordinator"
	SolrNodeRoleSet                          = "set"
)

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SolrList contains a list of Solr
type SolrList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solr `json:"items"`
}
