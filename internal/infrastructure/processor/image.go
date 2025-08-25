package processor

import (
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
)

func ResizeImage(inputPath, outputPath string, width, height int) (string, error) {
	// Görüntüyü dosyadan aç
	img, err := imaging.Open(inputPath)
	if err != nil {
		return "", err
	}

	// Görüntüyü yeniden boyutlandırmak için:
	resizedImg := imaging.Resize(img, width, height, imaging.Lanczos)

	base := filepath.Base(inputPath)
	outputPath = filepath.Join(filepath.Dir(outputPath), "resized_"+strconv.Itoa(width)+"_"+strconv.Itoa(height)+"_"+base)

	// kaydetmek için:
	err = imaging.Save(resizedImg, outputPath)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
