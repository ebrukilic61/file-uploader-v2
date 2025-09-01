package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		BasePath: basePath,
	}
}

func (l *LocalStorage) Upload(file multipart.File, metadata map[string]string) (string, error) {
	filename := metadata["filename"]
	folder := metadata["folder"]
	fullPath := filepath.Join(l.BasePath, folder, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return "", fmt.Errorf("klasör oluşturulamadı: %w", err)
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("dosya yazılamadı: %w", err)
	}

	return fullPath, nil
}

func (l *LocalStorage) Download(fileID string) (multipart.File, error) {
	return os.Open(filepath.Join(l.BasePath, fileID))
}

func (l *LocalStorage) Delete(fileID string) error {
	return os.Remove(filepath.Join(l.BasePath, fileID))
}

// UploadOriginal - Original image'i uploads/original klasörüne kaydeder
func (m *LocalStorage) UploadOriginal(file multipart.File, filename string) (string, error) {
	originalPath := filepath.Join(m.BasePath, "media/original")
	return m.uploadFile(file, originalPath, filename)
}

// UploadVariant - Variant image'i uploads/media/variants klasörüne kaydeder
func (m *LocalStorage) UploadVariant(file multipart.File, filename string) (string, error) {
	variantsPath := filepath.Join(m.BasePath, "media/variants")
	return m.uploadFile(file, variantsPath, filename)
}

// Upload - StorageStrategy interface'i için genel upload metodu
func (m *LocalStorage) UploadImage(file multipart.File, metadata map[string]string) (string, error) {
	filename := metadata["name"]
	folder := metadata["folder"]

	if folder == "media/original" {
		return m.UploadOriginal(file, filename)
	} else if folder == "variants" {
		return m.UploadVariant(file, filename)
	}

	// Default olarak original klasörüne kaydet
	return m.UploadOriginal(file, filename)
}

// uploadFile - Genel dosya upload metodu
func (m *LocalStorage) uploadFile(file multipart.File, folderPath, filename string) (string, error) {
	// Klasörü oluştur
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("klasör oluşturulamadı: %w", err)
	}

	// Dosya yolunu oluştur
	fullPath := filepath.Join(folderPath, filename)

	// Dosyayı oluştur
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer outFile.Close()

	// Dosyayı kopyala
	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("dosya yazılamadı: %w", err)
	}

	return fullPath, nil
}

// CopyFile - Bir dosyayı başka bir yere kopyalar (variant oluşturmak için)
func (m *LocalStorage) CopyFile(sourcePath, destinationPath string) error {
	// Hedef klasörü oluştur
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("hedef klasör oluşturulamadı: %w", err)
	}

	// Kaynak dosyayı aç
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("kaynak dosya açılamadı: %w", err)
	}
	defer sourceFile.Close()

	// Hedef dosyayı oluştur
	destFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("hedef dosya oluşturulamadı: %w", err)
	}
	defer destFile.Close()

	// Dosyayı kopyala
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("dosya kopyalanamadı: %w", err)
	}

	return nil
}

// GetVariantPath - Variant dosyası için path oluşturur
func (m *LocalStorage) GetVariantPath(originalFilename, variantType string) string {
	// Dosya uzantısını al
	ext := filepath.Ext(originalFilename)
	// Dosya adını al (uzantısız)
	nameWithoutExt := strings.TrimSuffix(originalFilename, ext)
	// Variant dosya adını oluştur
	variantFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, variantType, ext)

	return filepath.Join(m.BasePath, "media/variants", variantFilename)
}

// GetOriginalPath - Original dosya için path oluşturur
func (m *LocalStorage) GetOriginalPath(filename string) string {
	return filepath.Join(m.BasePath, "media/original", filename)
}

// DeleteFile - Dosyayı siler
func (m *LocalStorage) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

// FileExists - Dosyanın var olup olmadığını kontrol eder
func (m *LocalStorage) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
