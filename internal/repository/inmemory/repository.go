// Package inmemory provide realization of repository for runtime in memory storage.
package inmemory

import (
	"context"
	"sync"

	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

// LinksRepositoryMemory - simple in memory storage.
type LinksRepositoryMemory struct {
	Shorts map[string]*model.LinkData
	Users  []model.User
	mu     sync.RWMutex
}

// NewInMemoryLinksRepository - provide new instance InMemoryLinksRepository
// Creates capacity based on config.
func NewInMemoryLinksRepository() *LinksRepositoryMemory {
	return &LinksRepositoryMemory{
		Shorts: make(map[string]*model.LinkData, dictionary.MapSize),
		Users:  make([]model.User, dictionary.MapSize),
	}
}

// SetLinks - writes linkData pointer to internal InMemoryLinksRepository map Shorts.
func (r *LinksRepositoryMemory) SetLinks(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, linkData := range links {
		if _, ok := r.Shorts[linkData.ShortURL]; ok {
			continue
		}

		r.Shorts[linkData.ShortURL] = &linkData
	}

	return links, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *LinksRepositoryMemory) FindShort(_ context.Context, short string) (model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linkData, ok := r.Shorts[short]
	if !ok {
		return model.LinkData{}, repository.ErrNotFoundLink
	}

	return *linkData, nil
}

// PingDB returns nil every time.
func (r *LinksRepositoryMemory) PingDB(_ context.Context) error {
	return nil
}

// SelectUserLinks selects user links by userID.
func (r *LinksRepositoryMemory) SelectUserLinks(_ context.Context, userID model.UserID) ([]model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lds := make([]model.LinkData, 0)

	for _, linkData := range r.Shorts {
		if linkData.UserID == userID {
			lds = append(lds, *linkData)
		}
	}

	return lds, nil
}

// CreateUser creates user.
func (r *LinksRepositoryMemory) CreateUser(_ context.Context, userID model.UserID) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user := model.User{ID: userID}
	r.Users = append(r.Users, user)

	return user, nil
}

// BatchMarkAsDeleted not implemented.
func (r *LinksRepositoryMemory) BatchMarkAsDeleted(_ context.Context, _ chan model.LinkData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}
