// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
)

var (
	// ErrNotFoundLink - link was not found.
	ErrNotFoundLink = errors.New("link was not found")
	// ErrExistingURL - can't crate record with existing shortURL.
	ErrExistingURL = errors.New("can't crate record with existing shortURL")
	// ErrEmptyDatabaseDSN - DatabaseDSN is empty.
	ErrEmptyDatabaseDSN = errors.New("DatabaseDSN is empty")
)

// LinksStorager - interface for shortener service.
type LinksStorager interface {
	Create(ctx context.Context, linkData model.LinkData) (model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
	PingDB(ctx context.Context, args config.StartupFlagsParser) error
}

// InMemoryLinksRepository - simple in memory storage.
type InMemoryLinksRepository struct {
	Shorts map[string]*model.LinkData
	mu     sync.RWMutex
}

// NewInMemoryLinksRepository - provide new instance InMemoryLinksRepository
// Creates capacity based on config.
func NewInMemoryLinksRepository() *InMemoryLinksRepository {
	return &InMemoryLinksRepository{
		Shorts: make(map[string]*model.LinkData, config.MapSize),
	}
}

// Create - writes linkData pointer to internal InMemoryLinksRepository map Shorts.
func (r *InMemoryLinksRepository) Create(_ context.Context, linkData model.LinkData) (model.LinkData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ld, ok := r.Shorts[linkData.ShortURL]; ok {
		return *ld, ErrExistingURL
	}

	r.Shorts[linkData.ShortURL] = &linkData

	return linkData, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *InMemoryLinksRepository) FindShort(_ context.Context, short string) (model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	linkData, ok := r.Shorts[short]
	if !ok {
		return model.LinkData{}, ErrNotFoundLink
	}

	return *linkData, nil
}

// Ping returns ping info from db.
func (r *InMemoryLinksRepository) PingDB(ctx context.Context, args config.StartupFlagsParser) error {
	return nil
}

// LocalMemoryLinksRepository - simple in memory storage.
type LocalMemoryLinksRepository struct {
	File string
	mu   sync.RWMutex
}

// NewLocalMemoryLinksRepository - provide new instance LocalMemoryLinksRepository
// Creates capacity based on config.
func NewLocalMemoryLinksRepository(filePath string) *LocalMemoryLinksRepository {
	basePath := config.BasePath
	if !strings.HasPrefix(filePath, "/") {
		filePath = filepath.Join(basePath, filePath)
	}

	filePath = filepath.Clean(filePath)

	_, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		logger.Sugar.Fatalf("error creating storage file: %w", err)
	}

	return &LocalMemoryLinksRepository{
		File: filePath,
	}
}

// Create - writes linkData pointer to internal LocalMemoryLinksRepository map Shorts.
func (r *LocalMemoryLinksRepository) Create(_ context.Context, linkData model.LinkData) (model.LinkData, error) {
	slStorage, err := r.Read()
	if err != nil {
		return model.LinkData{}, fmt.Errorf("unable to crate link: %w", err)
	}

	ld, ok := slContains(linkData.ShortURL, slStorage)
	if ok {
		return ld, nil
	}

	slStorage = append(slStorage, linkData)
	if err := r.Write(slStorage); err != nil {
		return linkData, fmt.Errorf("unable to crate record: %w", err)
	}

	return linkData, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *LocalMemoryLinksRepository) FindShort(_ context.Context, short string) (model.LinkData, error) {
	slStorage, err := r.Read()
	if err != nil {
		return model.LinkData{}, fmt.Errorf("unable to find link: %w", err)
	}

	ld, ok := slContains(short, slStorage)
	if ok {
		return ld, nil
	}

	return model.LinkData{}, ErrNotFoundLink
}

func (r *LocalMemoryLinksRepository) Read() ([]model.LinkData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slByte, err := os.ReadFile(r.File)
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("unable to open file: %w", err)
	}

	var slStorage = make([]model.LinkData, 0)
	if len(slByte) == 0 {
		return slStorage, nil
	}

	err = json.Unmarshal(slByte, &slStorage)
	if err != nil {
		return slStorage, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	return slStorage, nil
}

func (r *LocalMemoryLinksRepository) Write(data []model.LinkData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bytesData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("unable Marshal data: %w", err)
	}

	err = os.WriteFile(r.File, bytesData, 0600)
	if err != nil {
		return fmt.Errorf("unable Write data to file: %w", err)
	}

	return nil
}

func slContains(shortURL string, slLinkData []model.LinkData) (model.LinkData, bool) {
	for _, linkInfo := range slLinkData {
		if linkInfo.ShortURL == shortURL {
			return linkInfo, true
		}
	}

	return model.LinkData{}, false
}

// PingDB returns ping info from db.
func (r *LocalMemoryLinksRepository) PingDB(ctx context.Context, args config.StartupFlagsParser) error {
	ps := args.GetDatabaseDSN()
	if ps == "" {
		return ErrEmptyDatabaseDSN
	}
	db, err := sql.Open("pgx", args.GetDatabaseDSN())
	if err != nil {
		return fmt.Errorf("unable to connect to DB: %w", err)
	}
	defer db.Close()
	return nil
}
