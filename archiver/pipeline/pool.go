package pipeline

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// Pool is a worker pool.
type Pool struct {
	Parallelism uint
	Workers     []*Worker
	jobChannel  chan *Job
	jobQueue    chan chan *Job
	syncGroup   sync.WaitGroup
}

// NewPool creates a new pool of the specified number of workers.
func NewPool(parallelism uint) *Pool {
	return &Pool{
		Parallelism: parallelism,
		Workers:     make([]*Worker, parallelism),
		jobChannel:  make(chan *Job, 10),
		jobQueue:    make(chan chan *Job),
	}
}

// Submit submits a job to the pool.
// Note: will block once the queue buffer (10) is filled.
// TODO: Investigate if this causes issues in reality.
func (pool *Pool) Submit(job *Job) {
	log.Debug("Submitting job to pool")
	pool.syncGroup.Add(1)
	pool.jobChannel <- job
}

// Start starts the workers of the pool.
func (pool *Pool) Start() {
	log.Debugf("Starting pool with %d worker(s)", len(pool.Workers))
	for i := uint(0); i < pool.Parallelism; i++ {
		worker := NewWorker(pool.jobQueue, &pool.syncGroup)
		pool.Workers = append(pool.Workers, worker)
		worker.Start()
	}

	go pool.process()
}

// process listens for scheduled jobs and enlists a worker to handle them.
func (pool *Pool) process() {
	for {
		select {
		// Wait for a submitted job
		case job := <-pool.jobChannel:
			log.Debug("Got job, scheduling to worker")
			// Get an available worker channel from the queue
			queue := <-pool.jobQueue
			// Submit the job
			queue <- job
		}
	}
}

// Wait waits for all scheduled jobs to complete.
func (pool *Pool) Wait() {
	log.Debug("Waiting for all jobs to complete")
	pool.syncGroup.Wait()
}

// TODO: Implement quitting
