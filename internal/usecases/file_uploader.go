package usecases

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/queue"
	consts "file-uploader/pkg/constants"
	"file-uploader/pkg/errors"
	fl "file-uploader/pkg/file"
)

type UploadService interface {
	GetUploadStatus(req *dto.UploadStatusRequestDTO) (*dto.UploadStatusResponse, error)
	UploadChunk(req *dto.UploadChunkRequestDTO, fileHeader *multipart.FileHeader) (*dto.UploadChunkResponse, error)
	CompleteUpload(req *dto.CompleteUploadRequestDTO) (*dto.CompleteUploadResponse, error)
	CancelUpload(req *dto.CancelUploadRequestDTO) (*dto.CancelUploadResponse, error)
	Shutdown() // worker pool'u kapatmak için
}

type uploadService struct {
	repo         repositories.FileUploadRepository
	storage      repositories.StorageStrategy
	mu           sync.Mutex
	workerPool   *queue.WorkerPool
	mediaService MediaService
}

func NewUploadService(repo repositories.FileUploadRepository, storage repositories.StorageStrategy, mediaService MediaService) UploadService {
	workerCount := 5
	if val, ok := os.LookupEnv("WORKER_POOL_SIZE"); ok {
		if wc, err := strconv.Atoi(val); err == nil {
			workerCount = wc
		}
	}
	workerPool := queue.NewWorkerPool(workerCount, repo) // 5 worker ile başlatalım
	return &uploadService{
		repo:         repo,
		storage:      storage,
		mu:           sync.Mutex{}, //sonradan ekledim
		workerPool:   workerPool,
		mediaService: mediaService,
	}
}

func (s *uploadService) GetUploadStatus(req *dto.UploadStatusRequestDTO) (*dto.UploadStatusResponse, error) {
	// Repository'den merge edilen chunk sayısını al
	mergedChunkCount, exists := s.repo.GetUploadedChunks(req.UploadID, req.Filename)

	var uploadedChunks int
	var uploadedStatus string
	if exists && mergedChunkCount > 0 {
		uploadedChunks = mergedChunkCount
		uploadedStatus = consts.StatusCompleted
	} else {
		uploadedChunks = 0
		uploadedStatus = consts.StatusFailed
	}

	response := &dto.UploadStatusResponse{
		UploadID:       req.UploadID,
		Filename:       req.Filename,
		UploadedChunks: uploadedChunks,
		Status:         uploadedStatus,
	}

	return response, nil
}

func (s *uploadService) UploadChunk(req *dto.UploadChunkRequestDTO, fileHeader *multipart.FileHeader) (*dto.UploadChunkResponse, error) {
	safeFilename := filepath.Base(req.Filename)

	// Chunk index doğrulama
	idx, err := strconv.Atoi(req.ChunkIndex)
	if err != nil || idx <= 0 {
		return nil, errors.ErrInvalidChunk(err)
	}

	// Idempotent kontrol
	if s.repo.ChunkExists(req.UploadID, safeFilename, idx) {
		return &dto.UploadChunkResponse{
			Status:     consts.StatusOK,
			UploadID:   req.UploadID,
			ChunkIndex: idx,
			Filename:   safeFilename,
			Message:    "chunk zaten var",
		}, nil
	}

	// Dosyayı aç
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.ErrFileCantOpen(err)
	}
	defer file.Close()

	// Hash doğrulama (eğer gerekiyorsa)
	if req.ChunkHash != "" {
		// Geçici olarak kaydet ve hash doğrula
		tempSaveErr := s.repo.SaveChunk(req.UploadID, safeFilename, idx, file)
		if tempSaveErr != nil {
			// Temp cleanup (isteğe bağlı)
			s.repo.CleanupTempFiles(req.UploadID)
			return nil, errors.ErrTmpFile(tempSaveErr)
		}
		file.Close()

		// Hash doğrulama için dosya yolu:
		chunkPath := filepath.Join("temp_uploads", req.UploadID, fmt.Sprintf("%s.part%d", safeFilename, idx))

		if err := fl.ValidateFileHash(chunkPath, req.ChunkHash); err != nil {
			s.repo.CleanupTempFiles(req.UploadID)
			log.Printf("WARN: Temp siliniyor: %v", err)
			return nil, err
		}

		return &dto.UploadChunkResponse{
			Status:     consts.StatusOK,
			UploadID:   req.UploadID,
			ChunkIndex: idx,
			Filename:   safeFilename,
		}, nil
	} else {
		// Hash doğrulama yoksa direkt kaydet
		if err := s.repo.SaveChunk(req.UploadID, safeFilename, idx, file); err != nil {
			s.repo.CleanupTempFiles(req.UploadID)
			return nil, errors.ErrTmpFile(err)
		}
	}

	chunkJob := queue.Job{
		UploadID:   req.UploadID,
		Type:       queue.JobSaveChunk,
		Filename:   safeFilename,
		ChunkIndex: idx,
		File:       file,
	}

	s.workerPool.AddJob(chunkJob)

	return &dto.UploadChunkResponse{
		Status:     consts.StatusQueued,
		UploadID:   req.UploadID,
		ChunkIndex: idx,
		Filename:   safeFilename,
		Message:    "chunk işleme kuyruğuna alındı",
	}, nil
}

