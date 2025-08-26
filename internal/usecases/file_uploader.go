package usecases

import (
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"sync"

	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/queue"
	"file-uploader/internal/pkg/fileutils"
)

type UploadService interface {
	GetUploadStatus(req *dto.UploadStatusRequestDTO) (*dto.UploadStatusResponse, error)
	UploadChunk(req *dto.UploadChunkRequestDTO, fileHeader *multipart.FileHeader) (*dto.UploadChunkResponse, error)
	CompleteUpload(req *dto.CompleteUploadRequestDTO) (*dto.CompleteUploadResponse, error)
	CancelUpload(req *dto.CancelUploadRequestDTO) (*dto.CancelUploadResponse, error)
	Shutdown() // worker pool'u kapatmak için
}

type uploadService struct {
	repo       repositories.FileUploadRepository
	storage    repositories.StorageStrategy
	mu         sync.Mutex
	workerPool *queue.WorkerPool
}

func NewUploadService(repo repositories.FileUploadRepository, storage repositories.StorageStrategy) UploadService {
	workerPool := queue.NewWorkerPool(5, repo) // 5 worker ile başlatalım
	return &uploadService{
		repo:       repo,
		storage:    storage,
		mu:         sync.Mutex{}, //sonradan ekledim
		workerPool: workerPool,
	}
}

func (s *uploadService) GetUploadStatus(req *dto.UploadStatusRequestDTO) (*dto.UploadStatusResponse, error) {
	// Repository'den merge edilen chunk sayısını al
	mergedChunkCount, exists := s.repo.GetUploadedChunks(req.UploadID, req.Filename)

	var uploadedChunks int
	var uploadedStatus string
	if exists && mergedChunkCount > 0 {
		uploadedChunks = mergedChunkCount
		uploadedStatus = "completed"
	} else {
		uploadedChunks = 0
		uploadedStatus = "failed"
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
		return nil, fmt.Errorf("geçersiz chunk_index")
	}

	// Idempotent kontrol
	if s.repo.ChunkExists(req.UploadID, safeFilename, idx) {
		return &dto.UploadChunkResponse{
			Status:     "ok",
			UploadID:   req.UploadID,
			ChunkIndex: idx,
			Filename:   safeFilename,
			Message:    "chunk zaten var",
		}, nil
	}

	// Dosyayı aç
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("dosya açılamadı: %w", err)
	}
	//defer file.Close()

	// Hash doğrulama (eğer gerekiyorsa)
	if req.ChunkHash != "" {
		// Geçici olarak kaydet ve hash doğrula
		tempSaveErr := s.repo.SaveChunk(req.UploadID, safeFilename, idx, file)
		if tempSaveErr != nil {
			log.Printf("WARN: Temp siliniyor: %v", tempSaveErr)
			// Temp cleanup (isteğe bağlı)
			s.repo.CleanupTempFiles(req.UploadID)
			return nil, tempSaveErr
		}
		file.Close() // Dosyayı kapat, çünkü SaveChunk içinde açılmıştı

		// Hash doğrulama için dosya yolunu oluştur
		chunkPath := filepath.Join("temp_uploads", req.UploadID, fmt.Sprintf("%s.part%d", safeFilename, idx))

		if err := fileutils.ValidateFileHash(chunkPath, req.ChunkHash); err != nil {
			s.repo.CleanupTempFiles(req.UploadID)
			log.Printf("WARN: Temp siliniyor: %v", err)
			return nil, err
		}

		return &dto.UploadChunkResponse{
			Status:     "ok",
			UploadID:   req.UploadID,
			ChunkIndex: idx,
			Filename:   safeFilename,
		}, nil
	} else {
		// Hash doğrulama yoksa direkt kaydet
		if err := s.repo.SaveChunk(req.UploadID, safeFilename, idx, file); err != nil {
			log.Printf("WARN: Temp dosyalar temizlenemedi: %v", err)
			s.repo.CleanupTempFiles(req.UploadID)
			return nil, err
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
		Status:     "queued",
		UploadID:   req.UploadID,
		ChunkIndex: idx,
		Filename:   safeFilename,
		Message:    "chunk işleme kuyruğuna alındı",
	}, nil
}

func (s *uploadService) CompleteUpload(req *dto.CompleteUploadRequestDTO) (*dto.CompleteUploadResponse, error) {
	//* cancel upload ile race condition yaşamaması adına lock eklendi
	s.mu.Lock()
	defer s.mu.Unlock()

	safeFilename := filepath.Base(req.Filename)

	mergeJob := queue.Job{
		UploadID:    req.UploadID,
		Type:        queue.JobMerge,
		Filename:    safeFilename,
		TotalChunks: req.TotalChunks,
	}

	s.workerPool.AddJob(mergeJob)

	/*
		err := s.repo.MergeChunks(req.UploadID, safeFilename, req.TotalChunks)
		if err != nil {
			return nil, err
		}

		if err := s.repo.CleanupTempFiles(req.UploadID); err != nil {
			slog.Warn("Temp dosyası temizlenemedi", "error", err)
		}

		return &dto.CompleteUploadResponse{
			Status:   "ok",
			Message:  "Chunked dosyalar başarıyla birleştirildi",
			Filename: req.Filename,
		}, nil
	*/
	return &dto.CompleteUploadResponse{
		Status:   "queued",
		Message:  "Chunked dosyalar başarıyla birleştirildi",
		Filename: req.Filename,
	}, nil
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

	/*
		if err := s.repo.CleanupTempFiles(req.UploadID); err != nil {
			slog.Warn("Geçici dosyalar temizlenemedi (CancelUpload)", "error", err)
		}

		return &dto.CancelUploadResponse{
			Status:  "ok",
			Message: "Upload iptal edildi",
		}, nil
	*/
	return &dto.CancelUploadResponse{
		Status:  "queued",
		Message: "Upload iptal edildi",
	}, nil
}

// Shutdown worker pool
func (s *uploadService) Shutdown() {
	if s.workerPool != nil {
		s.workerPool.Shutdown()
	}
}
