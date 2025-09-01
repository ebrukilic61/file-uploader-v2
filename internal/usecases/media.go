package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/processor"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type MediaService interface {
	// Images
	CreateMedia(media *dto.ImageDTO, filepath string) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
	GetAllMedia() ([]*dto.ImageDTO, error)
	GetMediaByStatus(status string) ([]*dto.ImageDTO, error)

	// Media Variant
	CreateVariantsForMedia(mediaID, originalPath string) error
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
func (u *mediaService) CreateMedia(media *dto.ImageDTO, finalPath string) error {
	// İş mantığı: desteklenen dosya tiplerini kontrol et
	if media.FileType != "image/png" && media.FileType != "image/jpeg" && media.FileType != "image/jpg" && media.FileType != "image/gif" {
		return fmt.Errorf("unsupported file type: %s", media.FileType)
	}
	// DTO’da storage pathini güncelle
	media.FilePath = finalPath

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

func (s *mediaService) CreateVariantsForMedia(mediaID, originalPath string) error {
	sizes, err := s.sizeRepo.GetAllSizes()
	if err != nil {
		return fmt.Errorf("failed to get media sizes: %w", err)
	}

	for _, size := range sizes {
		baseName := filepath.Base(originalPath)
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		outputDir := filepath.Join("uploads", "media", "variants", fmt.Sprintf("%s", mediaID)) // id klasörü içerisinde oluşturuldu ki karmaşıklık yaşanmasın
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("variants klasörü oluşturulamadı: %w", err)
		}

		variantName := fmt.Sprintf("%s_%s_%dx%d", nameWithoutExt, size.VariantType, size.Width, size.Height)

		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%dx%d%s", nameWithoutExt, size.VariantType, size.Width, size.Height, ext)) // isimlendirme
		resizedPath, err := processor.ResizeImage(originalPath, outputPath, processor.ResizeOption{
			Width:   size.Width,
			Height:  size.Height,
			Quality: 100,
		})

		if err != nil {
			return fmt.Errorf("görsel için yeniden boyutlandırma hatası: %w", err)
		}

		variant := &dto.MediaVariant{
			VariantID:   uuid.New().String(),
			MediaID:     mediaID,
			FilePath:    resizedPath,
			Width:       size.Width,
			Height:      size.Height,
			VariantName: variantName,
		}

		// DB’ye kaydet
		if err := s.variantRepo.CreateVariant(variant); err != nil {
			return fmt.Errorf("media varyantı oluşturulamadı: %w", err)
		}
	}

	return nil
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
