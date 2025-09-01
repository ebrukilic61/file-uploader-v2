package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"fmt"
	"mime/multipart"
)

type MediaService interface {
	// Images
	CreateMedia(media *dto.ImageDTO, file multipart.File) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
	GetAllMedia() ([]*dto.ImageDTO, error)
	GetMediaByStatus(status string) ([]*dto.ImageDTO, error)

	// Media Variant
	CreateVariant(variant *dto.MediaVariant, variant_name string) error
	GetVariantByID(id string) (*dto.MediaVariant, error)
	UpdateVariant(variant *dto.MediaVariant) error
	DeleteVariant(id string) error

	// Media Size
	CreateSize(size *dto.MediaSize) error
	GetSizeByName(name string) (*dto.MediaSize, error)
	UpdateSize(size *dto.MediaSize) error
	DeleteSize(name string) error
}

type mediaService struct {
	mediaRepo   repositories.MediaRepository
	variantRepo repositories.MediaVariantRepository
	sizeRepo    repositories.MediaSizeRepository
	storage     repositories.StorageStrategy
}

func NewMediaService(
	mediaRepo repositories.MediaRepository,
	variantRepo repositories.MediaVariantRepository,
	sizeRepo repositories.MediaSizeRepository,
	storage repositories.StorageStrategy,
) MediaService {
	return &mediaService{
		mediaRepo:   mediaRepo,
		variantRepo: variantRepo,
		sizeRepo:    sizeRepo,
		storage:     storage,
	}
}

// Images
func (u *mediaService) CreateMedia(media *dto.ImageDTO, file multipart.File) error {
	// İş mantığı: desteklenen dosya tiplerini kontrol et
	if media.FileType != "image/png" && media.FileType != "image/jpeg" && media.FileType != "image/jpg" && media.FileType != "image/gif" {
		return fmt.Errorf("unsupported file type: %s", media.FileType)
	}

	metadata := map[string]string{
		"name": media.OriginalName,
		"type": media.FileType,
	}

	filePath, err := u.storage.UploadImage(file, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload media: %w", err)
	}

	// DTO’da storage pathini güncelle
	media.FilePath = filePath

	// DB’ye kaydet
	return u.mediaRepo.CreateMedia(media)
}

func (s *mediaService) GetMediaByID(id string) (*dto.ImageDTO, error) {
	return s.mediaRepo.GetMediaByID(id)
}

func (s *mediaService) UpdateMediaStatus(id string, status string) error {
	validStatuses := []string{"active", "inactive", "deleted"}
	if !contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.mediaRepo.UpdateMediaStatus(id, status)
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (u *mediaService) GetAllMedia() ([]*dto.ImageDTO, error) {
	return u.mediaRepo.GetAllMedia()
}

func (u *mediaService) GetMediaByStatus(status string) ([]*dto.ImageDTO, error) {
	return u.mediaRepo.GetMediaByStatus(status)
}

// Media Variant -> varyant tipi belirlemek gerekir mi
func (s *mediaService) CreateVariant(variant *dto.MediaVariant, variant_type string) error { //bir medya için varyant üretir
	// variant_type kontrol için eklendi
	if variant_type == "" {
		return fmt.Errorf("variant_type is required") // ek olarak belli başlı variant isimleri olacak, onalardan biri olmalı, bu bilgi de media sizes tablosu içinde!!!
	}
	variant.VariantName = variant_type
	return s.variantRepo.CreateVariant(variant)
}

func (s *mediaService) GetVariantByID(id string) (*dto.MediaVariant, error) {
	return s.variantRepo.GetVariantByID(id)
}

func (s *mediaService) UpdateVariant(variant *dto.MediaVariant) error {
	return s.variantRepo.UpdateVariant(variant)
}

func (s *mediaService) DeleteVariant(id string) error {
	return s.variantRepo.DeleteVariant(id)
}

// Media Size
func (s *mediaService) CreateSize(size *dto.MediaSize) error {
	return s.sizeRepo.CreateSize(size)
}

func (s *mediaService) GetSizeByName(name string) (*dto.MediaSize, error) {
	return s.sizeRepo.GetSizeByName(name)
}

func (s *mediaService) UpdateSize(size *dto.MediaSize) error {
	return s.sizeRepo.UpdateSize(size)
}

func (s *mediaService) DeleteSize(id string) error {
	return s.sizeRepo.DeleteSize(id)
}
