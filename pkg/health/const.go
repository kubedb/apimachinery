package health

type HealthCheckFailureLabel string

const (
	HealthCheckClientFailure HealthCheckFailureLabel = "ClientFailure"
	HealthCheckPingFailure   HealthCheckFailureLabel = "PingFailure"
	HealthCheckWriteFailure  HealthCheckFailureLabel = "WriteFailure"
)
