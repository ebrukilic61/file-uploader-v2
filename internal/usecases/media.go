package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
)

type MediaService interface {
	GetMedia(mediaID string) (*dto.Media, error)
	ListMedia() ([]dto.Media, error)
	DeleteMedia(mediaID string) error
	UpdateMediaMetadata(mediaID string, metadata dto.Metadata) error
}

type mediaService struct {
	repo repositories.MediaRepository
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
