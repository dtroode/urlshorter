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
	Ctx     context.Context
	Timeout time.Duration
	Fn      func(ctx context.Context) (any, error)
	ResCh   chan *Result
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

func (p *Pool) Close() {
	close(p.jobs)
}

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
