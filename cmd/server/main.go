package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "file-uploader/docs"

	"file-uploader/pkg/config"
	consts "file-uploader/pkg/constants"

	"file-uploader/internal/delivery/http/routers"
	"file-uploader/internal/infrastructure/db"
	"file-uploader/internal/infrastructure/queue"
	infra_repo "file-uploader/internal/infrastructure/repositories"
	"file-uploader/internal/infrastructure/storage"
	"file-uploader/internal/usecases"

	_ "file-uploader/migrations"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

func main() {
	cfg := config.LoadConfig()
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("DB bağlantısı başarısız: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("sql.DB alınamadı: %v", err)
	}

	if os.Getenv("RUN_AUTO_MIGRATION") == "true" {
		if err := goose.Up(sqlDB, "."); err != nil {
			log.Fatalf("failed to apply migrations: %v", err)
		}
	}

	config.EnsureDirs()
	goose.SetBaseFS(nil)

	app := fiber.New(fiber.Config{
		BodyLimit: int(cfg.Upload.MaxFileSize),
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Swagger UI
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Repositories & Services
	fileRepo := infra_repo.NewFileUploadRepository(cfg.Upload.TempDir, cfg.Upload.UploadsDir)
	localStorage := storage.NewLocalStorage(cfg.Upload.UploadsDir)
	mediaRepo := infra_repo.NewMediaRepository(database)
	variantRepo := infra_repo.NewMediaVariantRepository(database)
	sizeRepo := infra_repo.NewMediaSizeRepository(database)
	videoRepo := infra_repo.NewVideoRepository(database)
	mediaService := usecases.NewMediaService(mediaRepo, variantRepo, sizeRepo, localStorage, videoRepo)

	uploadService := usecases.NewUploadService(fileRepo, localStorage, rdb, mediaService)

	// Routes
	routers.SetupUploadRoutes(app, uploadService)
	routers.SetupMediaRoutes(app, cfg, database)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": consts.StatusOK})
	})

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	// Processed queue listener (callback tetikleyici)
	go startProcessedQueueListener(rdb, uploadService)

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server başlatılamadı: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown sinyali alındı, server kapatılıyor...")

	ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctxShut); err != nil {
		log.Fatalf("Server düzgün kapatılamadı: %v", err)
	}
	log.Println("Server düzgün bir şekilde kapatıldı")
}

// Processed queue listener
func startProcessedQueueListener(rdb *redis.Client, uploadService usecases.UploadService) {
	ctx := context.Background()
	for {
		val, err := rdb.BRPop(ctx, 0, "processed_queue").Result()
		if err != nil {
			log.Println("BRPop failed:", err)
			time.Sleep(time.Second)
			continue
		}

		var processed queue.ProcessedJob
		if err := json.Unmarshal([]byte(val[1]), &processed); err != nil {
			log.Println("Deserialize processed job failed:", err)
			continue
		}

		if err := uploadService.HandleMergeSuccess(processed.UploadID, processed.Filename, processed.MergedFilePath, processed.TotalChunks); err != nil {
			log.Printf("HandleMergeSuccess error: %v", err)
			fmt.Printf("Total chunk sayısı: %d\n", processed.TotalChunks)
		} else {
			log.Printf("HandleMergeSuccess executed: %s", processed.Filename)
			fmt.Printf("Total chunk sayısı: %d\n", processed.TotalChunks)
		}
	}
}
