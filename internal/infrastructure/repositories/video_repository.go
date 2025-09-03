package repositories

import (
	"file-uploader/internal/domain/dto"
	"file-uploader/internal/domain/entities"
	"file-uploader/pkg/constants"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) CreateVideo(video *dto.VideoDTO) error { //buna db işlemleri için gerek var
	if video.VideoID == "" {
		video.VideoID = uuid.New().String()
	}
	entity := entities.Video{
		VideoID:      uuid.MustParse(video.VideoID),
		Width:        video.Width,
		Height:       video.Height,
		Status:       video.Status,
		OriginalName: video.OriginalName,
		FileType:     video.FileType,
		FilePath:     video.FilePath,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return r.db.Create(&entity).Error
}

func (r *VideoRepository) GetVideoByID(id string) (*dto.VideoDTO, error) {
	var entity entities.Video
	if err := r.db.First(&entity, "video_id = ?", id).Error; err != nil {
		return nil, err
	}

	video := dto.VideoDTO{
		VideoID:      entity.VideoID.String(),
		OriginalName: entity.OriginalName,
		FileType:     entity.FileType,
		FilePath:     entity.FilePath,
		Status:       entity.Status,
		Width:        entity.Width,
		Height:       entity.Height,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
	return &video, nil

}

func (r *VideoRepository) ResizeWidth(video *entities.Video) error { //video boyutlarını güncellemek için
	var existingVideo entities.Video
	if err := r.db.First(&existingVideo, "video_id = ?", video.VideoID).Error; err != nil {
		return err
	}
	existingVideo.Width = video.Width
	existingVideo.Height = video.Height
	existingVideo.Status = video.Status
	existingVideo.FilePath = video.FilePath
	return r.db.Save(&existingVideo).Error
}

func (r *VideoRepository) ResizeHeight(video *entities.Video) error {
	var existingVideo entities.Video
	if err := r.db.First(&existingVideo, "video_id = ?", video.VideoID).Error; err != nil {
		return err
	}
	existingVideo.Height = video.Height
	existingVideo.Width = video.Width
	existingVideo.Status = video.Status
	existingVideo.FilePath = video.FilePath
	return r.db.Save(&existingVideo).Error
}

func (r *VideoRepository) ResizeVideo(video *entities.Video) error {
	var existingVideo entities.Video
	if err := r.db.First(&existingVideo, "video_id = ?", video.VideoID).Error; err != nil {
		return err
	}
	existingVideo.Width = video.Width
	existingVideo.Height = video.Height
	existingVideo.Status = constants.VideoStatusResized
	existingVideo.FilePath = video.FilePath
	return r.db.Save(&existingVideo).Error
}
