package repositories

import (
	"file-uploader/internal/domain/dto"
)

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
