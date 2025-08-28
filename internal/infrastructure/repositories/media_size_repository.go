package repositories

import (
	"file-uploader/internal/domain/dto"
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

func (r *mediaSizeRepository) GetSizeByName(name string) (*dto.MediaSize, error) {
	var size dto.MediaSize
	if err := r.db.First(&size, "media_name = ?", name).Error; err != nil {
		return nil, err
	}
	return &size, nil
}

func (r *mediaSizeRepository) UpdateSize(size *dto.MediaSize) error {
	return r.db.Save(size).Error
}

func (r *mediaSizeRepository) DeleteSize(name string) error {
	return r.db.Delete(&dto.MediaSize{}, "media_name = ?", name).Error
}
