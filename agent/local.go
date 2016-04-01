package agent

import (
	"github.com/sofuture/kubernotes/cluster"
)

// Represents and interface to manage running processes or jobs.
type Local interface {

	// Establish connection with job management backend.
	Connect() error

	// Disconnect from the job management backend.
	Disconnect()

	// Retrieve the list of jobs that we're managing.
	GetManagedJobs() ([]cluster.Job, error)

	// Create a job.
	CreateJob(job *cluster.Job) error

	// Start an existing job.
	StartJob(job *cluster.Job) error

	// Stop a running job.
	StopJob(job *cluster.Job) error

	// Destroy an existing job.
	DestroyJob(job *cluster.Job) error

	// GetLogs
	GetLogs(job *cluster.Job, count int) (string, error)
}
