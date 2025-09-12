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
	RegisterLinks(ctx context.Context, longURLs []string) ([]model.LinkData, error)
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

// RegisterLinks - register the Link with provided longURL.
func (ls *BaseLinkService) RegisterLinks(ctx context.Context, longURLs []string) ([]model.LinkData, error) {
	logger.Sugar.Infof("Long urls to short: %v", longURLs)
	lds := make([]model.LinkData, 0, len(longURLs))

	for _, longURL := range longURLs {
		shortURL, err := ls.cutURL(ctx, longURL)
		if err != nil {
			return lds, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
		}

		lds = append(lds, model.LinkData{
			UUID:     model.LinkUUIDv7(uuidv7.New().String()),
			ShortURL: shortURL,
			LongURL:  longURL,
		})
	}

	lds, err := ls.linksRepo.Create(ctx, lds)
	if err != nil && !errors.Is(err, repository.ErrExistingLink) {
		return lds, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}
	if errors.Is(err, repository.ErrExistingLink) {
		return lds, repository.ErrExistingLink
	}

	return lds, nil
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
	if err != nil {
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
