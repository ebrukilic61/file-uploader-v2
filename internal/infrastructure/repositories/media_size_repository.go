package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
	"file-uploader/internal/domain/repositories"

	"gorm.io/gorm"
)

type mediaSizeRepository struct {
	db *gorm.DB
}

func NewMediaSizeRepository(db *gorm.DB) repositories.MediaSizeRepository {
	return &mediaSizeRepository{
		db: db,
	}
}

func (r *mediaSizeRepository) CreateSize(size *dto.MediaSize) error {
	return r.db.Create(size).Error
}

/* //Örnek çalışma şekli
size := &dto.MediaSize{
	Name:   "Thumbnail",
	Width:  150,
	Height: 150,
}

*/

func (r *mediaSizeRepository) GetAllSizes() ([]*dto.MediaSize, error) {
	var sizes []*dto.MediaSize
	if err := r.db.Find(&sizes).Error; err != nil {
		return nil, err
	}
	return sizes, nil
}

func (r *mediaSizeRepository) GetSizeByName(name string) (*dto.MediaSize, error) {
	var size dto.MediaSize
	if err := r.db.First(&size, "variant_type = ?", name).Error; err != nil {
		return nil, err
	}
	return &size, nil
}

func (r *mediaSizeRepository) UpdateSize(size *dto.MediaSize) error {
	var existingSizes entities.MediaSize
	if err := r.db.First(&existingSizes, "variant_type = ?", size.VariantType).Error; err != nil {
		return err
	}
	existingSizes.Width = size.Width
	existingSizes.Height = size.Height
	return r.db.Save(&existingSizes).Error
}

func (r *mediaSizeRepository) DeleteSize(name string) error {
	return r.db.Delete(&dto.MediaSize{}, "variant_type = ?", name).Error
}
