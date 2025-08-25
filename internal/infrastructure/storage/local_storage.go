package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func (l *LocalStorage) Upload(file multipart.File, metadata map[string]string) (string, error) {
	filename := metadata["filename"]
	folder := metadata["folder"]
	fullPath := filepath.Join(l.BasePath, folder, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return "", fmt.Errorf("klasör oluşturulamadı: %w", err)
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("dosya yazılamadı: %w", err)
	}

	return fullPath, nil
}

func (l *LocalStorage) Download(fileID string) (multipart.File, error) {
	return os.Open(filepath.Join(l.BasePath, fileID))
}

func (l *LocalStorage) Delete(fileID string) error {
	return os.Remove(filepath.Join(l.BasePath, fileID))
}
