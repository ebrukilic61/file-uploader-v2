package repositories

import "file-uploader/internal/domain/dto"

/*
type MediaRepository interface {
	SaveMedia(media dto.Media) error
	GetMedia(mediaID string) (*dto.Media, error)

	// Jobs
	CreateJob(mediaID string, job dto.MediaJob) error
	GetJob(jobID string) (*dto.MediaJob, error)
	UpdateJobStatus(jobID string, status string) error
	UpdateJobOutput(jobID string, outputPath string, width, height int) error
	ListJobs(mediaID string) ([]dto.MediaJob, error)
}
*/

//* Single Responsibility Principle (SRP): Her repo sadece bir tabloya odaklanıyor, yönetimi ve test edilmesi kolay

type MediaRepository interface {
	CreateMedia(media *dto.ImageDTO) error
	GetMediaByID(id string) (*dto.ImageDTO, error)
	UpdateMediaStatus(id string, status string) error
}

type MediaVariantRepository interface {
	CreateVariant(variant *dto.MediaVariant) error
	GetVariantByID(id string) (*dto.MediaVariant, error)
	GetMediaVariantByID(id string) (*dto.MediaVariant, error)
	UpdateVariant(variant *dto.MediaVariant) error
	DeleteVariant(id string) error
}

type MediaSizeRepository interface {
	CreateSize(size *dto.MediaSize) error
	GetSizeByName(name string) (*dto.MediaSize, error)
	UpdateSize(size *dto.MediaSize) error
	DeleteSize(name string) error
}
