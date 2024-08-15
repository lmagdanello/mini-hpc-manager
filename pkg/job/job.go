package job

type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
)

type Job struct {
	ID      string    `json:"id"`      // Job ID
	Name    string    `json:"name"`    // Job name
	Command string    `json:"command"` // Command to run in the container
	Status  JobStatus `json:"status"`  // Job status (pending, running, complete, failed)
	CPU     int       `json:"cpu"`     // CPU shares
	Memory  int       `json:"memory"`  // Memory limit (in bytes)
	Image   string    `json:"image"`   // Docker image
	Log     string    `json:"log"`     // Container logs
}
