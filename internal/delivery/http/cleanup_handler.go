package handlers

import (
	"file-uploader/internal/usecases"
)

type CleanupHandler struct {
	cleanupUC usecases.CleanupService
}

func NewCleanupHandler(cleanupUC usecases.CleanupService) *CleanupHandler {
	return &CleanupHandler{
		cleanupUC: cleanupUC,
	}
}

// Manuel trigger için
func (h *CleanupHandler) CancelUpload(uploadID string) error {
	return h.cleanupUC.CleanupTempFiles(uploadID)
}
