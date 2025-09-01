package dto

import "time"

type ImageDTO struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	FilePath     string    `json:"file_path"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MediaVariant struct {
	VariantID   string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"variant_id"`
	MediaID     string `gorm:"type:uuid" json:"media_id"`
	VariantName string `gorm:"type:varchar(100)" json:"variant_name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FilePath    string `gorm:"type:varchar(255)" json:"file_path"`
}

type MediaSize struct {
	VariantType string `json:"variant_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}
