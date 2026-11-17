package enums

type NatsJobStatus string

const (
	NatsJobStatusQueued     NatsJobStatus = "queued"
	NatsJobStatusProcessing NatsJobStatus = "processing"
	NatsJobStatusFailed     NatsJobStatus = "failed"
	NatsJobStatusCompleted  NatsJobStatus = "completed"
)
