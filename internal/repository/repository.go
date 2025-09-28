// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"context"

	"errors"

	_ "github.com/jackc/pgx/v5/stdlib" // import driver for "database/sql"

	"github.com/Pklerik/urlshortener/internal/model"
)

var (
	// ErrNotFoundLink - link was not found.
	ErrNotFoundLink = errors.New("link was not found")
	// ErrEmptyDatabaseDSN - DatabaseDSN is empty.
	ErrEmptyDatabaseDSN = errors.New("DatabaseDSN is empty")
	// ErrCollectingDBConf - unable to collect DB conf.
	ErrCollectingDBConf = errors.New("unable to collect DB conf")
	// ErrExistingLink - link already in exist.
	ErrExistingLink = errors.New("link already in exist")
)

// LinksStorager - interface for shortener service.
type LinksStorager interface {
	SetLinks(ctx context.Context, links []model.LinkData) ([]model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
	SelectUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error)
	CreateUser(ctx context.Context) (string, error)
	PingDB(ctx context.Context) error
}
