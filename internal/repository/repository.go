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
	_ "github.com/jackc/pgx/v5/stdlib" // import driver for "database/sql"

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
	PingDB(ctx context.Context) error
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
func (r *InMemoryLinksRepository) PingDB(_ context.Context) error {
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
func (r *LocalMemoryLinksRepository) PingDB(_ context.Context) error {
	return nil
}

// DBLinksRepository provide base struct for db implementation.
type DBLinksRepository struct {
	db *sql.DB
}

// NewDBLinksRepository - provide new instance DBLinksRepository.
func NewDBLinksRepository(ctx context.Context, dbConf dbconf.DBConfigurer) *DBLinksRepository {
	db, err := ConnectDB(dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	logger.Sugar.Infof("SUCCESS connecting to db: %v", db.Stats())

	err = migrations.MakeMigrations(ctx, db, dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
		return nil
	}

	return &DBLinksRepository{
		db: db,
	}
}

// ConnectDB connecting to DB.
func ConnectDB(dbConf dbconf.DBConfigurer) (*sql.DB, error) {
	if os.Getenv("GOOSE_DRIVER") == "" {
		if err := os.Setenv("GOOSE_DRIVER", dbconf.DefaultGooseDrier); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

	if os.Getenv("GOOSE_DBSTRING") == "" {
		if err := os.Setenv("GOOSE_DBSTRING", dbConf.GetConnString()); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

	if os.Getenv("GOOSE_MIGRATION_DIR") == "" {
		dir := filepath.Join(dictionary.BasePath, "migrations")
		if err := os.Setenv("GOOSE_MIGRATION_DIR", dir); err != nil {
			return nil, fmt.Errorf("cant set env variable: %w", err)
		}
	}

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
		err error
	)

	insertedLinks, err := r.insertBatch(ctx, links)
	if err != nil {
		return links, fmt.Errorf("links insertion error: %w", err)
	}

	if len(insertedLinks) != len(links) {
		return links, ErrExistingLink
	}

	return insertedLinks, nil
}

func (r *DBLinksRepository) insertBatch(ctx context.Context, links []model.LinkData) ([]model.LinkData, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return links, fmt.Errorf("error creating tx error: %w", err)
	}
	defer tx.Rollback()

	query, queryArgs := prepareInsertionQuery(links)

	rows, err := tx.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("error inserting link data: %w", err)
	}

	linksData, err := collectInsertedLinks(rows, len(links))
	if err != nil {
		return linksData, fmt.Errorf("error collecting insertion results: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return links, fmt.Errorf("committing insertion error: %w", err)
	}

	return linksData, nil
}

func prepareInsertionQuery(links []model.LinkData) (string, []interface{}) {
	var (
		queryArgs    = make([]interface{}, 0, 3*len(links))
		placeholders = make([]string, 0, 3*len(links))
	)
	for pgi, linkData := range links {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", pgi*3+1, pgi*3+2, pgi*3+3))
		queryArgs = append(queryArgs, linkData.UUID, linkData.ShortURL, linkData.LongURL)
		logger.Sugar.Infof("Short url: %s sets for long: %s", linkData.ShortURL, linkData.LongURL)
	}

	return fmt.Sprintf("INSERT INTO links (id, short_url, long_url) VALUES %s ON CONFLICT (short_url) DO NOTHING RETURNING id, short_url, long_url", strings.Join(placeholders, ", ")), queryArgs
}

func collectInsertedLinks(rows *sql.Rows, lenLinks int) ([]model.LinkData, error) {
	linksData := make([]model.LinkData, 0, lenLinks)
	for rows.Next() {
		linkData := model.LinkData{}

		err := rows.Scan(&linkData.UUID, &linkData.ShortURL, &linkData.LongURL)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Sugar.Errorf("error scanning INSERT result: %w", err)
				return linksData, fmt.Errorf("error scanning INSERT result: %w", err)
			}
		}

		linksData = append(linksData, linkData)
	}

	return linksData, nil
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
	defer tx.Rollback()

	ld, err = r.getShort(ctx, tx, short)
	if err != nil {
		return *ld, fmt.Errorf("crate error: %w", err)
	}

	return *ld, nil
}

// PingDB returns nil every time.
func (r *DBLinksRepository) PingDB(_ context.Context) error {
	err := r.db.Ping()
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
