package pipeline

import (
	"fmt"

	"github.com/AlexGustafsson/larch/formats/warc"
)

// Status is the status of a job.
type Status int

const (
	// Idle is the status of an idle job
	Idle Status = iota
	// Scheduled is the status of a job that has been scheduled
	Scheduled = iota
	// Active is the status of a job that is currently being processed
	Active = iota
	// Completed is the status of a job that completed successfully
	Completed = iota
	// Failed is the status of a job that completed unsuccessfully
	Failed = iota
)

// JobCompletedCallback is called with the resulting record whenever a job is completed.
type JobCompletedCallback func(job *Job, records ...*warc.Record)

// JobFailedCallback is called with the error whenver a job fails.
type JobFailedCallback func(job *Job, err error)

// PerformJobCallback is called whenever the job is performed.
type PerformJobCallback func(job *Job)

// JobHandler is the function invoked to perform the job. May return nil or one or more records.
type JobHandler func(job *Job) ([]*warc.Record, error)

// Job is a job to produce a WARC record.
type Job struct {
	// Name is a human-readable name of the job.
	Name string
	// Description is a human-readable description for the job.
	Description string
	// Tries is the current number of performed tries
	Tries uint
	// handler is the function invoked to perform the job.
	handler JobHandler
	// MaximumTries is the maximum number of tries before ignoring the job.
	MaximumTries uint
	// Status is the current status of the job.
	Status Status
	// JobCompletedCallback is called with the resulting record whenever a job is completed.
	JobCompletedCallback JobCompletedCallback
	// JobFailedCallback is called with the error whenver a job fails.
	JobFailedCallback JobFailedCallback
	// PerformJobCallback is called whenever the job is performed.
	PerformJobCallback PerformJobCallback
}

// NewJob creates a new job with default settings.
func NewJob(name string, description string, handler JobHandler) *Job {
	return &Job{
		Name:         name,
		Description:  description,
		Tries:        0,
		handler:      handler,
		MaximumTries: 1,
		Status:       Idle,
	}
}

// Perform invokes the job.
func (job *Job) Perform() {
	job.Status = Active

	if job.PerformJobCallback != nil {
		job.PerformJobCallback(job)
	}

	if job.handler == nil {
		job.Fail(fmt.Errorf("No handler available for job"))
	}

	record, err := job.handler(job)
	if err != nil {
		job.Fail(err)
	}

	job.Complete(record...)
}

// Complete completes the job with the given record.
func (job *Job) Complete(records ...*warc.Record) {
	job.Tries++
	job.Status = Completed
	if job.JobCompletedCallback != nil {
		// TODO: Do we need a nil check for records here?
		job.JobCompletedCallback(job, records...)
	}
}

// Fail fails the job with the given error.
func (job *Job) Fail(err error) {
	job.Tries++
	job.Status = Failed
	if job.JobFailedCallback != nil {
		job.JobFailedCallback(job, err)
	}
}
