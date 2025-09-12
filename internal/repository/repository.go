// Package repository provide abstract implementation for data storage of model/model.go struct
package repository

import (
	"context"
	"database/sql"
	"slices"

	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	_ "github.com/jackc/pgx/v5/stdlib" // import driver for "database/sql"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/migrations"
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
	Create(ctx context.Context, links []model.LinkData) ([]model.LinkData, error)
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
		Shorts: make(map[string]*model.LinkData, dictionary.MapSize),
	}
}

// Create - writes linkData pointer to internal InMemoryLinksRepository map Shorts.
func (r *InMemoryLinksRepository) Create(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, linkData := range links {
		if _, ok := r.Shorts[linkData.ShortURL]; ok {
			continue
		}

		r.Shorts[linkData.ShortURL] = &linkData
		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
	}

	return links, nil
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

// PingDB returns nil every time.
func (r *InMemoryLinksRepository) PingDB(_ context.Context, _ config.StartupFlagsParser) error {
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
	basePath := dictionary.BasePath
	if !strings.HasPrefix(filePath, "/") {
		filePath = filepath.Join(basePath, filePath)
	}

	filePath = filepath.Clean(filePath)

	_, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		logger.Sugar.Fatalf("error creating storage file: %w", err)
	}

	logger.Sugar.Info("Creating file by path: %s", filePath)

	return &LocalMemoryLinksRepository{
		File: filePath,
	}
}

// Create - writes linkData pointer to internal LocalMemoryLinksRepository map Shorts.
func (r *LocalMemoryLinksRepository) Create(_ context.Context, links []model.LinkData) ([]model.LinkData, error) {
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

	if err := r.Write(slStorage); err != nil {
		logger.Sugar.Errorf("unable to crate record: %w", err)
		return links, fmt.Errorf("unable to crate record: %w", err)
	}

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
func (r *LocalMemoryLinksRepository) PingDB(_ context.Context, _ config.StartupFlagsParser) error {
	return nil
}

// DBLinksRepository provide base struct for db implementation.
type DBLinksRepository struct {
	db *sql.DB
}

// NewDBLinksRepository - provide new instance DBLinksRepository.
func NewDBLinksRepository(ctx context.Context, parsedArgs config.StartupFlagsParser) *DBLinksRepository {
	db, err := ConnectDB(parsedArgs)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	logger.Sugar.Infof("SUCCESS connecting to db: %v", db.Stats())

	err = migrations.MakeMigrations(ctx, db, parsedArgs.GetDatabaseConf())
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	return &DBLinksRepository{
		db: db,
	}
}

// ConnectDB connecting to DB.
func ConnectDB(parsedArgs config.StartupFlagsParser) (*sql.DB, error) {
	if os.Getenv("GOOSE_DRIVER") == "" {
		if err := os.Setenv("GOOSE_DRIVER", dbconf.DefaultGooseDrier); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

	if os.Getenv("GOOSE_DBSTRING") == "" {
		if err := os.Setenv("GOOSE_DBSTRING", parsedArgs.GetDatabaseConf().GetConnString()); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

	if os.Getenv("GOOSE_MIGRATION_DIR") == "" {
		dir := filepath.Join(dictionary.BasePath, "migrations")
		if err := os.Setenv("GOOSE_MIGRATION_DIR", dir); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

	dbConf := parsedArgs.GetDatabaseConf()
	logger.Sugar.Infof("ConnString: Database: %s, User: %s, Options: %v",
		dbConf.(*dbconf.Conf).Database,
		dbConf.(*dbconf.Conf).User,
		dbConf.(*dbconf.Conf).Options,
	)

	if dbConf == nil {
		return nil, ErrCollectingDBConf
	}

	ps := dbConf.GetConnString()

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to DB: %w", err)
	}

	return db, nil
}

// Create - writes linkData pointer to internal DBLinksRepository map Shorts.
func (r *DBLinksRepository) Create(ctx context.Context, links []model.LinkData) ([]model.LinkData, error) {
	var (
		err         error
		newLinksIDs = make([]int, 0, len(links))
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return links, fmt.Errorf("error creating tx error: %w", err)
	}

	for i, linkData := range links {
		existingLD, err := r.getShort(ctx, tx, linkData.ShortURL)

		if err != nil {
			return links, fmt.Errorf("error Create: %w", err)
		}

		if existingLD == nil {
			newLinksIDs = append(newLinksIDs, i)
		}
	}

	if len(newLinksIDs) == 0 {
		return links, ErrExistingLink
	}

	for i, linkData := range links {
		if !slices.Contains(newLinksIDs, i) {
			continue
		}

		res, err := tx.ExecContext(ctx, "INSERT INTO links (id, short_url, long_url) VALUES($1, $2, $3)", linkData.UUID, linkData.ShortURL, linkData.LongURL)
		if err != nil {
			return links, fmt.Errorf("error inserting data to db: %w", err)
		}

		if rows, err := res.RowsAffected(); err != nil || rows != 1 {
			return links, fmt.Errorf("error parsing result of insertion data to db: %w", err)
		}

		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
	}

	if err := tx.Commit(); err != nil {
		return links, fmt.Errorf("committing insertion error: %w", err)
	}

	if len(newLinksIDs) != len(links) {
		return links, ErrExistingLink
	}
	return links, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns nil.
func (r *DBLinksRepository) FindShort(ctx context.Context, short string) (model.LinkData, error) {
	var (
		ld  = new(model.LinkData)
		err error
	)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return *ld, fmt.Errorf("error creating tx error: %w", err)
	}

	ld, err = r.getShort(ctx, tx, short)
	if err != nil {
		return *ld, fmt.Errorf("crate error: %w", err)
	}

	return *ld, nil
}

// PingDB returns nil every time.
func (r *DBLinksRepository) PingDB(_ context.Context, args config.StartupFlagsParser) error {
	_, err := ConnectDB(args)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	return nil
}

func (r *DBLinksRepository) getShort(ctx context.Context, tx *sql.Tx, short string) (*model.LinkData, error) {
	linkData := model.LinkData{}

	row := tx.QueryRowContext(ctx, "SELECT id, short_url, long_url FROM links WHERE short_url LIKE $1", short)

	err := row.Scan(&linkData.UUID, &linkData.ShortURL, &linkData.LongURL)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Sugar.Errorf("error selecting db data: %w", err)

			return nil, fmt.Errorf("error selecting db data: %w", err)
		}

		return nil, nil
	}

	logger.Sugar.Infof("LD selected: %s", linkData)

	return &linkData, nil
}
