package v1alpha1

type MongoSource struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
	Mongoshake     *Mongoshake    `yaml:"mongoshake" json:"mongoshake,omitempty"`
}
type MongoTarget struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
}
type Mongoshake struct {
	// SyncMode: full, incr, or fullSync
	SyncMode string `yaml:"syncMode" json:"syncMode,omitempty"`
	// Source is the mongoshake collector binary path
	// +optional
	Collector string `yaml:"collector" json:"collector,omitempty"`
	// Conf is the mongoshake conf file path
	// +optional
	Conf string `yaml:"conf" json:"conf,omitempty"`
	// ExtraOptions contains additional raw mongoshake command-line flags
	// +optional
	ExtraOptions []string `yaml:"extraOptions" json:"extraOptions,omitempty"`
}
