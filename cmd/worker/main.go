package main //worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"file-uploader/internal/infrastructure/queue"
	infra_repo "file-uploader/internal/infrastructure/repositories"
	"file-uploader/pkg/config"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
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

	fileRepo := infra_repo.NewFileUploadRepository(cfg.Upload.TempDir, cfg.Upload.UploadsDir)

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
		case queue.JobCleanup:
			processCleanup(job, fileRepo)
		default:
			log.Println("Unknown job type:", job.Type)
		}
	}
}

func processChunk(job *queue.Job, repo *infra_repo.FileUploadRepository) {
	log.Printf("Processing chunk %d for file %s (UploadID: %s)", job.ChunkIndex, job.Filename, job.UploadID)

	// If chunk processing is needed, do it here
	// For now, we'll just log that the chunk was processed
	// The actual chunk file should already be saved to disk from UploadChunk

	log.Printf("Chunk %d processed successfully for %s", job.ChunkIndex, job.Filename)
}

func processMerge(job *queue.Job, repo *infra_repo.FileUploadRepository, rdb *redis.Client, ctx context.Context) {
	const maxRetries = 5
	const retryDelay = 2 * time.Second

	log.Printf("Starting merge process for %s (UploadID: %s, TotalChunks: %d)",
		job.Filename, job.UploadID, job.TotalChunks)

	var mergedFilePath string
	var err error

	for i := 0; i < maxRetries; i++ {
		mergedFilePath, err = repo.MergeChunks(job.UploadID, job.Filename, job.TotalChunks)
		if err != nil && strings.Contains(err.Error(), "eksik chunk") {
			log.Printf("Missing chunks, retry %d/%d after %v...", i+1, maxRetries, retryDelay)
			time.Sleep(retryDelay)
			continue
		} else if err != nil {
			log.Printf("Merge failed for %s: %v", job.Filename, err)
			return
		} else {
			log.Printf("Merge successful for %s: %s", job.Filename, mergedFilePath)
			break
		}
	}

	if err != nil {
		log.Printf("Merge failed after %d retries for %s: %v", maxRetries, job.Filename, err)
		return
	}

	// Push to processed queue for callback
	processed := queue.ProcessedJob{
		UploadID:       job.UploadID,
		Filename:       job.Filename,
		MergedFilePath: mergedFilePath,
	}
	serialized, _ := json.Marshal(processed)
	rdb.LPush(ctx, "processed_queue", serialized)
	log.Printf("Processed job pushed to processed_queue: %s", job.Filename)
}

func processCleanup(job *queue.Job, repo *infra_repo.FileUploadRepository) {
	log.Printf("Processing cleanup for UploadID: %s", job.UploadID)
	repo.CleanupTempFiles(job.UploadID)
	log.Printf("Cleanup completed for UploadID: %s", job.UploadID)
}
