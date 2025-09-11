// Package service provide all business logic for applications.
package service

import (
	"context"

	//nolint
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/samborkent/uuidv7"
)

var (
	// ErrEmptyLongURL - error for empty short url.
	ErrEmptyLongURL = errors.New("ShortURL is empty")
	// ErrCollision - sets error if shortURL existed for different long.
	ErrCollision = errors.New("collision for url in db")
)

// LinkServicer provide service contract for link handling.
type LinkServicer interface {
	RegisterLink(ctx context.Context, longURL string) (model.LinkData, error)
	GetShort(ctx context.Context, shortURL string) (model.LinkData, error)
	PingDB(ctx context.Context, args config.StartupFlagsParser) error
}

// BaseLinkService - structure for service repository realization.
type BaseLinkService struct {
	linksRepo repository.LinksStorager
}

// NewLinksService - provide instance of service.
func NewLinksService(repo repository.LinksStorager) *BaseLinkService {
	return &BaseLinkService{linksRepo: repo}
}

// RegisterLink - register the Link with provided longURL.
func (ls *BaseLinkService) RegisterLink(ctx context.Context, longURL string) (model.LinkData, error) {
	shortURL, err := ls.cutURL(ctx, longURL)
	if err != nil {
		return model.LinkData{}, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}

	err = ls.checkCollision(ctx, shortURL, longURL)
	if err != nil {
		return model.LinkData{}, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}

	ld, err := ls.linksRepo.Create(ctx, model.LinkData{UUID: model.LinkUUIDv7(uuidv7.New().String()), ShortURL: shortURL, LongURL: longURL})

	if err != nil && !errors.Is(err, repository.ErrExistingURL) {
		return ld, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}

	logger.Sugar.Infof("Short url: %s sets for long: %s", ld.ShortURL, ld.LongURL)

	return ld, nil
}

// cutURL - provide shortURl based on Long.
func (ls *BaseLinkService) cutURL(_ context.Context, longURL string) (string, error) {
	//nolint
	h := sha256.New()

	_, err := io.WriteString(h, longURL)
	if err != nil {
		return "", fmt.Errorf("(ls *BaseLinkService) cutURL: %w", err)
	}

	shortURL := fmt.Sprintf("%x", h.Sum(nil))[:8]

	return shortURL, nil
}

// checkCollision makes sure that link doesn't have long representation already.
func (ls *BaseLinkService) checkCollision(ctx context.Context, shortURL, longURL string) error {
	ld, err := ls.linksRepo.FindShort(ctx, shortURL)

	if errors.Is(err, repository.ErrNotFoundLink) {
		return nil
	}

	if err != nil && err != repository.ErrExistingURL {
		return fmt.Errorf("(ls *BaseLinkService) collisionCheck: %w", err)
	}

	if ld.LongURL != longURL {
		return fmt.Errorf("(ls *BaseLinkService) collisionCheck: %w", ErrCollision)
	}

	return nil
}

// GetShort - provide model.LinkData and error
// If shortURL is absent returns err.
func (ls *BaseLinkService) GetShort(ctx context.Context, shortURL string) (model.LinkData, error) {
	ld, err := ls.linksRepo.FindShort(ctx, shortURL)
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) GetShort: %w", err)
	}

	return ld, nil
}

// PingDB - provide error if DB is not accessed.
func (ls *BaseLinkService) PingDB(ctx context.Context, args config.StartupFlagsParser) error {
	if err := ls.linksRepo.PingDB(ctx, args); err != nil {
		return fmt.Errorf("PingDB error: %w", err)
	}

	return nil
}
