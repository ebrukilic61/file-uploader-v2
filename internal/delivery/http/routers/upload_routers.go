package routers

import (
	"file-uploader/internal/delivery/http/handlers"
	infra_repo "file-uploader/internal/infrastructure/repositories"
	"file-uploader/internal/infrastructure/storage"
	"file-uploader/internal/usecases"
	"file-uploader/pkg/config"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func SetupUploadRoutes(app *fiber.App, cfg *config.Config, database *gorm.DB) {
	fileRepo := infra_repo.NewFileUploadRepository(cfg.Upload.TempDir, cfg.Upload.UploadsDir)
	localStorage := storage.NewLocalStorage(cfg.Upload.UploadsDir) // Genel dosya yüklemeleri
	mediaRepo := infra_repo.NewMediaRepository(database)
	variantRepo := infra_repo.NewMediaVariantRepository(database)
	sizeRepo := infra_repo.NewMediaSizeRepository(database)

	mediaService := usecases.NewMediaService(mediaRepo, variantRepo, sizeRepo, localStorage)

	cleanupUC := usecases.NewCleanupService(fileRepo)
	c := cron.New(cron.WithSeconds())

	c.AddFunc("0 */5 * * * *", func() {
		if err := cleanupUC.CleanupOldTempFiles(24 * time.Hour); err != nil {
			log.Printf("Error cleaning up old temp files: %v", err)
		}
	})
	c.Start() // cron job'u başlatır

	uploadService := usecases.NewUploadService(fileRepo, localStorage, mediaService)

	uploadHandler := handlers.NewUploadHandler(uploadService)

	// Routes:
	api := app.Group("/api/v1")
	api.Post("/upload/chunk", uploadHandler.UploadChunk)
	api.Post("/upload/complete", uploadHandler.CompleteUpload)
	api.Post("/upload/cancel", uploadHandler.CancelUpload)
	api.Get("/upload/status", uploadHandler.UploadStatus)
}
