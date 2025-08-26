package handlers

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/usecases"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	service usecases.MediaService
}

func NewMediaHandler(service usecases.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
}

// RegisterMedia
//
// @Summary      Register Media
// @Description  Registers a new media file
// @Tags         Media
// @Accept       json
// @Produce      json
// @Param        filename   query     string true "File name"
// @Param        fileSize   query     string true "File size"
// @Param        fileType   query     string true "File type"
// @Param        filePath   query     string true "File path"
// @Success      200       {object}  dto.MediaRegisterRequestDTO
// @Failure      400       {object}  dto.ErrorResponse "Missing parameter"
// @Failure      500       {object}  dto.ErrorResponse "Internal server error"
// @Router       /media/register [post]
func (h *MediaHandler) RegisterMedia(c *fiber.Ctx) error {
	fileSizeStr := c.FormValue("file_size")
	fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid file_size"})
	}
	req := dto.MediaRegisterRequestDTO{
		Filename: c.FormValue("filename"),
		FileType: c.FormValue("file_type"),
		FileSize: fileSize,
		FilePath: c.FormValue("file_path"),
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	media, err := h.service.RegisterMedia(req.Filename, req.FileType, req.FileSize, req.FilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(media)
}

// GetMedia
//
// @Summary      Get Media
// @Description  Retrieves a media file by its ID
// @Tags         Media
// @Accept       json
// @Produce      json
// @Param        id   path     string true "Media ID"
// @Success      200  {object} dto.MediaResponseDTO
// @Failure      404  {object} dto.ErrorResponse "Media not found"
// @Failure      500  {object} dto.ErrorResponse "Internal server error"
// @Router       /media/{id} [get]
func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	mediaID := c.Params("id")
	media, err := h.service.GetMedia(mediaID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(media)
}

func (h *MediaHandler) ListMedia(c *fiber.Ctx) error {
	mediaList, err := h.service.ListMedia()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(mediaList)
}

func (h *MediaHandler) DeleteMedia(c *fiber.Ctx) error {
	mediaID := c.Params("id")
	if err := h.service.DeleteMedia(mediaID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) UpdateMediaMetadata(c *fiber.Ctx) error {
	mediaID := c.Params("id")
	var metadata dto.Metadata
	if err := c.BodyParser(&metadata); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.service.UpdateMediaMetadata(mediaID, metadata); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
