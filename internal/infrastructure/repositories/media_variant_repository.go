package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"

	"gorm.io/gorm"
)

type mediaVariantRepository struct {
	db *gorm.DB
}

func NewMediaVariantRepository(db *gorm.DB) repositories.MediaVariantRepository {
	return &mediaVariantRepository{
		db: db,
	}
}

func (r *mediaVariantRepository) CreateVariant(variant *dto.MediaVariant) error {
	return r.db.Create(variant).Error
}

func (r *mediaVariantRepository) GetVariantByID(var_id string) (*dto.MediaVariant, error) {
	var variant dto.MediaVariant
	if err := r.db.First(&variant, var_id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *mediaVariantRepository) GetMediaVariantByID(media_id string) (*dto.MediaVariant, error) {
	var variant dto.MediaVariant
	if err := r.db.First(&variant, media_id).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *mediaVariantRepository) UpdateVariant(variant *dto.MediaVariant) error {
	return r.db.Save(variant).Error
}

func (r *mediaVariantRepository) DeleteVariant(var_id string) error {
	return r.db.Delete(&dto.MediaVariant{}, var_id).Error
}
