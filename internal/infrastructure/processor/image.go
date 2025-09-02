package processor

import (
	"file-uploader/internal/domain/dto"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

type ResizeOption struct {
	Width   int
	Height  int
	Quality int // 1-100
}

type MediaService interface {
	CreateMedia(media *dto.ImageDTO, filePath string) error
	CreateVariantsForMedia(mediaID string, filePath string) error
	CreateVideo(video *dto.VideoDTO) error
	ResizeVideo(videoID string, width int64, height int64, video *dto.VideoDTO) error
}

// Image işle
func ProcessImageFile(mediaService MediaService, filename, finalFilePath string) error {
	imageDTO := &dto.ImageDTO{
		OriginalName: filename,
		FileType:     getMimeTypeFromExtension(filename),
		FilePath:     finalFilePath,
		Status:       "processing",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	file, err := os.Open(finalFilePath)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	if err := mediaService.CreateMedia(imageDTO, finalFilePath); err != nil {
		return fmt.Errorf("media oluşturulamadı: %w", err)
	}

	if err := mediaService.CreateVariantsForMedia(imageDTO.ID, finalFilePath); err != nil {
		return fmt.Errorf("media varyantları oluşturulamadı: %w", err)
	}

	log.Printf("INFO: Image %s başarıyla işlendi. Path: %s", filename, imageDTO.FilePath)
	return nil
}

func getMimeTypeFromExtension(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/avi"
	case ".mkv":
		return "video/mkv"
	default:
		return "application/octet-stream"
	}
}

func ResizeImage(inputPath, outputPath string, options ResizeOption) (string, error) {
	img, err := imaging.Open(inputPath)
	if err != nil {
		return "", err
	}

	resizedImg := imaging.Fit(img, options.Width, options.Height, imaging.Lanczos)

	err = imaging.Save(resizedImg, outputPath, imaging.JPEGQuality(options.Quality))
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

func ResizeAndSaveMultiple(inputPath, outputDir string, options []ResizeOption) ([]string, error) {
	img, err := imaging.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("resim açılamadı: %w", err)
	}

	var savedFiles []string
	base := filepath.Base(inputPath)

	for _, opt := range options {
		// Oran koruyarak resize
		resizedImg := imaging.Fit(img, opt.Width, opt.Height, imaging.Lanczos)

		outputPath := filepath.Join(outputDir,
			fmt.Sprintf("resized_%dx%d_%s", opt.Width, opt.Height, base),
		)

		// JPEG kalite ayarı ile kaydet
		err := imaging.Save(resizedImg, outputPath, imaging.JPEGQuality(opt.Quality))
		if err != nil {
			return savedFiles, fmt.Errorf("dosya kaydedilemedi: %w", err)
		}

		savedFiles = append(savedFiles, outputPath)
	}

	return savedFiles, nil
}

// Örnek Kullanım:
/*
options := []processor.ResizeOption{
	{Width: 1000, Height: 1000, Quality: 100},
	{Width: 800, Height: 800, Quality: 90},
	{Width: 200, Height: 200, Quality: 80},
}

savedFiles, err := processor.ResizeAndSaveMultiple("input.jpg", "output/", options)
if err != nil {
	fmt.Println("Hata:", err)
} else {
	fmt.Println("Oluşturulan dosyalar:", savedFiles)
}
*/
