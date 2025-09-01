package repositories

import (
	fl "file-uploader/pkg/file"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	//"strings"
	"sync"
	"time"
)

type fileUploadRepository struct {
	tempDir      string
	uploadsDir   string
	mergedChunks map[string]int // key: uploadID + filename, value: merged chunk sayısı
	chunkMutex   sync.RWMutex
	fileMutex    sync.Mutex
	activeOps    map[string]int
	opsMutex     sync.Mutex
}

func (r *fileUploadRepository) UploadsDir() string {
	return r.uploadsDir
}

func NewFileUploadRepository(tempDir, uploadsDir string) *fileUploadRepository {
	return &fileUploadRepository{
		tempDir:      tempDir,
		uploadsDir:   uploadsDir,
		mergedChunks: make(map[string]int),
		activeOps:    make(map[string]int),
	}
}

func (r *fileUploadRepository) incrementActiveOps(uploadID string) {
	r.opsMutex.Lock()
	defer r.opsMutex.Unlock()
	r.activeOps[uploadID]++
}

func (r *fileUploadRepository) decrementActiveOps(uploadID string) {
	r.opsMutex.Lock()
	defer r.opsMutex.Unlock()
	r.activeOps[uploadID]--
	if r.activeOps[uploadID] <= 0 {
		delete(r.activeOps, uploadID)
	}
}

func (r *fileUploadRepository) getActiveOps(uploadID string) int {
	r.opsMutex.Lock()
	defer r.opsMutex.Unlock()
	return r.activeOps[uploadID]
}

func (r *fileUploadRepository) SaveChunk(uploadID, filename string, chunkIndex int, file multipart.File) error {
	r.incrementActiveOps(uploadID)
	defer r.decrementActiveOps(uploadID)

	r.fileMutex.Lock()
	defer r.fileMutex.Unlock()

	saveDir := filepath.Join(r.tempDir, uploadID)
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("geçici klasör oluşturulamadı: %w", err)
	}

	finalPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, chunkIndex))
	tmpPath := fmt.Sprintf("%s.tmp.%d", finalPath, time.Now().UnixNano())

	if r.ChunkExists(uploadID, filename, chunkIndex) {
		return nil
	}

	// Dosyayı geçici konuma kaydetmek için:
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("geçici dosya oluşturulamadı: %w", err)
	}
	//defer tmpFile.Close()

	var fileClosed bool
	defer func() {
		if !fileClosed {
			tmpFile.Close()
		}
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			log.Printf("UYARI: Tmp dosya silinemedi: %s, error: %v", tmpPath, err)
		}
	}()

	if _, err := io.Copy(tmpFile, file); err != nil {
		//		os.Remove(tmpPath)
		return fmt.Errorf("chunk kaydedilemedi: %w", err)
	}

	// Dosyayı kapat (rename öncesi)
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("dosya kapatılamadı: %w", err)
	}
	fileClosed = true

	// Atomic rename
	/*
			// Fallback: copy + remove
		if r.ChunkExists(uploadID, filename, chunkIndex) {
			os.Remove(tmpPath)
			return nil // Chunk zaten var
		}

	*/
	if err := os.Rename(tmpPath, finalPath); err != nil {
		if copyErr := fl.CopyFile(tmpPath, finalPath); copyErr != nil {
			//		os.Remove(tmpPath)
			return fmt.Errorf("chunk yazılamadı: %w", copyErr)
		}
		//		os.Remove(tmpPath)
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
	r.chunkMutex.Lock()
	defer r.chunkMutex.Unlock()

	if r.mergedChunks == nil {
		r.mergedChunks = make(map[string]int)
		fmt.Printf("DEBUG: mergedChunks map oluşturuldu\n")
	}

	key := fl.MakeKey(uploadID, filename)
	r.mergedChunks[key] = merged

	fmt.Printf("DEBUG: SET - Key: %s, Chunk sayısı: %d\n", key, merged)
}

func (r *fileUploadRepository) GetUploadedChunks(uploadID, filename string) (int, bool) {
	r.chunkMutex.RLock()
	defer r.chunkMutex.RUnlock()

	key := fl.MakeKey(uploadID, filename)
	fmt.Printf("DEBUG: GET - Looking for key: %s\n", key)

	if r.mergedChunks == nil {
		return 0, false
	}

	merged, exists := r.mergedChunks[key]
	fmt.Printf("DEBUG: GET - Found: %t, Chunk sayısı: %d\n", exists, merged)
	return merged, exists
}

