package qdrant

// PointId is either a uint64 or a UUID string.
type PointId = interface{}

func PointIdFromNum(n uint64) PointId  { return n }
func PointIdFromUUID(u string) PointId { return u }

type PointStruct struct {
	Id      PointId                `json:"id"`
	Vector  interface{}            `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type UpsertPointsRequest struct {
	CollectionName string        `json:"-"`
	Points         []PointStruct `json:"points"`
}

type UpsertPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result Result  `json:"result"`
}

type GetPointsRequest struct {
	CollectionName string    `json:"-"`
	Ids            []PointId `json:"ids"`
}

type GetPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result []Point `json:"result"`
}

type Point struct {
	Id      PointId                `json:"id"`
	Vector  interface{}            `json:"vector,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}
