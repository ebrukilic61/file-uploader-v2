package file

import (
	"path/filepath"
	"strings"
)

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
