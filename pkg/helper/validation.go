package helper

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func GetMimeTypeFromExtension(filename string) string {
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

func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif"}

	for _, imgExt := range imageExtensions {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func IsVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExtensions := []string{".mp4", ".avi", ".mkv"}
	for _, v := range videoExtensions {
		if ext == v {
			return true
		}
	}
	return false
}
