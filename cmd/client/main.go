package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const LIMIT = 5

type UploadProgress struct {
	mu          sync.RWMutex
	totalChunks int
	uploaded    int
	failed      int
	isCancelled bool
	startTime   time.Time
}

func (up *UploadProgress) IncrementUploaded() {
	up.mu.Lock()
	defer up.mu.Unlock()
	up.uploaded++
}

func (up *UploadProgress) IncrementFailed() {
	up.mu.Lock()
	defer up.mu.Unlock()
	up.failed++
}

func (up *UploadProgress) SetCancelled() {
	up.mu.Lock()
	defer up.mu.Unlock()
	up.isCancelled = true
}

func (up *UploadProgress) IsCancelled() bool {
	up.mu.RLock()
	defer up.mu.RUnlock()
	return up.isCancelled
}

func (up *UploadProgress) GetProgress() (uploaded, failed, total int) {
	up.mu.RLock()
	defer up.mu.RUnlock()
	return up.uploaded, up.failed, up.totalChunks
}

// Gerçekten benzersiz ID üretir
func generateUniqueUploadID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("upload-%d-%x", timestamp, randomBytes)
}

func main() {
	server := flag.String("server", "http://localhost:3000/api/v1", "Server base URL")
	filePath := flag.String(
		"file",
		"C:\\Users\\PC_2250__\\Desktop\\veri_4gb\\tears-of-steel-2s.mp4",
		"Yüklenecek dosyanın yolu",
	)
	chunkSize := flag.Int64("chunk-size", 10*1024*1024, "Chunk size in bytes (default 10MB)")
	uploadID := flag.String("upload-id", "", "Upload session ID")
	flag.Parse()

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("Dosya açılamadı: %v\n", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Fatalf("Dosya bilgisi alınamadı: %v\n", err)
	}

	filename := filepath.Base(stat.Name())
	totalSize := stat.Size()
	if *chunkSize <= 0 {
		log.Fatal("chunk-size > 0 olmalı")
	}
	totalChunks := int((totalSize + *chunkSize - 1) / *chunkSize)

	if strings.TrimSpace(*uploadID) == "" {
		*uploadID = generateUniqueUploadID()
	}

	fmt.Printf("Sunucu: %s\n", *server)
	fmt.Printf("Dosya: %s (%d bytes)\n", filename, totalSize)
	fmt.Printf("Upload ID: %s\n", *uploadID)
	fmt.Printf("Chunk size: %d bytes | Total chunks: %d\n", *chunkSize, totalChunks)
	fmt.Println("Ctrl+C ile iptal edebilirsiniz...")

	// İptal sinyalini yakala
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sem := make(chan struct{}, LIMIT)
	var wg sync.WaitGroup
	progress := &UploadProgress{totalChunks: totalChunks, startTime: time.Now()}

	// Progress gösterimi için ayrı goroutine
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				uploaded, failed, total := progress.GetProgress()
				if uploaded+failed > 0 {
					fmt.Printf("\rİlerleme: %d/%d tamamlandı, %d hata", uploaded, total, failed)
				}
			}
		}
	}()

	// Ana yükleme döngüsü
	for i := 1; i <= totalChunks; i++ {
		if progress.IsCancelled() {
			log.Println("\nUpload iptal edildi, chunk gönderme durduruldu.")
			break
		}

		wg.Add(1)
		go func(chunkNum int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if progress.IsCancelled() {
				return
			}

			start := int64(chunkNum-1) * (*chunkSize)
			end := start + (*chunkSize)
			if end > totalSize {
				end = totalSize
			}
			length := end - start

			buf := make([]byte, length)
			if _, err := file.ReadAt(buf, start); err != nil && err != io.EOF {
				log.Printf("\nChunk %d okunamadı: %v\n", chunkNum, err)
				progress.IncrementFailed()
				progress.SetCancelled() // Hata durumunda da iptal et
				return
			}

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			writer.WriteField("upload_id", *uploadID)
			writer.WriteField("chunk_index", fmt.Sprintf("%d", chunkNum))
			writer.WriteField("filename", filename)

			part, err := writer.CreateFormFile("file", filename)
			if err != nil {
				log.Printf("\nForm dosyası oluşturulamadı: %v\n", err)
				progress.IncrementFailed()
				progress.SetCancelled()
				return
			}

			if _, err := part.Write(buf); err != nil {
				log.Printf("\nForm dosyasına yazılamadı: %v\n", err)
				progress.IncrementFailed()
				progress.SetCancelled()
				return
			}
			writer.Close()

			resp, err := http.Post(
				strings.TrimRight(*server, "/")+"/upload/chunk",
				writer.FormDataContentType(),
				&body,
			)
			if err != nil {
				log.Printf("\nChunk %d gönderilemedi: %v\n", chunkNum, err)
				progress.IncrementFailed()
				progress.SetCancelled()
				return
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("\n[%d/%d] HTTP %d %s\n", chunkNum, totalChunks, resp.StatusCode, string(respBody))
				progress.IncrementFailed()
				progress.SetCancelled()
				return
			}

			progress.IncrementUploaded()
		}(i)
	}

	// `wg.Wait()` ve `signal` kanalı için birleşme noktası
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		// Tüm chunk'lar tamamlandı
		done <- true // Progress gösterimini durdur
		uploaded, failed, total := progress.GetProgress()
		fmt.Printf("\nUpload tamamlandı: %d/%d başarılı, %d hata\n", uploaded, total, failed)

		if failed > 0 {
			log.Println("Hata nedeniyle upload tamamlanamadı")
		} else {
			// Tamamlama isteği sadece başarılı olursa gönderilir
			fmt.Println("Dosya birleştiriliyor...")
			var completeBody bytes.Buffer
			cw := multipart.NewWriter(&completeBody)
			cw.WriteField("upload_id", *uploadID)
			cw.WriteField("total_chunks", fmt.Sprintf("%d", totalChunks))
			cw.WriteField("filename", filename)
			cw.Close()

			resp, err := http.Post(
				strings.TrimRight(*server, "/")+"/upload/complete",
				cw.FormDataContentType(),
				&completeBody,
			)
			if err != nil {
				log.Fatalf("Tamamlama isteği gönderilemedi: %v\n", err)
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("Tamamlama yanıtı: HTTP %d %s\n", resp.StatusCode, string(respBody))
		}

	case <-sigCh:
		// Ctrl+C yakalandı, iptal isteği gönder
		progress.SetCancelled()
		done <- true // Progress gösterimini durdur
		fmt.Println("\nUpload iptal ediliyor...")

		var cancelBody bytes.Buffer
		cw := multipart.NewWriter(&cancelBody)
		cw.WriteField("upload_id", *uploadID)
		cw.Close()

		resp, err := http.Post(
			strings.TrimRight(*server, "/")+"/upload/cancel",
			cw.FormDataContentType(),
			&cancelBody,
		)
		if err != nil {
			log.Fatalf("İptal isteği gönderilemedi: %v\n", err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("İptal yanıtı: HTTP %d %s\n", resp.StatusCode, string(respBody))
		fmt.Println("Upload başarıyla iptal edildi")
	}
}