func (r *fileUploadRepository) CleanupTempFiles(uploadID string) error {
	maxWait := 50 // 50 * 100ms = 5 saniye
	for i := 0; i < maxWait; i++ {
		if r.getActiveOps(uploadID) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	r.fileMutex.Lock()
	defer r.fileMutex.Unlock()

	saveDir := filepath.Join(r.tempDir, uploadID)

	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return nil
	}

	// Retry ile silme işlemi
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		if err := os.RemoveAll(saveDir); err != nil {
			lastErr = err
			log.Printf("CleanupTempFiles attempt %d failed for %s: %v",
				attempt+1, saveDir, err)
			continue
		}

		// Başarılı, chunk tracking'i temizle
		r.cleanupChunkTracking(uploadID)
		return nil
	}

	return fmt.Errorf("cleanup başarısız (3 deneme): %w", lastErr)
}

func (r *fileUploadRepository) cleanupChunkTracking(uploadID string) {
	r.chunkMutex.Lock()
	defer r.chunkMutex.Unlock()

	// Bu uploadID ile ilgili tüm kayıtları sil
	for key := range r.mergedChunks {
		if len(key) > len(uploadID) && key[:len(uploadID)] == uploadID {
			delete(r.mergedChunks, key)
		}
	}
}

func (r *fileUploadRepository) MergeChunks(uploadID, filename string, totalChunks int) (string, error) {
	r.incrementActiveOps(uploadID)
	defer r.decrementActiveOps(uploadID)

	r.fileMutex.Lock()
	defer r.fileMutex.Unlock()

	saveDir := filepath.Join(r.tempDir, uploadID)
	finalFileName := fl.MakeKey(uploadID, filename)
	finalPath := filepath.Join(r.uploadsDir, "media", "original", finalFileName)
	//finalPath := filepath.Join("uploads", "media", "original", originalFileName)

	fmt.Printf("DEBUG: Merging to %s\n", finalPath) // Debug log

	// Eksik chunk kontrolü
	missing := make([]int, 0)
	for i := 1; i <= totalChunks; i++ {
		if !r.ChunkExists(uploadID, filename, i) {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("eksik chunk(lar) var: %v", missing) // Return boş string ve error
	}

	// uploads klasörünü oluştur
	if err := os.MkdirAll(r.uploadsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("uploads klasörü oluşturulamadı: %w", err) // Return boş string ve error
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
		/*
			if cleanupErr := r.CleanupTempFiles(uploadID); cleanupErr != nil {
				log.Printf("UYARI! Temp klasörü temizlenemedi: %v", cleanupErr)
			}
		*/
		return "", fmt.Errorf("final dosya oluşturulamadı: %w", err) // Return boş string ve error
	}
	defer outFile.Close()

	merged := make([]int, 0, totalChunks)

	// chunkları birleştir
	for i := 1; i <= totalChunks; i++ {
		partPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, i))
		func() {
			partFile, err := os.Open(partPath)
			if err != nil {
				log.Printf("UYARI: Chunk %d açılamadı: %v", i, err)
				return
			}
			defer partFile.Close()

			bytesWritten, err := io.Copy(outFile, partFile)
			if err != nil {
				log.Printf("UYARI: Chunk %d kopyalanamadı: %v", i, err)
				return
			}

			merged = append(merged, i)
			fmt.Printf("DEBUG: Chunk %d merged, %d bytes\n", i, bytesWritten)
		}()
	}

	// Tüm chunk'ların başarıyla merge edilip edilmediğini kontrol et
	if len(merged) != totalChunks {
		return "", fmt.Errorf("merge işlemi başarısız: %d/%d chunk merge edildi", len(merged), totalChunks)
	}

	// Başarılı chunk'ları kaydet
	r.SetUploadedChunks(uploadID, filename, len(merged))

	// Dosya boyutunu kontrol et
	if stat, err := outFile.Stat(); err == nil {
		fmt.Printf("DEBUG: Final file size: %d bytes\n", stat.Size())
	}

	r.cleanupChunkFiles(saveDir, filename, totalChunks)

	return finalPath, nil
}

// Helper function: Chunk dosyalarını temizle
func (r *fileUploadRepository) cleanupChunkFiles(saveDir, filename string, totalChunks int) {
	for i := 1; i <= totalChunks; i++ {
		partPath := filepath.Join(saveDir, fmt.Sprintf("%s.part%d", filename, i))
		if err := os.Remove(partPath); err != nil {
			log.Printf("UYARI: Chunk dosyası silinemedi %s: %v", partPath, err)
		}
	}

	// Eğer upload dizini boşsa onu da sil
	if entries, err := os.ReadDir(saveDir); err == nil && len(entries) == 0 {
		if err := os.Remove(saveDir); err != nil {
			log.Printf("UYARI: Upload dizini silinemedi %s: %v", saveDir, err)
		}
	}
}

func (r *fileUploadRepository) TempDir() string {
	return r.tempDir
}
