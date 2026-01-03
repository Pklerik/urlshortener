// Package service provide all business logic for applications.
package service

import (
	"context"

	"errors"

	"github.com/Pklerik/urlshortener/internal/model"
)

var (
	// ErrEmptyLongURL - error for empty short url.
	ErrEmptyLongURL = errors.New("ShortURL is empty")
	// ErrCollision - sets error if shortURL existed for different long.
	ErrCollision = errors.New("collision for url in db")
	// ErrEmptySecret - secret not found.
	ErrEmptySecret = errors.New("empty secret")
)

// LinkServicer provide service contract for link handling.
type LinkServicer interface {
	RegisterLinks(ctx context.Context, longURLs []string, userID model.UserID) ([]model.LinkData, error)
	GetShort(ctx context.Context, shortURL string) (model.LinkData, error)
	ProvideUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error)
	MarkAsDeleted(ctx context.Context, userID model.UserID, shortLinks model.ShortUrls) error
	PingDB(ctx context.Context) error
	GetSecret(name string) (any, bool)
	GetStats(ctx context.Context) (model.Stats, error)
}
