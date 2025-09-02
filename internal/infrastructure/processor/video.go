package processor

import (
	"file-uploader/internal/domain/dto"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mowshon/moviego"
)

// Video işle
func ProcessVideoFile(mediaService MediaService, filename, finalFilePath string) error {
	videoDTO := &dto.VideoDTO{
		OriginalName: filename,
		FileType:     getMimeTypeFromExtension(filename),
		FilePath:     finalFilePath,
		Status:       "processing",
		Height:       0,
		Width:        0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	file, err := os.Open(finalFilePath)
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	if err := mediaService.CreateVideo(videoDTO); err != nil {
		return fmt.Errorf("video oluşturulamadı: %w", err)
	}

	if err := mediaService.ResizeVideo(videoDTO.VideoID, 1920, 1280, videoDTO); err != nil {
		log.Printf("UYARI: Video boyutlandırma başarısız: %v", err)
	}

	log.Printf("INFO: Video %s başarıyla işlendi. Path: %s", filename, videoDTO.FilePath)
	return nil
}

func ResizeVideo(inputPath string, outputPath string, width int64, height int64) error {
	video, err := moviego.Load(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load video: %w", err)
	}

	if err := video.Resize(width, height).Output(outputPath).Run(); err != nil {
		return fmt.Errorf("failed to resize video: %w", err)
	}

	return nil
}

func ResizeByWidth(inputPath, outputPath string, width int64) error {
	video, err := moviego.Load(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load video: %w", err)
	}

	if err := video.ResizeByWidth(int64(width)).
		Output(outputPath).
		Run(); err != nil {
		return fmt.Errorf("failed to resize video: %w", err)
	}

	return nil
}

// Height'e göre resize (aspect ratio korunur)
func ResizeByHeight(inputPath, outputPath string, height int64) error {
	video, err := moviego.Load(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load video: %w", err)
	}

	if err := video.ResizeByHeight(int64(height)).
		Output(outputPath).
		Run(); err != nil {
		return fmt.Errorf("failed to resize video: %w", err)
	}

	return nil
}

func GetVideoDimensions(filePath string) (int64, int64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	w, _ := strconv.ParseInt(parts[0], 10, 64)
	h, _ := strconv.ParseInt(parts[1], 10, 64)
	return w, h, nil
}
