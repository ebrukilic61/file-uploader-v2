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
	VariantID   string `json:"variant_id"`
	MediaID     string `json:"media_id"`
	VariantName string `json:"variant_name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FilePath    string `json:"file_path"`
}

type MediaSize struct {
	VariantType string `json:"variant_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type VariantRequestDTO struct {
	FilePath string `json:"file_path"`
}
