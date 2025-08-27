package entities

import "time"

type Media struct {
	MediaID   string    `json:"id"`
	Filename  string    `json:"file_name"`
	FileType  string    `json:"file_type"`
	FileSize  int64     `json:"file_size"`
	FilePath  string    `json:"file_path"`
	Metadata  Metadata  `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Metadata struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Format   string `json:"format,omitempty"`   // jpg, png vs.
	Duration int    `json:"duration,omitempty"` // video için
	Size     int64  `json:"size,omitempty"`     // byte cinsinden
}
