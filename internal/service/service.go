// Package service provide all business logic for applications.
package service

import (
	"context"
	//nolint
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

var (
	// ErrEmptyLongURL - error for empty short url.
	ErrEmptyLongURL = errors.New("ShortURL is empty")
	// ErrCollision - sets error if shortURL existed for different long.
	ErrCollision = errors.New("collision for url in db")
)

type LinkServicer interface {
	RegisterLink(ctx context.Context, longURL string) (model.LinkData, error)
	GetShort(ctx context.Context, shortURL string) (model.LinkData, error)
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
	var shortURL string

	//nolint
	h := md5.New()

	_, err := io.WriteString(h, longURL)
	if err != nil {
		return model.LinkData{}, fmt.Errorf("RegisterLink longURL hash writing: %w", err)
	}

	shortURL = fmt.Sprintf("%x", h.Sum(nil))[:8]

	ld, err := ls.linksRepo.FindShort(ctx, shortURL)
	if err == nil {
		if ld.LongURL != longURL {
			return ld, ErrCollision
		}

		log.Printf("Short url: %s sets for long: %s", ld.ShortURL, ld.LongURL)

		return ld, nil
	}

	ld, err = ls.linksRepo.Create(ctx, model.LinkData{ShortURL: shortURL, LongURL: longURL})
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}

	return ld, nil
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
