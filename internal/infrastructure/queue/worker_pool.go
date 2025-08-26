package queue

import (
	"context"
	"file-uploader/internal/domain/repositories"
	"sync"
)

type WorkerPool struct {
	JobChan chan Job
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWorkerPool(workerCount int, repo repositories.FileUploadRepository) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		JobChan: make(chan Job, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			ID:      i,
			JobChan: pool.JobChan,
			Wg:      &pool.wg,
			Repo:    repo,
		}
		pool.wg.Add(1)
		worker.Start(pool.ctx)
	}
	return pool
}

/*
func NewMediaWorkerPool(workerCount int, repo repositories.MediaRepository) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		JobChan: make(chan Job, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			ID:      i,
			JobChan: pool.JobChan,
			Wg:      &pool.wg,
			Repo:    repo,
		}
		pool.wg.Add(1)
		worker.Start(pool.ctx)
	}
	return pool
}
*/

func (p *WorkerPool) AddJob(job Job) {
	p.JobChan <- job
}

func (p *WorkerPool) Shutdown() {
	p.cancel()
	close(p.JobChan)
	p.wg.Wait()
}
