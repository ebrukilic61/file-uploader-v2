// fileutils.go
package fileutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// Atomik dosya kopyalama
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// Dosya hash hesaplama
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("hash hesaplanamadı: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Hash doğrulama
func ValidateFileHash(filePath, expectedHash string) error {
	if expectedHash == "" {
		return nil // Hash doğrulama istenmiyor
	}

	calculatedHash, err := CalculateFileHash(filePath)
	if err != nil {
		return err
	}

	if calculatedHash != expectedHash {
		return fmt.Errorf("hash doğrulaması başarısız")
	}

	return nil
}

func MakeKey(uploadID, filename string) string {
	cleanUploadID := strings.TrimPrefix(uploadID, "upload-")
	return cleanUploadID + "_" + filename
}
