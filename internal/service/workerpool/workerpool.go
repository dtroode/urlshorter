package workerpool

import (
	"context"
	"time"
)

// Result represents the result of a job execution.
type Result struct {
	Err   error
	Value any
}

// Job represents a job to be executed by the worker pool.
type Job struct {
	Ctx     context.Context
	Timeout time.Duration
	Fn      func(ctx context.Context) (any, error)
	ResCh   chan *Result
}

// Pool represents a worker pool for executing jobs.
type Pool struct {
	jobs  chan *Job
	limit int
}

// NewPool creates new Pool instance.
func NewPool(limit, queueSize int) *Pool {
	if queueSize == 0 {
		queueSize = limit * 5
	}

	return &Pool{
		jobs:  make(chan *Job, queueSize),
		limit: limit,
	}
}

// Start starts the worker pool with specified number of workers.
func (p *Pool) Start() {
	for range p.limit {
		go p.worker()
	}
}

// Close closes the worker pool and stops accepting new jobs.
func (p *Pool) Close() {
	close(p.jobs)
}

// Submit submits a new job to the worker pool.
func (p *Pool) Submit(
	ctx context.Context,
	timeout time.Duration,
	fn func(ctx context.Context) (any, error),
	expectResult bool,
) *Job {
	var resCh chan *Result
	if expectResult {
		resCh = make(chan *Result, 1)
	}

	job := &Job{
		Ctx:     ctx,
		Timeout: timeout,
		Fn:      fn,
		ResCh:   resCh,
	}

	p.jobs <- job
	return job
}

// worker represents a worker goroutine that processes jobs.
func (p *Pool) worker() {
	for job := range p.jobs {
		func() {
			ctx, cancel := context.WithTimeout(job.Ctx, job.Timeout)
			defer cancel()

			res, err := job.Fn(ctx)

			if job.ResCh != nil {
				r := &Result{
					Value: res,
					Err:   err,
				}

				job.ResCh <- r
			}
		}()
	}
}
