package handlers

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/usecases"

	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	service usecases.MediaService
}

func NewMediaHandler(service usecases.MediaService) *MediaHandler {
	return &MediaHandler{service: service}
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
	return c.JSON(dto.MediaResponseDTO{
		MediaID:  media.MediaID,
		Filename: media.Filename,
	})
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
