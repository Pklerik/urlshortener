// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"context"

	"github.com/Pklerik/urlshortener/internal/model"
)

// LinksStorager - interface for shortener service.
type LinksStorager interface {
	Create(ctx context.Context, linkData model.LinkData) (model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
}
