package entities

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	VideoID      uuid.UUID `gorm:"type:uuid;primaryKey"` // DB’de UUID tipinde
	OriginalName string    `gorm:"type:varchar(255);not null"`
	FileType     string    `gorm:"type:varchar(50)"`
	FilePath     string    `gorm:"type:varchar(500);not null"`
	Status       string    `gorm:"type:varchar(50)"`
	Height       int64
	Width        int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
