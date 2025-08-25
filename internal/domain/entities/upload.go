package entities

import "time"

// Upload represents a file upload session
type Upload struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	TotalChunks int       `json:"total_chunks"`
	Status      string    `json:"status"` // "uploading", "completed", "failed"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UploadChunk represents a single chunk of a file upload
type UploadChunk struct {
	UploadID   string    `json:"upload_id"`
	ChunkIndex int       `json:"chunk_index"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	Hash       string    `json:"hash,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// UploadStatus represents the current status of an upload
type UploadStatus struct {
	UploadID       string `json:"upload_id"`
	Filename       string `json:"filename"`
	TotalChunks    int    `json:"total_chunks"`
	UploadedChunks int    `json:"uploaded_chunks"`
	Progress       int    `json:"progress"` // Percentage (0-100)
	Status         string `json:"status"`
}
