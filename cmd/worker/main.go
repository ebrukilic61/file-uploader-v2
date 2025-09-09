package main //worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"file-uploader/internal/infrastructure/db"
	"file-uploader/internal/infrastructure/queue"
	infra_repo "file-uploader/internal/infrastructure/repositories"
	"file-uploader/internal/usecases"
	"file-uploader/pkg/config"
	"file-uploader/pkg/constants"
	fe "file-uploader/pkg/errors"
	fl "file-uploader/pkg/file"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.LoadConfig()
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	fmt.Println("Redis Host:", redisHost)
	fmt.Println("Redis Port:", redisPort)
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	db, err := db.NewPostgresDB()
	if err != nil {
		log.Fatal("DB bağlantısı kurulamadı:", err)
	}
	log.Println("DB bağlantısı başarılı!")

	fileRepo := infra_repo.NewFileUploadRepository(cfg.Upload.TempDir, cfg.Upload.UploadsDir, db)

	// cleanup içerisinde yazıldı cron job için
	cleanupUC := usecases.NewCleanupService(fileRepo)
	c := cron.New(cron.WithSeconds())

	c.AddFunc("0 0 * * * *", func() { // her saat başı çalışır
		log.Println("Running scheduled cleanup of old temp files...")
		if err := cleanupUC.CleanupOldTempFiles(2 * time.Hour); err != nil { // 2 saatten eski temp dosyaları siler
			log.Printf("Error cleaning up old temp files: %v", err)
		}
	})
	c.Start() // cron job'u başlatmak için
	// BRPOP loop to process jobs
	for {
		val, err := rdb.BRPop(ctx, 0, "job_queue").Result()
		if err != nil {
			log.Println("BRPop failed:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		job, err := queue.DeserializeJob(val[1])
		if err != nil {
			log.Println("DeserializeJob failed:", err)
			continue
		}

		switch job.Type {
		case queue.JobSaveChunk:
			processChunk(job, fileRepo)
		case queue.JobMerge:
			processMerge(job, fileRepo, rdb, ctx)
		case queue.JobRetry: //* process retry job'a düşünce burası işlenecek
			processRetryMerge(job, fileRepo, rdb, ctx)
		case queue.JobCleanup:
			processCleanup(job, fileRepo)
		default:
			log.Println("Unknown job type:", job.Type)
		}
	}
}

func processChunk(job *queue.Job, repo *infra_repo.FileUploadRepository) {
	log.Printf("Processing chunk %d for file %s (UploadID: %s)", job.ChunkIndex, job.Filename, job.UploadID)
	if exists := repo.ChunkExists(job.UploadID, job.Filename, job.ChunkIndex); exists {
		log.Printf("Chunk %d for file %s already exists, skipping", job.ChunkIndex, job.Filename)
		return
	}

	// Diske kaydetmek için:
	if err := repo.SaveChunkBytes(job.UploadID, job.Filename, job.ChunkIndex, job.FileContent); err != nil {
		log.Printf("Failed to save chunk %d: %v", job.ChunkIndex, err)
		return
	}

	// Chunk path:
	chunkPath := filepath.Join("temp_uploads", job.UploadID, fmt.Sprintf("%s.part%d", job.Filename, job.ChunkIndex))

	// Hash doğrulama:
	if job.ChunkHash != "" {
		if err := fl.ValidateFileHash(chunkPath, job.ChunkHash); err != nil {
			log.Printf("Hash validation failed for chunk %d: %v", job.ChunkIndex, err)
			repo.CleanupTempFiles(job.UploadID)
			return
		}
	}

	// Repo’yu güncellemek için:
	if err := repo.SetUploadedChunks(job.UploadID, job.Filename, job.ChunkIndex); err != nil {
		log.Printf("Failed to update uploaded chunks for %s: %v", job.Filename, err)
		return
	}
	log.Printf("Chunk %d for file %s saved successfully", job.ChunkIndex, job.Filename)
}

func processMerge(job *queue.Job, repo *infra_repo.FileUploadRepository, rdb *redis.Client, ctx context.Context) {
	//* exponential backoff ile merge işlemi gerçekleştirildi
	const maxRetries = 5
	retryDelay := 1 * time.Second

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	log.Printf("Dosya %s için merge işlemi başlıyor (UploadID: %s, TotalChunks: %d)",
		job.Filename, job.UploadID, job.TotalChunks)

	var mergedFilePath string
	var err error

	for i := 0; i < maxRetries; i++ {
		mergedFilePath, err = repo.MergeChunks(job.UploadID, job.Filename, job.TotalChunks)
		backoff := time.Duration(retryDelay << i) // Exponential backoff
		if err != nil {
			var uploadErr *fe.UploadError
			if errors.As(err, &uploadErr) && uploadErr.Code == "missing_chunk" {
				job.LastError = err.Error()
				log.Printf("Eksik chunk hatası, %d/%d tekrar %v saniye sonra...", i+1, maxRetries, backoff)
				time.Sleep(backoff)
				continue
			} else {
				log.Printf("Merge işlemi yapılamadı %s: %v", job.Filename, err)
				return
			}
		} else {
			log.Printf("Merge işlemi başarıyla gerçekleşti %s: %s", job.Filename, mergedFilePath)
			break
		}
	}

	if err != nil {
		log.Printf("Merge işlemi başarısız oldu %s: %v", job.Filename, err)
		job.LastError = err.Error()
		payload := []byte{}
		if p, marshalErr := json.Marshal(job); marshalErr == nil {
			payload = p
		} else {
			log.Printf("failed merge işlemi için job verisi oluşturulamadı: %v", marshalErr)
		}
		if saveErr := repo.SaveFailedUpload(job.UploadID, job.Filename, string(job.Type), err.Error(), payload); saveErr != nil {
			log.Printf("failed upload kaydı yapılamadı: %v", saveErr)
		}

		if job.RetryCount < constants.MaxRetryJobs {
			retryJob := queue.Job{
				UploadID:   job.UploadID,
				Filename:   job.Filename,
				Type:       queue.JobRetry,
				RetryCount: job.RetryCount + 1,
			}
			retryPayload, _ := json.Marshal(retryJob)
			err := rdb.LPush(ctx, "job_queue", retryPayload)
			if err != nil {
				log.Printf("retry job queue'ya eklenemedi: %v / RetryCount: %d", err, retryJob.RetryCount)
			} else {
				log.Printf("Otomatik retry yapılarak job queue'ya eklendi: %s (RetryCount: %d tamamlandı)", job.Filename, retryJob.RetryCount)
			}
		} else {
			log.Printf("Max retry sayısına ulaşıldı, retry yapılmayacak: %s", job.Filename)
		}
		return
	}

	// Push to processed queue for callback
	processed := queue.ProcessedJob{
		UploadID:       job.UploadID,
		Filename:       job.Filename,
		MergedFilePath: mergedFilePath,
		TotalChunks:    job.TotalChunks,
	}
	serialized, err := json.Marshal(processed)
	if err != nil {
		log.Printf("Failed to serialize processed job %s: %v", job.Filename, err)
		return
	}
	rdb.LPush(ctx, "processed_queue", serialized)
	log.Printf("Processed job pushed to processed_queue: %s", job.Filename)
}

func processRetryMerge(job *queue.Job, repo *infra_repo.FileUploadRepository, rdb *redis.Client, ctx context.Context) {
	log.Printf("Processing retry merge for file %s (UploadID: %s)", job.Filename, job.UploadID)
	finalPath, totalChunks, err := repo.RetryMerge(job.UploadID, job.Filename)
	if err != nil {
		log.Printf("Retry merge failed for %s: %v", job.Filename, err)
		return
	}
	log.Printf("Retry merge succeeded for %s: %s", job.Filename, finalPath)

	// Push to processed queue for callback
	processed := queue.ProcessedJob{
		UploadID:       job.UploadID,
		Filename:       job.Filename,
		MergedFilePath: finalPath,
		TotalChunks:    totalChunks,
	}
	serialized, err := json.Marshal(processed)
	if err != nil {
		log.Printf("Failed to serialize processed job %s: %v", job.Filename, err)
		return
	}
	rdb.LPush(ctx, "processed_queue", serialized)
	log.Printf("Processed job pushed to processed_queue: %s", job.Filename)
}

func processCleanup(job *queue.Job, repo *infra_repo.FileUploadRepository) {
	log.Printf("Processing cleanup for UploadID: %s", job.UploadID)
	repo.CleanupTempFiles(job.UploadID)
	log.Printf("Cleanup completed for UploadID: %s", job.UploadID)
}
