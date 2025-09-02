package usecases

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/processor"
	"file-uploader/internal/infrastructure/queue"
	consts "file-uploader/pkg/constants"
	"file-uploader/pkg/errors"
	"file-uploader/pkg/file"
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
				log.Printf("ERROR: handleMergeSuccess hata verdi %s: %v", filename, err)
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
	if file.IsImageFile(mergedFilePath) {
		//return s.processImageFile(uploadID, filename, mergedFilePath)
		return processor.ProcessImageFile(s.mediaService, filename, mergedFilePath)
	}
	if file.IsVideoFile(mergedFilePath) {
		//return s.processVideoFile(uploadID, filename, mergedFilePath)
		return processor.ProcessVideoFile(s.mediaService, filename, mergedFilePath)
	}
	log.Printf("INFO: image olmayan bir dosya yüklendi: %s", filename)
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
