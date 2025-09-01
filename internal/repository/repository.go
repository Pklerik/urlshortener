// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
)

var (
	// ErrNotFoundLink - link was not found.
	ErrNotFoundLink = errors.New("link was not found")
	// ErrExistingURL - can't crate record with existing shortURL.
	ErrExistingURL = errors.New("can't crate record with existing shortURL")
)

// LinksStorager - interface for shortener service.
type LinksStorager interface {
	Create(ctx context.Context, linkData model.LinkData) (model.LinkData, error)
	FindShort(ctx context.Context, short string) (model.LinkData, error)
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

// LocalMemoryLinksRepository - simple in memory storage.
type LocalMemoryLinksRepository struct {
	File string
	mu   sync.RWMutex
}

// NewLocalMemoryLinksRepository - provide new instance LocalMemoryLinksRepository
// Creates capacity based on config.
func NewLocalMemoryLinksRepository(filePath string) *LocalMemoryLinksRepository {
	_, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Sugar.Fatalf("error creating storage file: %w", err)
	}
	return &LocalMemoryLinksRepository{
		File: filePath,
	}
}

type LocalMemoryProducer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewLocalMemoryProducer(filename string) (*LocalMemoryProducer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &LocalMemoryProducer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *LocalMemoryProducer) WriteEvent(event *model.LinkData) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type LocalMemoryConsumer struct {
	file *os.File
	// добавляем reader в Consumer
	reader *bufio.Reader
}

func NewLocalMemoryConsumer(filename string) (*LocalMemoryConsumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &LocalMemoryConsumer{
		file: file,
		// создаём новый Reader
		reader: bufio.NewReader(file),
	}, nil
}

func (c *LocalMemoryConsumer) ReadEvent() (*model.LinkData, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// преобразуем данные из JSON-представления в структуру
	event := model.LinkData{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// Create - writes linkData pointer to internal LocalMemoryLinksRepository map Shorts.
func (r *LocalMemoryLinksRepository) Create(_ context.Context, linkData model.LinkData) (model.LinkData, error) {
	slStorage, err := r.Read()
	for _, linkInfo := range slStorage {
		if linkInfo.ShortURL == linkData.ShortURL {
			return linkInfo, nil
		}
	}
	if err != nil {
		return model.LinkData{}, fmt.Errorf("unable to finde link: %w", err)
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
	for _, linkInfo := range slStorage {
		if linkInfo.ShortURL == short {
			return linkInfo, nil
		}
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
	var slStorage []model.LinkData = make([]model.LinkData, 0)
	// buf := make([]byte, 0)
	// fileStream.Read(buf)
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
	err = os.WriteFile(r.File, bytesData, 0666)
	if err != nil {
		return fmt.Errorf("unable Write data to file: %w", err)
	}
	return nil
}
