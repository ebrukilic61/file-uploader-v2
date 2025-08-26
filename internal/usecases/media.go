package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type MediaService interface {
	RegisterMedia(filename, fileType string, fileSize int64, filePath string) (*dto.Media, error)
	GetMedia(mediaID string) (*dto.Media, error)
	ListMedia() ([]dto.Media, error)
	DeleteMedia(mediaID string) error
	UpdateMediaMetadata(mediaID string, metadata dto.Metadata) error
}

type mediaService struct {
	repo repositories.MediaRepository
}

func (s *mediaService) RegisterMedia(filename, fileType string, fileSize int64, filePath string) (*dto.Media, error) {
	media := &dto.Media{
		MediaID:   uuid.New().String(),
		Filename:  filename,
		FileType:  fileType,
		FileSize:  fileSize,
		FilePath:  filePath,
		Metadata:  dto.Metadata{},
		CreatedAt: time.Now(),
	}

	s.repo.RegisterMedia(media) //repoya ekleniyor
	return media, nil
}

func (s *mediaService) GetMedia(mediaID string) (*dto.Media, error) {
	return s.repo.GetByID(mediaID)
}

func (s *mediaService) ListMedia() ([]dto.Media, error) {
	return s.repo.GetAll()
}

func (s *mediaService) DeleteMedia(mediaID string) error {
	return s.repo.Delete(mediaID)
}

func (s *mediaService) UpdateMediaMetadata(mediaID string, metadata dto.Metadata) error {
	return s.repo.UpdateMetadata(mediaID, metadata)
}
