package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"

	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) repositories.MediaRepository {
	return &mediaRepository{
		db: db,
	}
}

func (r *mediaRepository) CreateMedia(media *dto.ImageDTO) error {
	return r.db.Create(media).Error
}

// Add stubs for other required methods if needed
func (r *mediaRepository) GetMediaByID(id string) (*dto.ImageDTO, error) {
	var media dto.ImageDTO
	if err := r.db.First(&media, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) UpdateMediaStatus(id string, status string) error {
	return r.db.Model(&dto.ImageDTO{}).Where("id = ?", id).Update("status", status).Error
}
