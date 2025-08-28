package db

import (
	"file-uploader/internal/domain/entities"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate( //* Migrate edilecek tüm tablolar eklenmeli entity içerisinden
		&entities.ImageDTO{},
		&entities.MediaVariant{},
		&entities.MediaSize{},
	)
}
