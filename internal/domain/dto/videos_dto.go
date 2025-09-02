package dto

import "time"

type VideoDTO struct {
	VideoID      string    `json:"video_id"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	FilePath     string    `json:"file_path"`
	Status       string    `json:"status"`
	Height       int64     `json:"height"`
	Width        int64     `json:"width"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Output struct {
	video VideoDTO `json:"video"`
}
