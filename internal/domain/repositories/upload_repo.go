package repositories

import (
	"mime/multipart"
)

type FileUploadRepository interface {
	//Chunk işlemleri
	SaveChunk(uploadID, filename string, chunkIndex int, file multipart.File) error
	ChunkExists(uploadID, filename string, chunkIndex int) bool
	SetUploadedChunks(uploadID, filename string, merged int)
	GetUploadedChunks(uploadID, filename string) (int, bool)
	// Dosya birleştirme / hash doğrulama / temizlik
	MergeChunks(uploadID, filename string, totalChunks int) error
	CleanupTempFiles(uploadID string) error
	UploadsDir() string
	TempDir() string
}

type StorageStrategy interface { //! bu uploaddan ayrılabilir!!!
	Upload(file multipart.File, metadata map[string]string) (string, error)
	Download(fileID string) (multipart.File, error)
	Delete(fileID string) error
}
