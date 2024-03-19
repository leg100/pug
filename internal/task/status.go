package task

import "time"

// Status is a stage in the lifecycle of a task.
type Status string

const (
	Pending  Status = "pending"
	Queued   Status = "queued"
	Running  Status = "running"
	Exited   Status = "exited"
	Errored  Status = "errored"
	Canceled Status = "canceled"
)

type statusTimestamps struct {
	started time.Time
	ended   time.Time
}

func (sd statusTimestamps) Elapsed() time.Duration {
	if sd.started.IsZero() {
		return 0
	}
	if sd.ended.IsZero() {
		return time.Since(sd.started)
	}
	return sd.ended.Sub(sd.started)
}
