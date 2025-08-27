package file

import "fmt"

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
