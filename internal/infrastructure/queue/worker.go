package queue

import (
	"context"
	"encoding/json"
	"file-uploader/internal/domain/repositories"
	"fmt"
	"log"
	"os"
	"sync"
)

type Worker struct {
	ID            int        // worker id
	JobChan       <-chan Job // iş kuyruğu
	Wg            *sync.WaitGroup
	Repo          repositories.FileUploadRepository
	MergeCallback func(uploadID, filename, mergedFilePath string)
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
	if job.FilePath == "" {
		return fmt.Errorf("file path is empty")
	}

	file, err := os.Open(job.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer file.Close()

	if err := w.Repo.SaveChunk(job.UploadID, job.Filename, job.ChunkIndex, file); err != nil {
		w.Repo.CleanupTempFiles(job.UploadID) // Hata durumunda temp dosyaları temizler
		return err
	}
	return nil
}

func (w *Worker) processMergeChunks(job Job) error {
	log.Printf("Worker %d: Merge işlemi başlatılıyor - UploadID: %s, Filename: %s", w.ID, job.UploadID, job.Filename)

	mergedFilePath, err := w.Repo.MergeChunks(job.UploadID, job.Filename, job.TotalChunks)
	if err != nil {
		log.Printf("Worker %d: Merge işlemi başarısız - %v", w.ID, err)
		return err
	}

	log.Printf("Worker %d: Merge işlemi tamamlandı - %s", w.ID, mergedFilePath)

	// Callback varsa merge sonrası çağır
	if w.MergeCallback != nil {
		w.MergeCallback(job.UploadID, job.Filename, mergedFilePath)
	}

	return nil
}

func (w *Worker) processCleanup(job Job) error {
	return w.Repo.CleanupTempFiles(job.UploadID)
}

func DeserializeJob(data string) (*Job, error) {
	var job Job
	if err := json.Unmarshal([]byte(data), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}
	return &job, nil
}

func SerializeJob(job Job) (string, error) {
	bytes, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to serialize job: %w", err)
	}
	return string(bytes), nil
}
