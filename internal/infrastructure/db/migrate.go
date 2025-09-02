package db

import (
	"file-uploader/internal/domain/entities"
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		//&entities.Image{},
		//&entities.MediaVariant{},
		//&entities.MediaSize{},
		&entities.Video{},
	)
	if err != nil {
		log.Fatalf("DB migrate başarısız: %v", err)
	}
	return err
}
