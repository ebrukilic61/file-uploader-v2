package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/queue"
)

type MediaService interface {
	RegisterMedia(req dto.MediaRegisterRequestDTO) (*dto.Media, error)
	ProcessResizeJobs(mediaID string, sizes []dto.Metadata) error
}

type mediaService struct {
	repo       repositories.MediaRepository
	workerPool *queue.WorkerPool
}
