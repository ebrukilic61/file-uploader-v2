package handlers

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/usecases"

	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	repo usecases.MediaService
}

func NewMediaHandler(repo usecases.MediaService) *MediaHandler {
	return &MediaHandler{repo: repo}
}

func (h *MediaHandler) CreateMedia(c *fiber.Ctx) error {
	var media dto.ImageDTO
	if err := c.BodyParser(&media); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	filepath := c.FormValue("filepath")
	return h.repo.CreateMedia(&media, filepath)
}

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	id := c.Params("id")
	media, err := h.repo.GetMediaByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "media not found"})
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "media not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) GetAllMedia(c *fiber.Ctx) error {
	media, err := h.repo.GetAllMedia()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to retrieve media"})
	}
	return c.JSON(media)
}

func (h *MediaHandler) CreateVariantsForMedia(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.CreateVariantsForMedia(id, c.FormValue("original_path")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create media variants"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) CreateSize(c *fiber.Ctx) error {
	var size dto.MediaSize
	if err := c.BodyParser(&size); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.repo.CreateSize(&size); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create media size"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *MediaHandler) UpdateSize(c *fiber.Ctx) error {
	var size dto.MediaSize
	if err := c.BodyParser(&size); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.repo.UpdateSize(&size); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update media size"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
