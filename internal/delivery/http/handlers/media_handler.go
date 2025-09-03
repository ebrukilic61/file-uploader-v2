package handlers

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/usecases"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MediaHandler struct {
	repo usecases.MediaService
}

func NewMediaHandler(repo usecases.MediaService) *MediaHandler {
	return &MediaHandler{repo: repo}
}

func (h *MediaHandler) CreateMedia(c *fiber.Ctx) error {
	var media dto.ImageDTO
	id := uuid.New()
	media.ID = id.String()

	file, err := c.FormFile("file")
	media.OriginalName = file.Filename
	media.Status = "processing"

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	savePath := fmt.Sprintf("./uploads/media/original/%s_%s", media.ID, file.Filename)

	if err := os.MkdirAll("./uploads/media/original", os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "belirtilen dosya yolu oluşturulamadı"})
	}

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "dosya kaydedilemedi"})
	}
	media.FilePath = savePath
	media.FileType = file.Header.Get("Content-Type")
	if err := h.repo.CreateMedia(&media, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "media oluşturulamadı"})
	}
	return c.JSON(media)
}

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	id := c.Params("id")
	media, err := h.repo.GetMediaByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "media bulunamadı"})
	}
	return c.JSON(media)
}

func (h *MediaHandler) UpdateMediaStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var status struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&status); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.repo.UpdateMediaStatus(id, status.Status); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "media bulunamadı"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) GetAllMedia(c *fiber.Ctx) error {
	media, err := h.repo.GetAllMedia()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "media alınamadı"})
	}
	return c.JSON(media)
}

func (h *MediaHandler) CreateSize(c *fiber.Ctx) error {
	var size dto.MediaSize
	if err := c.BodyParser(&size); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.repo.CreateSize(&size); err != nil {
		fmt.Println("db insert error: ", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "media boyutu oluşturulamadı"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) UpdateSize(c *fiber.Ctx) error {
	var size dto.MediaSize
	if err := c.BodyParser(&size); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.repo.UpdateSize(&size); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "media boyutu güncellenemedi"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) GetVideoByID(c *fiber.Ctx) error {
	id := c.Params("video_id")
	video, err := h.repo.GetVideoByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "video bulunamadı"})
	}
	return c.JSON(video)
}

func (h *MediaHandler) ResizeByWidth(c *fiber.Ctx) error {
	id := c.Params("video_id")
	var req struct {
		Width int64 `json:"width"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	video, err := h.repo.GetVideoByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "video bulunamadı"})
	}
	if err := h.repo.ResizeByWidth(id, req.Width, video); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "video yeniden boyutlandırılamadı"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) ResizeByHeight(c *fiber.Ctx) error {
	id := c.Params("video_id")

	height, _ := strconv.Atoi(c.Query("height", "0"))

	if height <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "height > 0 olmalı"})
	}

	video, err := h.repo.GetVideoByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "video bulunamadı"})
	}
	if err := h.repo.ResizeByHeight(id, int64(height), video); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "video yeniden boyutlandırılamadı"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

/*
func (h *MediaHandler) ResizeVideo(c *fiber.Ctx) error {
	id := c.Params("video_id")
	var req struct {
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	video, err := h.repo.GetVideoByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "video not found"})
	}
	if err := h.repo.ResizeVideo(id, req.Width, req.Height, video); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resize video"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
*/

func (h *MediaHandler) ResizeVideo(c *fiber.Ctx) error {
	id := c.Params("video_id")

	width, _ := strconv.Atoi(c.Query("width", "0"))
	height, _ := strconv.Atoi(c.Query("height", "0"))

	if width <= 0 || height <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "width and height must be > 0"})
	}

	video, err := h.repo.GetVideoByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "video not found"})
	}

	if err := h.repo.ResizeVideo(id, int64(width), int64(height), video); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) CreateVideo(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "video dosyası yüklemeniz gerekmekte"})
	}

	var video dto.VideoDTO
	videoID := uuid.New()
	video.VideoID = videoID.String()
	video.OriginalName = fileHeader.Filename
	video.Status = "processing"
	video.FileType = fileHeader.Header.Get("Content-Type")

	saveDir := "./uploads/videos/original"

	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create upload directory"})
	}
	savePath := fmt.Sprintf("%s/%s_%s", saveDir, video.VideoID, fileHeader.Filename)
	if err := c.SaveFile(fileHeader, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save video"})
	}
	video.FilePath = savePath

	if err := h.repo.CreateVideo(&video); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(video)
}
