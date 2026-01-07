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

// LinksRepository - interface for shortener service.
type LinksRepository interface {
	SetLinks(ctx context.Context, links []model.LinkData) ([]model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
	SelectUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error)
	BatchMarkAsDeleted(ctx context.Context, links chan model.LinkData) error
	CreateUser(ctx context.Context, userID model.UserID) (model.User, error)
	PingDB(ctx context.Context) error
	GetStats(ctx context.Context) (model.Stats, error)
}
