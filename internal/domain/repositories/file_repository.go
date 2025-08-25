package repositories

import (
	"file-uploader/internal/pkg/fileutils"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type fileUploadRepository struct {
	tempDir      string
	uploadsDir   string
	mergedChunks map[string]int // key: uploadID + filename, value: merged chunk sayısı
	mutex        sync.RWMutex
}

func (r *fileUploadRepository) UploadsDir() string {
	return r.uploadsDir
}

func NewFileUploadRepository(tempDir, uploadsDir string) *fileUploadRepository {
	return &fileUploadRepository{
		tempDir:    tempDir,
		uploadsDir: uploadsDir,
	}
}

func (r *fileUploadRepository) SaveChunk(uploadID, filename string, chunkIndex int, file multipart.File) error {
	saveDir := filepath.Join(r.tempDir, uploadID)
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("geçici klasör oluşturulamadı: %w", err)
	}

	finalPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, chunkIndex))
	tmpPath := fmt.Sprintf("%s.tmp.%d", finalPath, time.Now().UnixNano())

	// Dosyayı geçici konuma kaydetmek için:
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("geçici dosya oluşturulamadı: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chunk kaydedilemedi: %w", err)
	}

	// Atomik rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		// Fallback: copy + remove
		if r.ChunkExists(uploadID, filename, chunkIndex) {
			os.Remove(tmpPath)
			return nil // Chunk zaten var
		}

		if copyErr := fileutils.CopyFile(tmpPath, finalPath); copyErr != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("chunk yazılamadı: %w", copyErr)
		}
		os.Remove(tmpPath)
	}

	return nil
}

func (r *fileUploadRepository) ChunkExists(uploadID, filename string, chunkIndex int) bool {
	saveDir := filepath.Join(r.tempDir, uploadID)
	finalPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, chunkIndex))
	_, err := os.Stat(finalPath)
	return err == nil
}

func (r *fileUploadRepository) SetUploadedChunks(uploadID, filename string, merged int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.mergedChunks == nil {
		r.mergedChunks = make(map[string]int)
		fmt.Printf("DEBUG: mergedChunks map oluşturuldu\n")
	}
	cleanUploadID := strings.TrimPrefix(uploadID, "upload-")
	key := cleanUploadID + "_" + filename
	r.mergedChunks[key] = merged

	fmt.Printf("DEBUG: SET - Key: %s, Chunk sayısı: %d\n",
		key, merged)
}

func (r *fileUploadRepository) GetUploadedChunks(uploadID, filename string) (int, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	cleanUploadID := strings.TrimPrefix(uploadID, "upload-") // upload_id'ye erişmek için upload- prefixini kaldırıyoruz
	key := cleanUploadID + "_" + filename

	fmt.Printf("DEBUG: GET - Looking for key: %s\n", key)

	if r.mergedChunks == nil {
		return 0, false
	}

	merged, exists := r.mergedChunks[key]
	fmt.Printf("DEBUG: GET - Found: %t, Chunk sayısı: %d\n", exists, merged)
	return merged, exists
}

func (r *fileUploadRepository) CleanupTempFiles(uploadID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	saveDir := filepath.Join(r.tempDir, uploadID)

	err := os.RemoveAll(saveDir)
	if err != nil {
		log.Printf("CleanupTempFiles error for %s: %v", saveDir, err)
		return err
	}
	return nil
}

func (r *fileUploadRepository) MergeChunks(uploadID, filename string, totalChunks int) error {
	saveDir := filepath.Join(r.tempDir, uploadID)
	finalFileName := fmt.Sprintf("%s_%s", uploadID, filename)
	finalPath := filepath.Join(r.uploadsDir, finalFileName)

	fmt.Printf("DEBUG: Merging to %s\n", finalPath) // Debug log

	// Eksik chunk kontrolü
	missing := make([]int, 0)
	for i := 1; i <= totalChunks; i++ {
		if !r.ChunkExists(uploadID, filename, i) {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("eksik chunk(lar) var: %v", missing)
	}

	// uploads klasörünü oluştur
	if err := os.MkdirAll(r.uploadsDir, os.ModePerm); err != nil {
		return fmt.Errorf("uploads klasörü oluşturulamadı: %w", err)
	}

	// Eğer dosya zaten varsa, eski olanı backup al
	if _, err := os.Stat(finalPath); err == nil {
		backupPath := finalPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
		if renameErr := os.Rename(finalPath, backupPath); renameErr != nil {
			log.Printf("UYARI: Eski dosya backup alınamadı: %v", renameErr)
		} else {
			fmt.Printf("DEBUG: Eski dosya backup alındı: %s\n", backupPath)
		}
	}

	outFile, err := os.Create(finalPath)
	if err != nil {
		if cleanupErr := r.CleanupTempFiles(uploadID); cleanupErr != nil {
			log.Printf("UYARI! Temp klasörü temizlenemedi: %v", cleanupErr)
		}
		return fmt.Errorf("final dosya oluşturulamadı: %w", err)
	}
	defer outFile.Close()

	merged := make([]int, 0, totalChunks)

	// chunkları birleştir
	for i := 1; i <= totalChunks; i++ {
		partPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, i))
		partFile, err := os.Open(partPath)
		if err != nil {
			if cleanupErr := r.CleanupTempFiles(uploadID); cleanupErr != nil {
				log.Printf("UYARI! Temp klasörü temizlenemedi: %v", cleanupErr)
			}
			//return fmt.Errorf("chunk %d açılamadı: %w", i, err)
			break //merge işlemi başarısız olan chunkta dur
		}

		bytesWritten, err := io.Copy(outFile, partFile)
		if err != nil {
			partFile.Close()
			if cleanupErr := r.CleanupTempFiles(uploadID); cleanupErr != nil {
				log.Printf("UYARI! Temp klasörü temizlenemedi: %v", cleanupErr)
			}
			//return fmt.Errorf("chunk %d kopyalanamadı: %w", i, err)
			break
		}
		partFile.Close()
		merged = append(merged, i)
		fmt.Printf("DEBUG: Chunk %d merged, %d bytes\n", i, bytesWritten)
	}
	// Başarılı chunk'ları kaydet
	r.SetUploadedChunks(uploadID, filename, len(merged))

	// Dosya boyutunu kontrol et
	stat, _ := outFile.Stat()
	fmt.Printf("DEBUG: Final file size: %d bytes\n", stat.Size())

	return nil
}
