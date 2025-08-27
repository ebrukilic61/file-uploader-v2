package repositories

import "file-uploader/internal/domain/dto"

/*
type MediaRepository interface {
	RegisterMedia(media *dto.Media) error
	GetByID(mediaID string) (*dto.Media, error)
	GetAll() ([]dto.Media, error)
	Delete(mediaID string) error
	UpdateMetadata(mediaID string, metadata dto.Metadata) error
}
*/

type MediaRepository interface {
	Create(media *dto.Media) error
	GetByID(mediaID string) (*dto.Media, error)
	GetAll() ([]dto.Media, error)
	ListMedia(filter dto.MediaFilter) ([]dto.Media, error)
	UpdateMetadata(mediaID string, metadata dto.Metadata) error
	Delete(mediaID string) error
}
