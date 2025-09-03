package usecases

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/processor"
	"file-uploader/pkg/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type MediaService interface { //video da eklenecek
	// Images
	CreateMedia(media *dto.ImageDTO, filepath string) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
	GetAllMedia() ([]*dto.ImageDTO, error)

	// Media Variant
	CreateVariantsForMedia(mediaID, originalPath string) error

	// Media Size
	CreateSize(size *dto.MediaSize) error
	UpdateSize(size *dto.MediaSize) error

	//Video
	CreateVideo(video *dto.VideoDTO) error
	GetVideoByID(id string) (*dto.VideoDTO, error)
	ResizeByWidth(id string, width int64, video *dto.VideoDTO) error
	ResizeByHeight(id string, height int64, video *dto.VideoDTO) error
	ResizeVideo(id string, width int64, height int64, video *dto.VideoDTO) error
}

type mediaService struct {
	mediaRepo   repositories.MediaRepository
	variantRepo repositories.MediaVariantRepository
	sizeRepo    repositories.MediaSizeRepository
	storage     repositories.StorageStrategy
	videoRepo   repositories.VideoRepository
}

func NewMediaService(
	mediaRepo repositories.MediaRepository,
	variantRepo repositories.MediaVariantRepository,
	sizeRepo repositories.MediaSizeRepository,
	storage repositories.StorageStrategy,
	videoRepo repositories.VideoRepository,
) MediaService {
	return &mediaService{
		mediaRepo:   mediaRepo,
		variantRepo: variantRepo,
		sizeRepo:    sizeRepo,
		storage:     storage,
		videoRepo:   videoRepo,
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

func (s *mediaService) CreateVideo(video *dto.VideoDTO) error {
	if video.FileType != "video/mp4" && video.FileType != "video/avi" && video.FileType != "video/mkv" {
		return fmt.Errorf("unsupported file type: %s", video.FileType)
	}
	if _, err := os.Stat(video.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("video file does not exist at path: %s", video.FilePath)
	}
	return s.videoRepo.CreateVideo(video)
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

// Media Size
func (s *mediaService) CreateSize(size *dto.MediaSize) error {
	return s.sizeRepo.CreateSize(size)
}

func (s *mediaService) UpdateSize(size *dto.MediaSize) error {
	return s.sizeRepo.UpdateSize(size)
}

// Video:
func (s *mediaService) GetVideoByID(id string) (*dto.VideoDTO, error) {
	return s.videoRepo.GetVideoByID(id)
}

func (s *mediaService) ResizeByWidth(id string, width int64, video *dto.VideoDTO) error {
	inputPath := fmt.Sprintf("./uploads/videos/original/%s%s", video.VideoID, filepath.Ext(video.FilePath))
	if video.Width <= 0 || video.Height <= 0 {
		origWidth, origHeight, err := helper.GetVideoDimensions(inputPath)
		log.Printf("Orijinal video boyutları: %dx%d", origWidth, origHeight)
		if err != nil {
			return fmt.Errorf("orijinal video boyutu alınamadı: %w", err)
		}
		video.Width = origWidth
		video.Height = origHeight
	}

	// orantılı height hesaplamak için:
	newHeight := int64((float64(video.Height) / float64(video.Width)) * float64(width))
	if newHeight%2 != 0 {
		newHeight += 1
	}

	video.Height = newHeight
	video.Width = width

	// Fiziksel dosya resize
	outputPath := fmt.Sprintf("./uploads/videos/resized/%s_%dx%d%s", video.VideoID, video.Width, video.Height, filepath.Ext(video.FilePath))
	if err := processor.ResizeByWidth(inputPath, outputPath, video.Width); err != nil {
		return err
	}
	video.FilePath = outputPath
	video.Status = "resized"

	// DTO -> Entity
	entity := entities.Video{
		VideoID:      uuid.MustParse(video.VideoID),
		Width:        video.Width,
		Height:       video.Height,
		Status:       video.Status,
		FilePath:     video.FilePath,
		OriginalName: video.OriginalName,
		FileType:     video.FileType,
	}

	return s.videoRepo.ResizeWidth(&entity)
}

func (s *mediaService) ResizeByHeight(id string, height int64, video *dto.VideoDTO) error {
	if video.Height <= 0 {
		return fmt.Errorf("orijinal video yüksekliği 0 veya bilinmiyor")
	}
	newWidth := int64((float64(video.Width) / float64(video.Height)) * float64(height))
	if newWidth%2 != 0 {
		newWidth += 1
	}
	video.Width = newWidth
	video.Height = height
	video.Status = "resized"

	// Fiziksel dosya resize
	outputPath := fmt.Sprintf("./uploads/videos/resized/%s_%dx%d%s", video.VideoID, video.Width, video.Height, filepath.Ext(video.FilePath))
	if err := processor.ResizeByHeight(video.FilePath, outputPath, video.Height); err != nil {
		return err
	}
	video.FilePath = outputPath

	// DTO -> Entity
	entity := entities.Video{
		VideoID:      uuid.MustParse(video.VideoID),
		Width:        video.Width,
		Height:       video.Height,
		Status:       video.Status,
		FilePath:     video.FilePath,
		OriginalName: video.OriginalName,
		FileType:     video.FileType,
	}

	return s.videoRepo.ResizeHeight(&entity)
}

func (s *mediaService) ResizeVideo(id string, width int64, height int64, video *dto.VideoDTO) error {
	video.Width = width
	video.Height = height

	baseDir, err := os.Getwd() // absolute path
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	resizedDir := filepath.Join(baseDir, "uploads", "videos", "resized")
	if err := os.MkdirAll(resizedDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create resized dir: %w", err)
	}

	outputPath := filepath.Join(
		resizedDir,
		fmt.Sprintf("%s_%dx%d%s", video.VideoID, video.Width, video.Height, filepath.Ext(video.FilePath)),
	)

	// ffmpeg ile resize
	if err := processor.ResizeVideo(video.FilePath, outputPath, width, height); err != nil {
		return fmt.Errorf("failed to resize video: %w", err)
	}

	video.FilePath = outputPath
	video.Status = "resized"

	// DTO -> Entity
	entity := entities.Video{
		VideoID:      uuid.MustParse(video.VideoID),
		Width:        video.Width,
		Height:       video.Height,
		Status:       video.Status,
		FilePath:     video.FilePath,
		OriginalName: video.OriginalName,
		FileType:     video.FileType,
	}

	return s.videoRepo.ResizeVideo(&entity)
}
