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

	"file-uploader/pkg/config"

	"file-uploader/internal/delivery/http/routers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger" // fiber için swagger handler
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
	routers.SetupUploadRoutes(app, cfg)

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
