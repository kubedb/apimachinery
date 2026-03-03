package types

type Snapshot struct {
	Name         string `json:"name"`
	Size         uint64 `json:"size"`
	CreationTime string `json:"creation_time"`
	Checksum     string `json:"checksum"`
}

type ListSnapshotsResponse struct {
	Time   float64    `json:"time"`
	Status string     `json:"status"`
	Result []Snapshot `json:"result"`
}
type CreateSnapshotResponse struct {
	Time   float64  `json:"time"`
	Status string   `json:"status"`
	Result Snapshot `json:"result"`
}
type DeleteSnapshotResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
}

type RecoverSnapshotResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
}
