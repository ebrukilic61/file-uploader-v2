package mapper

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
)

func MediaToDTO(m *entities.ImageDTO) dto.ImageDTO {
	return dto.ImageDTO{
		ID:           m.ID.String(),
		OriginalName: m.OriginalName,
		FileType:     m.FileType,
		FilePath:     m.FilePath,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
