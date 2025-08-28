package dto

import "time"

type Media struct {
	MediaID      string     `json:"id"`
	Filename     string     `json:"file_name"`
	FileType     string     `json:"file_type"` // image / video
	FileSize     int64      `json:"file_size"`
	Status       string     `json:"status"` // pending, processing, completed, failed
	OriginalPath string     `json:"original_path"`
	Jobs         []MediaJob `json:"jobs,omitempty"` // her bir iş
	Metadata     Metadata   `json:"metadata,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// İş tanımı
type MediaJob struct {
	JobID      string            `json:"job_id"`
	MediaID    string            `json:"media_id"`
	Type       string            `json:"type"`             // resize, compress
	Status     string            `json:"status"`           // pending, processing, completed, failed
	Params     map[string]string `json:"params,omitempty"` // width/height veya resolution
	Width      int               `json:"width,omitempty"`
	Height     int               `json:"height,omitempty"`
	OutputPath string            `json:"output_path,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type Metadata struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Format   string `json:"format,omitempty"`   // jpg, png vs.
	Duration int    `json:"duration,omitempty"` // video için
	Size     int64  `json:"size,omitempty"`     // byte cinsinden
}

// API üzerinden media register isteği
type MediaRegisterRequestDTO struct {
	Filename string `json:"filename"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	FilePath string `json:"file_path"`
}

// API üzerinden job bilgisi dönebilir
type MediaJobResponseDTO struct {
	JobID      string            `json:"job_id"`
	Type       string            `json:"type"`
	Status     string            `json:"status"`
	Params     map[string]string `json:"params"`
	OutputPath string            `json:"output_path"`
}
