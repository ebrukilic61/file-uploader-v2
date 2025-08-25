package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Upload   UploadConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type UploadConfig struct {
	TempDir     string
	UploadsDir  string
	MaxFileSize int64 // bytes
	ChunkSize   int64 // bytes
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "3000"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Upload: UploadConfig{
			TempDir:     getEnv("UPLOAD_TEMP_DIR", "temp_uploads"),
			UploadsDir:  getEnv("UPLOAD_DIR", "uploads"),
			MaxFileSize: getEnvAsInt64("UPLOAD_MAX_FILE_SIZE", 5*1024*1024*1024), // 5GB
			ChunkSize:   getEnvAsInt64("UPLOAD_CHUNK_SIZE", 10*1024*1024),        // 10MB
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "file_uploader"),
		},
	}

	// Proje kökü:
	projectRoot, err := findProjectRoot()
	if err != nil {
		panic(err)
	}

	// Klasörleri proje köküne göre oluşturmak için:
	config.Upload.TempDir = filepath.Join(projectRoot, "cmd", "server", config.Upload.TempDir)
	config.Upload.UploadsDir = filepath.Join(projectRoot, "cmd", "server", config.Upload.UploadsDir)

	if err := os.MkdirAll(config.Upload.TempDir, 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(config.Upload.UploadsDir, 0755); err != nil {
		panic(err)
	}

	return config
}

func findProjectRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Root'a ulaştık, go.mod bulunamadı
			return os.Getwd()
		}
		current = parent
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func ensureDir(dir string) error {
	// Mutlak yol
	if !filepath.IsAbs(dir) {
		// Proje kökü:
		projectRoot, err := findProjectRoot()
		if err != nil {
			return err
		}
		dir = filepath.Join(projectRoot, dir)
	}
	return os.MkdirAll(dir, 0755)
}
