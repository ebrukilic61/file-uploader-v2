package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "file-uploader/docs"

	handlers "file-uploader/internal/delivery/http"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/pkg/config"
	"file-uploader/internal/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // fiber için swagger handler
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.LoadConfig()

	app := fiber.New(fiber.Config{
		BodyLimit: int(cfg.Upload.MaxFileSize),
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Routes
	setupUploadRoutes(app, cfg)

	// server ayakta mı değil mi kontrolü:
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Temp directory: %s", cfg.Upload.TempDir)
	log.Printf("Uploads directory: %s", cfg.Upload.UploadsDir)

	//graceful shutdown:
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server başlatılamadı: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown sinyali alındı, server kapatılıyor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server düzgün kapatılamadı: %v", err)
	}
	log.Println("Server düzgün bir şekilde kapatıldı")
}

func setupUploadRoutes(app *fiber.App, cfg *config.Config) {
	fileRepo := repositories.NewFileUploadRepository(cfg.Upload.TempDir, cfg.Upload.UploadsDir)
	//localStorage := &storage.LocalStorage{BasePath: "uploads"}
	cleanupUC := usecases.NewCleanupService(fileRepo)
	c := cron.New(cron.WithSeconds())

	c.AddFunc("0 */5 * * * *", func() {
		if err := cleanupUC.CleanupOldTempFiles(24 * time.Hour); err != nil {
			log.Printf("Error cleaning up old temp files: %v", err)
		}
	})
	c.Start() // cron job'u başlatır

	//uploadService := usecases.NewUploadService(fileRepo, localStorage)
	uploadService := usecases.NewUploadService(fileRepo, nil)

	uploadHandler := handlers.NewUploadHandler(uploadService)

	// Routes:
	api := app.Group("/api/v1")
	api.Post("/upload/chunk", uploadHandler.UploadChunk)
	api.Post("/upload/complete", uploadHandler.CompleteUpload)
	api.Post("/upload/cancel", uploadHandler.CancelUpload)
	api.Get("/upload/status", uploadHandler.UploadStatus)
}
