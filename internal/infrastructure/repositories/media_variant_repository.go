package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
	"file-uploader/internal/domain/repositories"
	"time"

	"github.com/google/uuid"
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

func (r *mediaVariantRepository) CreateVariant(dtoVariant *dto.MediaVariant) error {
	entity := r.dtoToEntity(dtoVariant)
	if err := r.db.Create(entity).Error; err != nil {
		return err
	}
	*dtoVariant = *r.entityToDTO(entity)
	return nil
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

func (r *mediaVariantRepository) dtoToEntity(dtoVariant *dto.MediaVariant) *entities.MediaVariant {
	return &entities.MediaVariant{
		VariantID:   uuid.MustParse(dtoVariant.VariantID),
		MediaID:     uuid.MustParse(dtoVariant.MediaID),
		VariantName: dtoVariant.VariantName,
		Width:       dtoVariant.Width,
		Height:      dtoVariant.Height,
		FilePath:    dtoVariant.FilePath,
		CreatedAt:   time.Now(),
	}
}

func (r *mediaVariantRepository) entityToDTO(entity *entities.MediaVariant) *dto.MediaVariant {
	return &dto.MediaVariant{
		VariantID:   entity.VariantID.String(),
		MediaID:     entity.MediaID.String(),
		VariantName: entity.VariantName,
		Width:       entity.Width,
		Height:      entity.Height,
		FilePath:    entity.FilePath,
	}
}
