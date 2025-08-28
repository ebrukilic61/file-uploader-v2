package handlers

import (
	"file-uploader/internal/domain/repositories"
)

type MediaHandler struct {
	repo repositories.MediaRepository
}

func NewMediaHandler(repo repositories.MediaRepository) *MediaHandler {
	return &MediaHandler{repo: repo}
}
