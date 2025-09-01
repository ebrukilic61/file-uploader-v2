package routers

import (
	"file-uploader/internal/delivery/http/handlers"
	infra_repo "file-uploader/internal/infrastructure/repositories"
	"file-uploader/internal/infrastructure/storage"
	"file-uploader/internal/usecases"
	"file-uploader/pkg/config"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupMediaRoutes(app *fiber.App, cfg *config.Config, database *gorm.DB) {
	mediaRepo := infra_repo.NewMediaRepository(database)
	variantRepo := infra_repo.NewMediaVariantRepository(database)
	sizeRepo := infra_repo.NewMediaSizeRepository(database)

	// Storage
	localStorage := storage.NewLocalStorage(cfg.Upload.UploadsDir)

	// Service
	mediaService := usecases.NewMediaService(mediaRepo, variantRepo, sizeRepo, localStorage)
	mediaHandler := handlers.NewMediaHandler(mediaService)

	api := app.Group("/api/v1")
	api.Get("/media/:id", mediaHandler.GetMedia)
	api.Post("/media", mediaHandler.CreateMedia)
	api.Post("/media/:id/variants", mediaHandler.CreateVariantsForMedia)
	api.Post("/media/size", mediaHandler.CreateSize)
	api.Put("/media/size", mediaHandler.UpdateSize)
}
