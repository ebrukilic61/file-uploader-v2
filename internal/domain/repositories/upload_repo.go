package repositories

import (
	"mime/multipart"
)

type FileUploadRepository interface {
	//Chunk işlemleri
	SaveChunk(uploadID, filename string, chunkIndex int, file multipart.File) error
	ChunkExists(uploadID, filename string, chunkIndex int) bool
	SetUploadedChunks(uploadID, filename string, merged int) error
	GetUploadedChunks(uploadID, filename string) (int, bool)
	// Dosya birleştirme / hash doğrulama / temizlik
	MergeChunks(uploadID, filename string, totalChunks int) (string, error)
	SaveFailedUpload(string, string, string, string, []byte) error
	GetFailedUpload(uploadID string) string
	DeleteFailedUpload(uploadID string) error
	RetryMerge(uploadID, filename string) (string, int, error)
	UpdateRetryStatus(uploadID, status string) error
	CleanupTempFiles(uploadID string) error
	UploadsDir() string
	TempDir() string
}
