package routers

import (
	"file-uploader/internal/delivery/http/handlers"
	"file-uploader/internal/usecases"

	"github.com/gofiber/fiber/v2"
)

func SetupUploadRoutes(app *fiber.App, uploadService usecases.UploadService) {

	uploadHandler := handlers.NewUploadHandler(uploadService)

	// Routes:
	api := app.Group("/api/v1")
	api.Post("/upload/chunk", uploadHandler.UploadChunk)
	api.Post("/upload/complete", uploadHandler.CompleteUpload)
	api.Post("/upload/cancel", uploadHandler.CancelUpload)
	api.Get("/upload/status", uploadHandler.UploadStatus)
}
