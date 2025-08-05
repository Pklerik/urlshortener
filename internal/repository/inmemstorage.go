package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/model"
)

var (
	// ErrNotFoundLink - link was not found.
	ErrNotFoundLink = errors.New("link was not found")
	// ErrExistingURL - can't crate record with existing shortURL.
	ErrExistingURL = errors.New("can't crate record with existing shortURL")
)

// InMemoryLinksRepository - simple in memory storage.
type InMemoryLinksRepository struct {
	Shorts map[string]*model.LinkData
	Longs  map[string]*model.LinkData
	mu     sync.RWMutex
}

// NewInMemoryLinksRepository - provide new instance InMemoryLinksRepository
// Creates capacity based on config.
func NewInMemoryLinksRepository() *InMemoryLinksRepository {
	return &InMemoryLinksRepository{
		Shorts: make(map[string]*model.LinkData, config.MapSize),
		Longs:  make(map[string]*model.LinkData, config.MapSize),
	}
}

// Create - writes linkData pointer to internal InMemoryLinksRepository maps (Shorts, Longs).
func (r *InMemoryLinksRepository) Create(_ context.Context, linkData model.LinkData) (model.LinkData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Shorts[linkData.ShortURL] = &linkData
	r.Longs[linkData.LongURL] = &linkData

	return linkData, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *InMemoryLinksRepository) FindShort(_ context.Context, short string) (model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linkData, ok := r.Shorts[short]
	if !ok {
		return model.LinkData{}, ErrNotFoundLink
	}

	return *linkData, nil
}

// FindLong - provide model.LinkData and error
// If longURL is absent returns ErrNotFoundLink.
func (r *InMemoryLinksRepository) FindLong(_ context.Context, long string) (model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linkData, ok := r.Longs[long]
	if !ok {
		return model.LinkData{}, ErrNotFoundLink
	}

	return *linkData, nil
}
