package grpc

type SnapShot struct {
	Data  []byte `json:"data"`
	Token string `json:"token"`
}
