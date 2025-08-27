package usecases

import (
	"file-uploader/internal/domain/repositories"
	"file-uploader/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

type CleanupService interface {
	CleanupTempFiles(uploadID string) error
	CleanupOldTempFiles(maxAge time.Duration) error
}

type cleanupService struct {
	repo repositories.FileUploadRepository
}

func NewCleanupService(repo repositories.FileUploadRepository) CleanupService {
	return &cleanupService{
		repo: repo,
	}
}

func (s *cleanupService) CleanupTempFiles(uploadID string) error {
	return s.repo.CleanupTempFiles(uploadID)
}

func (s *cleanupService) CleanupOldTempFiles(maxAge time.Duration) error {
	tempDir := s.repo.TempDir() // repo’dan tempDir getter
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(tempDir, entry.Name())
			info, err := os.Stat(dirPath)
			if err != nil {
				return errors.ErrCannotStat(err)
			}

			if now.Sub(info.ModTime()) > maxAge {
				if err := os.RemoveAll(dirPath); err != nil {
					return errors.ErrCannotRemove(err)
				} else {
					log.Printf("Removed old temp folder: %s", dirPath) //* BUARAYA RESPONSE GELEBİLİR SUCCESS İÇİN
					return nil
				}
			}
		}
	}
	return nil
}
