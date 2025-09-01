package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Image struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	OriginalName string
	FileType     string
	FilePath     string
	Status       string `gorm:"type:varchar(20)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"` // soft delete
}

type MediaVariant struct {
	VariantID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	MediaID     uuid.UUID
	VariantName string
	Width       int
	Height      int
	FilePath    string
	CreatedAt   time.Time
}

type MediaSize struct {
	VariantType string `gorm:"primaryKey"`
	Width       int
	Height      int
}

func (m *Image) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
