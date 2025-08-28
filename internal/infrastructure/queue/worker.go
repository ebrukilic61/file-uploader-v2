package queue

import (
	"context"
	"file-uploader/internal/domain/repositories"
	"file-uploader/pkg/errors"
	"fmt"
	"log"
	"sync"
)

type Worker struct {
	ID      int        // worker id
	JobChan <-chan Job // iş kuyruğu
	Wg      *sync.WaitGroup
	Repo    repositories.FileUploadRepository
}

type WorkerMedia struct {
	ID      int        // worker id
	JobChan <-chan Job // iş kuyruğu
	Wg      *sync.WaitGroup
	Repo    repositories.MediaRepository
}

func (w *WorkerMedia) Start(ctx context.Context) { // worker başlatma fonksiyonu
	go func() {
		defer w.Wg.Done()
		for {
			select {
			case job, ok := <-w.JobChan: //channeldan iş alınır
				if !ok {
					log.Printf("Worker %d: Job channel closed", w.ID)
					return
				}
				select {
				case <-ctx.Done():
					log.Printf("Worker %d: job %d cancelled", w.ID, job.UploadID)
					continue
				default:
					//w.processJobMedia(job)
				}
			case <-ctx.Done():
				log.Printf("Worker %d: Stopping due to context cancellation", w.ID)
				return
			}
		}
	}()
}

func processJobMedia(job Job) {
	log.Printf("Processing media job: %s", job.UploadID)
	// Media job processing logic goes here
}

func (w *Worker) Start(ctx context.Context) { // worker başlatma fonksiyonu
	go func() {
		defer w.Wg.Done()
		for {
			select {
			case job, ok := <-w.JobChan: //channeldan iş alınır
				if !ok {
					log.Printf("Worker %d: Job channel closed", w.ID)
					return
				}
				select {
				case <-ctx.Done():
					log.Printf("Worker %d: job %d cancelled", w.ID, job.UploadID)
					continue
				default:
					w.processJob(job)
				}
			case <-ctx.Done():
				log.Printf("Worker %d: Stopping due to context cancellation", w.ID)
				return
			}
		}
	}()
}

func (w *Worker) processJob(job Job) {
	log.Printf("Worker %d: Processing job %s for upload %s", w.ID, job.Type, job.UploadID)

	var err error

	switch job.Type {
	case JobSaveChunk:
		err = w.processSaveChunk(job)
	case JobMerge:
		err = w.processMergeChunks(job)
	case JobCleanup:
		err = w.processCleanup(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	if err != nil {
		log.Printf("Worker %d: Job %s failed: %v", w.ID, job.Type, err)
	} else {
		log.Printf("Worker %d: Job %s succeeded", w.ID, job.Type)
	}
}

func (w *Worker) processSaveChunk(job Job) error {
	if job.File == nil {
		return fmt.Errorf("file is nil")
	}
	defer job.File.Close()

	if err := w.Repo.SaveChunk(job.UploadID, job.Filename, job.ChunkIndex, job.File); err != nil {
		errors.ErrChunkNotSave(err)
		w.Repo.CleanupTempFiles(job.UploadID) // Hata durumunda temp dosyaları temizle
		return err
	}
	return nil
}

func (w *Worker) processMergeChunks(job Job) error {
	err := w.Repo.MergeChunks(job.UploadID, job.Filename, job.TotalChunks)
	if err != nil {
		errors.ErrChunksNotMerged(err)
		w.Repo.CleanupTempFiles(job.UploadID) // Hata durumunda temp dosyaları temizle
		return err
	}
	if err := w.Repo.CleanupTempFiles(job.UploadID); err != nil {
		//log.Printf("failed to cleanup temp files: %w", err)
		errors.ErrCannotRemove(err)
		return err
	}
	return nil
}

func (w *Worker) processCleanup(job Job) error {
	return w.Repo.CleanupTempFiles(job.UploadID)
}
