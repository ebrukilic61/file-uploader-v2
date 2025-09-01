package dto

type UploadChunkRequestDTO struct {
	UploadID   string `json:"upload_id" form:"upload_id"`
	ChunkIndex string `json:"chunk_index" form:"chunk_index"`
	Filename   string `json:"filename" form:"filename"`
	ChunkHash  string `json:"chunk_hash" form:"chunk_hash"`
}

type CompleteUploadRequestDTO struct {
	UploadID    string `json:"upload_id" form:"upload_id"`
	TotalChunks int    `json:"total_chunks" form:"total_chunks"`
	Filename    string `json:"filename" form:"filename"`
}

type UploadStatusRequestDTO struct {
	UploadID string `json:"upload_id" form:"upload_id"`
	Filename string `json:"filename" form:"filename"`
}

type CancelUploadRequestDTO struct {
	UploadID string `json:"upload_id" form:"upload_id"`
}

type CancelUploadResponse struct {
	Status  string `json:"status"`            // "ok" veya "failed" şeklinde
	Message string `json:"message,omitempty"` // opsiyonel açıklama
}

type UploadStatusResponse struct {
	UploadID       string `json:"upload_id"`
	Filename       string `json:"filename"`
	UploadedChunks int    `json:"uploaded_chunks"`
	Status         string `json:"status,omitempty"` // "completed", "failed" gibi (opsiyonel)
}

type UploadChunkResponse struct {
	Status     string `json:"status"`
	UploadID   string `json:"upload_id"`
	ChunkIndex int    `json:"chunk_index"`
	Filename   string `json:"filename"`
	Message    string `json:"message,omitempty"`
}

type CompleteUploadResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type ErrorResponse struct {
	Error         string `json:"error"`
	MissingChunks []int  `json:"missing_chunks,omitempty"`
}

type ProcessingResult struct {
	ResizedPath   string
	ConvertedPath string
	ThumbPath     string
	Err           error
}
