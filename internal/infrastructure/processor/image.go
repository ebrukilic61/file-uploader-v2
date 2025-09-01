package processor

import (
	"fmt"
	"path/filepath"

	"github.com/disintegration/imaging"
)

type ResizeOption struct {
	Width   int
	Height  int
	Quality int // 1-100
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
