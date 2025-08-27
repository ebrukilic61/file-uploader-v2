package errors

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if ue, ok := err.(*UploadError); ok {
		// Orijinal hatayı logla (debug için)
		if ue.Err != nil {
			log.Printf("Upload error [%s]: %v", ue.Code, ue.Err)
		}

		// Status kodunu seç
		var status int
		switch ue.Code {
		case "not_found":
			status = fiber.StatusNotFound
		case "chunk_not_open", "invalid_chunk":
			status = fiber.StatusBadRequest
		default:
			status = fiber.StatusInternalServerError
		}

		// Client’a sadece Code + Message gönder
		return c.Status(status).JSON(fiber.Map{
			"error":   ue.Code,
			"message": ue.Message,
		})
	}

	// Yakalanmayan hatalar için fallback
	log.Printf("Unexpected error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "internal_error",
		"message": "Sunucu hatası",
	})
}
