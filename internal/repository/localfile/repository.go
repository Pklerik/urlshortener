// Package localfile provide realization of repository for local file storage.
package localfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-json"

	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
)

// LinksRepositoryFile - simple in memory storage.
type LinksRepositoryFile struct {
	File string
	mu   sync.RWMutex
}

// FullData - all service data.
type FullData struct {
	Links []model.LinkData            `json:"links"`
	Users map[model.UserID]model.User `json:"users"`
}

// NewLocalMemoryLinksRepository - provide new instance LocalMemoryLinksRepository
// Creates capacity based on config.
func NewLocalMemoryLinksRepository(filePath string) *LinksRepositoryFile {
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

	return &LinksRepositoryFile{
		File: fullPath,
	}
}

// SetLinks - writes linkData pointer to internal LocalMemoryLinksRepository map Shorts.
func (r *LinksRepositoryFile) SetLinks(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
	fullData, err := r.Read()
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("unable to crate link: %w", err)
	}

	for _, linkData := range links {
		_, ok := slContains(linkData.ShortURL, fullData.Links)
		if ok {
			continue
		}

		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
		fullData.Links = append(fullData.Links, linkData)
		fullData.Users[linkData.UserID] = model.User{ID: linkData.UserID}
	}
	// асинхронная запись в фаил.
	go func() {
		if err := r.Write(fullData); err != nil {
			logger.Sugar.Errorf("unable to save file: %w", err)
		}
	}()

	return links, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns ErrNotFoundLink.
func (r *LinksRepositoryFile) FindShort(_ context.Context, short string) (model.LinkData, error) {
	data, err := r.Read()
	if err != nil {
		return model.LinkData{}, fmt.Errorf("unable to find link: %w", err)
	}

	ld, ok := slContains(short, data.Links)
	if ok {
		return ld, nil
	}

	return model.LinkData{}, repository.ErrNotFoundLink
}

func (r *LinksRepositoryFile) Read() (FullData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fullData := FullData{
		Users: make(map[model.UserID]model.User, dictionary.MapSize),
	}

	slByte, err := os.ReadFile(r.File)
	if err != nil {
		return fullData, fmt.Errorf("unable to open file: %w", err)
	}

	if len(slByte) == 0 {
		return fullData, nil
	}

	err = json.Unmarshal(slByte, &fullData)
	if err != nil {
		return fullData, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	return fullData, nil
}

func (r *LinksRepositoryFile) Write(data FullData) error {
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
func (r *LinksRepositoryFile) PingDB(_ context.Context) error {
	return nil
}

// SelectUserLinks selects user links by userID.
func (r *LinksRepositoryFile) SelectUserLinks(_ context.Context, userID model.UserID) ([]model.LinkData, error) {
	lds := make([]model.LinkData, 0)

	data, err := r.Read()
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("SelectUserLinks: %w", err)
	}

	for _, linkData := range data.Links {
		if linkData.UserID == userID {
			lds = append(lds, linkData)
		}
	}

	return lds, nil
}

// CreateUser creates user.
func (r *LinksRepositoryFile) CreateUser(_ context.Context, userID model.UserID) (model.User, error) {
	data, err := r.Read()
	if err != nil {
		return model.User{}, fmt.Errorf("SelectUserLinks: %w", err)
	}

	user := model.User{ID: userID}
	data.Users[user.ID] = user

	go func() {
		if err := r.Write(data); err != nil {
			logger.Sugar.Errorf("unable to save file: %w", err)
		}
	}()

	return user, nil
}

// BatchMarkAsDeleted not implemented.
func (r *LinksRepositoryFile) BatchMarkAsDeleted(_ context.Context, _ chan model.LinkData) error {
	return nil
}
