package usecases

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"sync"

	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/repositories"
	"file-uploader/internal/infrastructure/processor"
	"file-uploader/internal/infrastructure/queue"
	consts "file-uploader/pkg/constants"
	"file-uploader/pkg/errors"
	"file-uploader/pkg/helper"

	"github.com/go-redis/redis/v8"
)

type UploadService interface {
	GetUploadStatus(req *dto.UploadStatusRequestDTO) (*dto.UploadStatusResponse, error)
	UploadChunk(req *dto.UploadChunkRequestDTO, fileHeader *multipart.FileHeader) (*dto.UploadChunkResponse, error)
	CompleteUpload(req *dto.CompleteUploadRequestDTO) (*dto.CompleteUploadResponse, error)
	CancelUpload(req *dto.CancelUploadRequestDTO) (*dto.CancelUploadResponse, error)
	HandleMergeSuccess(uploadID, filename, mergedFilePath string, totalChunks int) error
	RetryMerge(uploadID, filename string) (string, error)
}

type uploadService struct { //* sadece jobları kuyruğa atacak
	repo         repositories.FileUploadRepository
	storage      repositories.StorageStrategy
	mu           sync.Mutex
	rdb          *redis.Client
	mediaService MediaService
}

func NewUploadService(repo repositories.FileUploadRepository, storage repositories.StorageStrategy, rdb *redis.Client, mediaService MediaService) UploadService {
	return &uploadService{
		repo:         repo,
		storage:      storage,
		mu:           sync.Mutex{}, //sonradan ekledim
		rdb:          rdb,
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.ErrTmpFile(err)
	}

	chunkJob := queue.Job{
		UploadID:    req.UploadID,
		Type:        queue.JobSaveChunk,
		Filename:    safeFilename,
		ChunkIndex:  idx,
		FileContent: fileBytes,
		ChunkHash:   req.ChunkHash,
	}

	serialized, err := json.Marshal(chunkJob)
	if err != nil {
		log.Println("Failed to serialize chunk job:", err)
		return nil, err
	}
	s.rdb.LPush(context.Background(), "job_queue", serialized)

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
	}

	serialized, err := json.Marshal(mergeJob)
	if err != nil {
		log.Printf("Merge job marshal failed: %v", err)
	} else {
		log.Printf("Merge job serialized: %s", string(serialized))
	}
	s.rdb.LPush(context.Background(), "job_queue", serialized)

	return &dto.CompleteUploadResponse{
		Status:   consts.StatusQueued,
		Message:  "Chunked dosyalar işleme kuyruğuna alındı",
		Filename: req.Filename,
	}, nil
}

func (s *uploadService) HandleMergeSuccess(uploadID, filename, mergedFilePath string, totalChunks int) error {
	s.repo.SetUploadedChunks(uploadID, filename, totalChunks) //* status failed olarak gözüküyordu, bunu düzeltmek adına merge success'in başarılı olma durumunda status set edildi
	if helper.IsImageFile(mergedFilePath) {
		return processor.ProcessImageFile(s.mediaService, filename, mergedFilePath)
	}
	if helper.IsVideoFile(mergedFilePath) {
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

	serialized, _ := json.Marshal(cleanupJob)
	s.rdb.LPush(context.Background(), "job_queue", serialized)

	return &dto.CancelUploadResponse{
		Status:  consts.StatusQueued,
		Message: "Upload iptal edildi",
	}, nil
}

func (s *uploadService) RetryMerge(uploadID, filename string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	finalPath, merged, err := s.repo.RetryMerge(uploadID, filename)
	if err != nil {
		log.Printf("Retry merge işlemi dosya %s için başarısız oldu: %v", filename, err)
		return "", err
	}
	log.Printf("Retry merge işlemi dosya %s için başarıyla gerçekleşti: %s", filename, finalPath)

	if err := s.HandleMergeSuccess(uploadID, filename, finalPath, merged); err != nil {
		log.Printf("Merge'den sonra verinin işlenmesi başarıyla gerçekleşti %s: %v", filename, err)
		return finalPath, err
	}

	if err := s.repo.UpdateRetryStatus(uploadID, consts.StatusUploaded); err != nil {
		log.Printf("Retry job status güncellenirken hata oluştu: %v", err)
		return finalPath, err
	}

	return finalPath, nil
}
