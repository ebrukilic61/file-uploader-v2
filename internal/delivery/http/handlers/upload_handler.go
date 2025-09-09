package handlers

import (
	"log"
	"strconv"

	"file-uploader/internal/domain/dto"
	"file-uploader/internal/usecases"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	uploadService usecases.UploadService
	//processingUploads sync.Map
}

func NewUploadHandler(uploadService usecases.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		//processingUploads: sync.Map{},
	}
}

// UploadStatus
//
// @Summary      Get Upload Status
// @Description  Returns the list of already uploaded chunks for a given upload session
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        upload_id  query     string true "Upload ID"
// @Param        filename   query     string true "File name"
// @Success      200       {object}  dto.UploadStatusResponse
// @Failure      400       {object}  dto.ErrorResponse "Missing parameter"
// @Failure      500       {object}  dto.ErrorResponse "Internal server error"
// @Router       /upload/status [get]
func (h *UploadHandler) UploadStatus(c *fiber.Ctx) error {
	req := &dto.UploadStatusRequestDTO{
		UploadID: c.Query("upload_id"),
		Filename: c.Query("filename"),
	}

	if req.UploadID == "" || req.Filename == "" {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: "Eksik parametre",
		})
	}

	response, err := h.uploadService.GetUploadStatus(req)
	if err != nil {
		return c.Status(500).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(response)
}

// UploadChunk
//
// @Summary      Upload Chunk
// @Description  Uploads a single chunk of a file
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        upload_id    formData  string true "Upload ID"
// @Param        chunk_index  formData  string true "Chunk index"
// @Param        filename     formData  string true "File name"
// @Param        chunk_hash   formData  string false "Chunk hash"
// @Param        file         formData  file   true "Chunk file"
// @Success      200          {object}  dto.UploadChunkResponse
// @Failure      400          {object}  dto.ErrorResponse
// @Router       /upload/chunk [post]
func (h *UploadHandler) UploadChunk(c *fiber.Ctx) error {
	req := &dto.UploadChunkRequestDTO{
		UploadID:   c.FormValue("upload_id"),
		ChunkIndex: c.FormValue("chunk_index"),
		Filename:   c.FormValue("filename"),
		ChunkHash:  c.FormValue("chunk_hash"),
	}

	if req.UploadID == "" || req.ChunkIndex == "" || req.Filename == "" {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: "Eksik parametre",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: "Dosya bulunamadı",
		})
	}

	response, err := h.uploadService.UploadChunk(req, fileHeader)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}
	log.Printf("DEBUG: Received chunk: uploadID=%s, filename=%s, chunkIndex=%s", req.UploadID, req.Filename, req.ChunkIndex)

	return c.JSON(response)
}

// CompleteUpload
//
// @Summary      Complete Upload
// @Description  Marks the upload as complete and merges chunks
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        upload_id     formData  string true "Upload ID"
// @Param        total_chunks  formData  int    true "Total chunks"
// @Param        filename      formData  string true "File name"
// @Success      200           {object}  dto.CompleteUploadResponse
// @Failure      400           {object}  dto.ErrorResponse
// @Router       /upload/complete [post]
func (h *UploadHandler) CompleteUpload(c *fiber.Ctx) error {
	totalChunks, _ := strconv.Atoi(c.FormValue("total_chunks"))

	req := &dto.CompleteUploadRequestDTO{
		UploadID:    c.FormValue("upload_id"),
		TotalChunks: totalChunks,
		Filename:    c.FormValue("filename"),
	}

	if req.UploadID == "" || req.Filename == "" || req.TotalChunks <= 0 {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: "Eksik veya geçersiz parametre",
		})
	}

	response, err := h.uploadService.CompleteUpload(req)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(response)
}

// CancelUpload
//
// @Summary      Cancel Upload
// @Description  Cancels an ongoing upload
// @Tags         Upload
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CancelUploadRequestDTO true "Cancel upload request"
// @Success      200      {object}  dto.CancelUploadResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /upload/cancel [post]
func (h *UploadHandler) CancelUpload(c *fiber.Ctx) error {
	var req dto.CancelUploadRequestDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	resp, err := h.uploadService.CancelUpload(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(resp)
}

func (h *UploadHandler) RetryMerge(c *fiber.Ctx) error {
	var req dto.CompleteRetryRequest

	// BodyParser ile:
	if err := c.BodyParser(&req); err != nil {
		log.Printf("BodyParser error: %v", err)
	}

	// Eğer body boşsa query’den alalım:
	if req.UploadID == "" {
		req.UploadID = c.Query("upload_id")
	}
	if req.Filename == "" {
		req.Filename = c.Query("filename")
	}

	log.Printf("DEBUG: Retry request: %+v", req)

	if req.UploadID == "" || req.Filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "upload_id veya filename eksik",
		})
	}

	finalPath, err := h.uploadService.RetryMerge(req.UploadID, req.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	} else {
		log.Printf("INFO: Retry merge işlemi başarıyla gerçekleşti: %s, dosya: %s", req.UploadID, req.Filename)
	}

	return c.JSON(fiber.Map{
		"merged_file": finalPath,
	})
}
