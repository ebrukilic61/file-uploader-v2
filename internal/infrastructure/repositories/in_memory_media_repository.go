package repositories

import (
	"errors"
	"file-uploader/internal/domain/dto"
	"fmt"
	"sync"
	"time"
)

type InMemoryMediaRepository struct {
	mu   sync.RWMutex
	data map[string]*dto.Media
}

func NewInMemoryMediaRepository() *InMemoryMediaRepository {
	return &InMemoryMediaRepository{
		data: make(map[string]*dto.Media),
	}
}

func (r *InMemoryMediaRepository) Create(media *dto.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[media.MediaID] = media
	return nil
}

func (r *InMemoryMediaRepository) ListMedia(filter dto.MediaFilter) ([]dto.Media, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]dto.Media, 0)
	for _, media := range r.data {
		// Tip filtresi varsa
		if filter.Type != "" && media.FileType != filter.Type {
			continue
		}
		// Kategori filtresi varsa
		if filter.Category != "" && media.Category != filter.Category {
			continue
		}
		result = append(result, *media)
	}

	return result, nil
}

func (r *InMemoryMediaRepository) GetByID(mediaID string) (*dto.Media, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	media, exists := r.data[mediaID]
	if !exists {
		return nil, fmt.Errorf("media not found")
	}
	return media, nil
}

func (r *InMemoryMediaRepository) GetAll() ([]dto.Media, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	medias := make([]dto.Media, 0, len(r.data))
	for _, media := range r.data {
		medias = append(medias, *media)
	}
	return medias, nil
}

func (r *InMemoryMediaRepository) Delete(mediaID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[mediaID]; !ok {
		return errors.New("media not found")
	}
	delete(r.data, mediaID)
	return nil
}

func (r *InMemoryMediaRepository) UpdateMetadata(mediaID string, metadata dto.Metadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.data[mediaID]; ok {
		m.Metadata = metadata
		m.UpdatedAt = time.Now() // metadata değiştiğinde updated_at güncelle
		return nil
	}
	return errors.New("media not found")
}
