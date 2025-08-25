package processor

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func ConvertVideo(inputPath, outputPath, format string) (string, error) {
	base := filepath.Base(inputPath)
	outputPath = filepath.Join(filepath.Dir(outputPath), fmt.Sprintf("%s_converted.%s", base, format))
	cmd := exec.Command("ffmpeg", "-i", inputPath, outputPath) //* winget install ffmpeg -> indirilmesi lazım
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputPath, nil
}

func GenerateThumbnail(inputPath, outputPath string, timePosition string) (string, error) {
	base := filepath.Base(inputPath)
	outputPath = filepath.Join(filepath.Dir(outputPath), fmt.Sprintf("%s_thumbnail.jpg", base))
	cmd := exec.Command("ffmpeg", "-ss", timePosition, "-i", inputPath, "-vframes", "1", outputPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputPath, nil
}
