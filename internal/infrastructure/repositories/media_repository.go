package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
	"file-uploader/internal/domain/repositories"

	"github.com/google/uuid"
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

func (r *mediaRepository) CreateMedia(mediaDTO *dto.ImageDTO) error {
	entity := r.dtoToEntity(mediaDTO)
	if err := r.db.Create(entity).Error; err != nil {
		return err
	}
	*mediaDTO = *r.entityToDTO(entity)
	return nil
}

func (r *mediaRepository) GetMediaByID(id string) (*dto.ImageDTO, error) {
	// String ID'yi UUID'ye dönüştür
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var entity entities.Image
	if err := r.db.First(&entity, "id = ?", parsedID).Error; err != nil {
		return nil, err
	}

	// Helper kullanarak Entity'den DTO'ya dönüşüm
	return r.entityToDTO(&entity), nil
}

func (r *mediaRepository) UpdateMediaStatus(id string, status string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	return r.db.Model(&entities.Image{}).Where("id = ?", parsedID).Update("status", status).Error
}

func (r *mediaRepository) GetAllMedia() ([]*dto.ImageDTO, error) {
	var entities []entities.Image
	if err := r.db.Find(&entities).Error; err != nil {
		return nil, err
	}

	// Slice dönüşümü için helper kullanımı
	var dtos []*dto.ImageDTO
	for _, entity := range entities {
		dtos = append(dtos, r.entityToDTO(&entity))
	}

	return dtos, nil
}

func (r *mediaRepository) GetMediaByStatus(status string) ([]*dto.ImageDTO, error) {
	var entities []entities.Image
	if err := r.db.Where("status = ?", status).Find(&entities).Error; err != nil {
		return nil, err
	}

	// Helper ile slice dönüşümü
	return r.entitiesToDTOs(entities), nil
}

func (r *mediaRepository) dtoToEntity(mediaDTO *dto.ImageDTO) *entities.Image {
	media := &entities.Image{
		OriginalName: mediaDTO.OriginalName,
		FileType:     mediaDTO.FileType,
		FilePath:     mediaDTO.FilePath,
		Status:       mediaDTO.Status,
	}
	if mediaDTO.ID != "" {
		if parsedID, err := uuid.Parse(mediaDTO.ID); err == nil {
			media.ID = parsedID
		}
	}
	return media
}

func (r *mediaRepository) entityToDTO(entity *entities.Image) *dto.ImageDTO {
	return &dto.ImageDTO{
		ID:           entity.ID.String(),
		OriginalName: entity.OriginalName,
		FileType:     entity.FileType,
		FilePath:     entity.FilePath,
		Status:       entity.Status,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}

func (r *mediaRepository) entitiesToDTOs(entities []entities.Image) []*dto.ImageDTO {
	var dtos []*dto.ImageDTO
	for _, entity := range entities {
		dtos = append(dtos, r.entityToDTO(&entity))
	}
	return dtos
}

// Helper function for bulk DTO to Entity conversion
func (r *mediaRepository) dtosToEntities(dtos []*dto.ImageDTO) []entities.Image {
	var entities []entities.Image
	for _, dto := range dtos {
		entities = append(entities, *r.dtoToEntity(dto))
	}
	return entities
}
