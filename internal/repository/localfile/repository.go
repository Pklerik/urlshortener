package localfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

// LocalMemoryLinksRepository - simple in memory storage.
type LocalMemoryLinksRepository struct {
	File string
	mu   sync.RWMutex
}

// NewLocalMemoryLinksRepository - provide new instance LocalMemoryLinksRepository
// Creates capacity based on config.
func NewLocalMemoryLinksRepository(filePath string) *LocalMemoryLinksRepository {
	logger.Sugar.Info("args file path: ", filePath)

	basePath := dictionary.BasePath
	logger.Sugar.Info("Set base path: ", basePath)

	fullPath := filePath
	if !strings.HasPrefix(filePath, "/") {
		fullPath = filepath.Clean(filepath.Join(basePath, filePath))
	}

	logger.Sugar.Info("Set full path: ", fullPath)

	_, err := os.OpenFile(fullPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Sugar.Fatalf("error creating storage file:%v : %w", fullPath, err)
	}

	logger.Sugar.Info("Creating file by path: ", fullPath)

	return &LocalMemoryLinksRepository{
		File: fullPath,
	}
}

// Create - writes linkData pointer to internal LocalMemoryLinksRepository map Shorts.
func (r *LocalMemoryLinksRepository) SetLinks(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
	slStorage, err := r.Read()
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("unable to crate link: %w", err)
	}

	for _, linkData := range links {
		_, ok := slContains(linkData.ShortURL, slStorage)
		if ok {
			continue
		}

		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
		slStorage = append(slStorage, linkData)
	}
	// асинхронная запись в фаил.
	go func() {
		if err := r.Write(slStorage); err != nil {
			logger.Sugar.Errorf("unable to save file: %w", err)
		}
	}()

	return links, nil
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

	return model.LinkData{}, repository.ErrNotFoundLink
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
func (r *LocalMemoryLinksRepository) PingDB(_ context.Context) error {
	return nil
}

func (r *LocalMemoryLinksRepository) AllUserURLs(ctx context.Context, userID string) ([]model.LinkData, error) {
	return []model.LinkData{}, nil
}

func (r *LocalMemoryLinksRepository) CreateUser(ctx context.Context) (string, error) {
	return "", nil
}
