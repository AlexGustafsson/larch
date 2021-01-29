package pipeline

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// Worker is a routine processing jobs as quickly as it can.
type Worker struct {
	jobChannel chan *Job
	queue      chan chan *Job
	quit       chan struct{}
	syncGroup  *sync.WaitGroup
}

// NewWorker creates a new worker for the specified job queue.
func NewWorker(queue chan chan *Job, syncGroup *sync.WaitGroup) *Worker {
	return &Worker{
		jobChannel: make(chan *Job),
		queue:      queue,
		quit:       make(chan struct{}),
		syncGroup:  syncGroup,
	}
}

// Start starts the worker
func (worker *Worker) Start() {
	log.Debug("Starting worker")
	go func() {
		for {
			worker.queue <- worker.jobChannel
			select {
			case job := <-worker.jobChannel:
				log.Debug("Got job")
				job.Perform()
				worker.syncGroup.Done()
			case <-worker.quit:
				log.Debug("Quitting worker")
				close(worker.jobChannel)
				return
			}
		}
	}()
}
