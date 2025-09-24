package inmemory

import (
	"context"
	"sync"

	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

// InMemoryLinksRepository - simple in memory storage.
type InMemoryLinksRepository struct {
	Shorts map[string]*model.LinkData
	mu     sync.RWMutex
}

// NewInMemoryLinksRepository - provide new instance InMemoryLinksRepository
// Creates capacity based on config.
func NewInMemoryLinksRepository() *InMemoryLinksRepository {
	return &InMemoryLinksRepository{
		Shorts: make(map[string]*model.LinkData, dictionary.MapSize),
	}
}

// Create - writes linkData pointer to internal InMemoryLinksRepository map Shorts.
func (r *InMemoryLinksRepository) Create(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, linkData := range links {
		if _, ok := r.Shorts[linkData.ShortURL]; ok {
			continue
		}

		r.Shorts[linkData.ShortURL] = &linkData
		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
	}

	return links, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *InMemoryLinksRepository) FindShort(_ context.Context, short string) (model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linkData, ok := r.Shorts[short]
	if !ok {
		return model.LinkData{}, repository.ErrNotFoundLink
	}

	return *linkData, nil
}

// PingDB returns nil every time.
func (r *InMemoryLinksRepository) PingDB(_ context.Context) error {
	return nil
}

func (r *InMemoryLinksRepository) AllUserURLs(ctx context.Context, userID string) ([]model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return []model.LinkData{}, nil
}
