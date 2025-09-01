package repositories

import "file-uploader/internal/domain/dto"

//* Single Responsibility Principle (SRP): Her repo sadece bir tabloya odaklanıyor, yönetimi ve test edilmesi kolay

type MediaRepository interface {
	CreateMedia(media *dto.ImageDTO) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
	GetAllMedia() ([]*dto.ImageDTO, error)
	GetMediaByStatus(status string) ([]*dto.ImageDTO, error)
}

type MediaVariantRepository interface {
	CreateVariant(variant *dto.MediaVariant) error
	GetVariantByID(id string) (*dto.MediaVariant, error)
	UpdateVariant(variant *dto.MediaVariant) error
	DeleteVariant(id string) error
}

type MediaSizeRepository interface {
	CreateSize(size *dto.MediaSize) error
	GetSizeByName(name string) (*dto.MediaSize, error)
	UpdateSize(size *dto.MediaSize) error
	DeleteSize(name string) error
}