func (s *uploadService) CompleteUpload(req *dto.CompleteUploadRequestDTO) (*dto.CompleteUploadResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	safeFilename := filepath.Base(req.Filename)

	mergeJob := queue.Job{
		UploadID:    req.UploadID,
		Type:        queue.JobMerge,
		Filename:    safeFilename,
		TotalChunks: req.TotalChunks,
		// Merge tamamlandığında çağrılacak callback
		OnMergeSuccess: func(uploadID, filename, mergedFilePath string) {
			if err := s.handleMergeSuccess(uploadID, filename, mergedFilePath); err != nil {
				log.Printf("ERROR: handleMergeSuccess failed for %s: %v", filename, err)
			}
		},
	}

	s.workerPool.AddJob(mergeJob)

	return &dto.CompleteUploadResponse{
		Status:   consts.StatusQueued,
		Message:  "Chunked dosyalar işleme kuyruğuna alındı",
		Filename: req.Filename,
	}, nil
}

func (s *uploadService) handleMergeSuccess(uploadID, filename, mergedFilePath string) error {
	if s.isImageFile(mergedFilePath) {
		return s.processImageFile(uploadID, filename, mergedFilePath)
	}
	log.Printf("INFO: Non-image file uploaded: %s", filename)
	return nil
}

func (s *uploadService) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif"}

	for _, imgExt := range imageExtensions {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func (s *uploadService) getMimeTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// Image dosyasını işle ve media service'e gönder
func (s *uploadService) processImageFile(uploadID, filename, mergedFilePath string) error {
	// Merge edilmiş dosyayı aç
	file, err := os.Open(mergedFilePath)
	if err != nil {
		return fmt.Errorf("failed to open merged file: %w", err)
	}
	defer file.Close()

	// ImageDTO oluştur
	imageDTO := &dto.ImageDTO{
		OriginalName: filename,
		FileType:     s.getMimeTypeFromExtension(filename),
		Status:       "processing",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Media service'e gönder
	if err := s.mediaService.CreateMedia(imageDTO, file); err != nil {
		return fmt.Errorf("media oluşturulamadı: %w", err)
	}

	// Başarılı olduğunda temp dosyayı sil
	if err := os.Remove(mergedFilePath); err != nil {
		log.Printf("WARN: temp file kaldırılamadı %s: %v", mergedFilePath, err)
	}

	log.Printf("INFO: Image %s başarıyla işlendi ve %s konumuna kaydedildi", filename, imageDTO.FilePath)
	return nil
}

func (s *uploadService) CancelUpload(req *dto.CancelUploadRequestDTO) (*dto.CancelUploadResponse, error) {
	//* complete upload ile race condition yaşamaması adına lock eklendi, aksi takdirde cleanup işlemi tetiklenmiyordu, çünkü sıra ona gelmiyordu
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanupJob := queue.Job{
		UploadID: req.UploadID,
		Type:     queue.JobCleanup,
	}

	s.workerPool.AddJob(cleanupJob)

	return &dto.CancelUploadResponse{
		Status:  consts.StatusQueued,
		Message: "Upload iptal edildi",
	}, nil
}

// Shutdown worker pool
func (s *uploadService) Shutdown() {
	if s.workerPool != nil {
		s.workerPool.Shutdown()
	}
}
