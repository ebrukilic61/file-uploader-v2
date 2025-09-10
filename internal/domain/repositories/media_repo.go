package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
)

//* Single Responsibility Principle (SRP): Her repo sadece bir tabloya odaklanıyor, yönetimi ve test edilmesi kolay

type MediaRepository interface {
	CreateMedia(media *dto.ImageDTO) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
	GetAllMedia() ([]*dto.ImageDTO, error)
	GetMediaByStatus(status string) ([]*dto.ImageDTO, error)
}

type MediaVariantRepository interface {
	//CreateVariant(variant *dto.MediaVariant) error
	CreateVariant(dtoVariant *dto.MediaVariant, repo MediaRepository) error
	RetryCreateVariant(variant *dto.MediaVariant) error
	GetVariantByID(id string) (*dto.MediaVariant, error)
	UpdateVariant(variant *dto.MediaVariant) error
	DeleteVariant(id string) error
}

type MediaSizeRepository interface {
	CreateSize(size *dto.MediaSize) error
	GetSizeByName(name string) (*dto.MediaSize, error)
	GetAllSizes() ([]*dto.MediaSize, error)
	UpdateSize(size *dto.MediaSize) error
	DeleteSize(name string) error
}

type VideoRepository interface {
	CreateVideo(video *dto.VideoDTO) error
	GetVideoByID(id string) (*dto.VideoDTO, error)
	ResizeWidth(video *entities.Video) error
	ResizeHeight(video *entities.Video) error
	ResizeVideo(video *entities.Video) error
}
