package types

// CollectionInfo represents basic information about a collection
type CollectionInfo struct {
	Name string `json:"name"`
}

// CollectionsResult contains the list of collections
type CollectionsResult struct {
	Collections []CollectionInfo `json:"collections"`
}

// UsageHardware represents hardware usage information
type UsageHardware struct {
	CPU                 int `json:"cpu"`
	PayloadIORead       int `json:"payload_io_read"`
	PayloadIOWrite      int `json:"payload_io_write"`
	PayloadIndexIORead  int `json:"payload_index_io_read"`
	PayloadIndexIOWrite int `json:"payload_index_io_write"`
	VectorIORead        int `json:"vector_io_read"`
	VectorIOWrite       int `json:"vector_io_write"`
}

// UsageInference represents inference usage information
type UsageInference struct {
	Models map[string]interface{} `json:"models"`
}

// Usage represents the usage information in the response
type Usage struct {
	Hardware  UsageHardware  `json:"hardware"`
	Inference UsageInference `json:"inference"`
}

// ListCollectionsResponse represents the response from listing all collections
type ListCollectionsResponse struct {
	Usage  *Usage             `json:"usage"`
	Time   float64            `json:"time"`
	Status string             `json:"status"`
	Result *CollectionsResult `json:"result"`
}

// CollectionInfoResponse represents the response from getting collection info
type CollectionInfoResponse struct {
	Usage  *Usage             `json:"usage"`
	Time   float64            `json:"time"`
	Status string             `json:"status"`
	Result *CollectionDetails `json:"result"`
}

// CollectionDetails contains detailed information about a collection
type CollectionDetails struct {
	Name           string                 `json:"name"`
	VectorsCount   uint64                 `json:"vectors_count"`
	PointsCount    uint64                 `json:"points_count"`
	PayloadSchema  map[string]interface{} `json:"payload_schema"`
	Status         string                 `json:"status"`
	Conditions     string                 `json:"conditions,omitempty"`
	OptimizeHidden *bool                  `json:"optimize_hidden,omitempty"`
	AutoMigrate    *bool                  `json:"auto_migrate,omitempty"`
	RAMUsage       uint64                 `json:"ram_usage,omitempty"`
	DiskUsage      uint64                 `json:"disk_usage,omitempty"`
}
