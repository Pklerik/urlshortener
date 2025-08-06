// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"context"

	"github.com/Pklerik/urlshortener/internal/model"
)

// LinksRepository - interface for shortener service.
type LinksRepository interface {
	Create(ctx context.Context, linkData model.LinkData) (model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
}
