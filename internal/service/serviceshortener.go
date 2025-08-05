// Package service provide all business logic for applications.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"

	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

var (
	//ErrEmptyLongURL - error for empty short url
	ErrEmptyLongURL = errors.New("ShortURL is empty")
)

// LinkService - structure for service repository realization.
type LinkService struct {
	linksRepo repository.LinksRepository
}

// NewLinksService - provide instance of service.
func NewLinksService(repo repository.LinksRepository) *LinkService {
	return &LinkService{linksRepo: repo}
}

// RegisterLink - register the Link with provided longURL.
func (ls *LinkService) RegisterLink(ctx context.Context, longURL string) (model.LinkData, error) {
	var shortURL string
	if longURL == "" {
		return model.LinkData{}, ErrEmptyLongURL
	}

	ld, err := ls.linksRepo.FindLong(ctx, longURL)
	if err == nil {
		return ld, nil
	}

	for {
		shortURL = rand.Text()[:10]

		ld, err = ls.linksRepo.FindShort(ctx, shortURL)
		if err != nil {
			break
		}
	}

	log.Printf("Short url: %s sets for long: %s", ld.ShortURL, ld.LongURL)

	ld, err = ls.linksRepo.Create(ctx, model.LinkData{ShortURL: shortURL, LongURL: longURL})
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) RegistaerLink: %w", err)
	}

	return ld, nil
}

// GetShort - provide model.LinkData and error
// If shortURL is absent returns err.
func (ls *LinkService) GetShort(ctx context.Context, shortURL string) (model.LinkData, error) {
	ld, err := ls.linksRepo.FindShort(ctx, shortURL)
	if err != nil {
		return ld, fmt.Errorf("(ls *LinkService) GetShort: %w", err)
	}

	return ld, nil
}
