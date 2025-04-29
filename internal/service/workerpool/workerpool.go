package workerpool

import (
	"context"
	"time"
)

type Result struct {
	Err   error
	Value any
}

type Job struct {
	ctx     context.Context
	timeout time.Duration
	fn      func(ctx context.Context) (any, error)
	resCh   chan *Result
}

func NewJob(
	ctx context.Context,
	timeout time.Duration,
	fn func(ctx context.Context) (any, error),
) *Job {
	return &Job{
		ctx:     ctx,
		timeout: timeout,
		fn:      fn,
		resCh:   make(chan *Result),
	}
}

func (j *Job) GetResult() *Result {
	res := <-j.resCh
	return res
}

type Pool struct {
	jobs  chan *Job
	limit int
}

func NewPool(limit, queueSize int) *Pool {
	if queueSize == 0 {
		queueSize = limit * 5
	}

	return &Pool{
		jobs:  make(chan *Job, queueSize),
		limit: limit,
	}
}

func (p *Pool) Start() {
	for range p.limit {
		go p.worker()
	}
}

func (p *Pool) AddJob(job *Job) {
	p.jobs <- job
}

func (p *Pool) worker() {
	for job := range p.jobs {
		func() {
			ctx, cancel := context.WithTimeout(job.ctx, job.timeout)
			defer cancel()

			res, err := job.fn(ctx)
			r := &Result{
				Value: res,
				Err:   err,
			}

			job.resCh <- r
		}()
	}
}
