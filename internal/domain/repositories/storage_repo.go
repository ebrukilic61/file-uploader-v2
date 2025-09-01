package repositories

import "mime/multipart"

type StorageStrategy interface {
	UploadImage(file multipart.File, metadata map[string]string) (string, error)
	CopyFile(sourcePath, destinationPath string) error
	GetVariantPath(originalFilename, variantType string) string
	GetOriginalPath(filename string) string
	DeleteFile(filePath string) error
	Delete(fileID string) error
	Download(fileID string) (multipart.File, error)
	FileExists(filePath string) bool
}
