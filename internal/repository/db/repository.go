package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/Pklerik/urlshortener/internal/dictionary"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/repository"
	"github.com/Pklerik/urlshortener/migrations"
)

// DBLinksRepository provide base struct for db implementation.
type DBLinksRepository struct {
	db *sql.DB
}

// NewDBLinksRepository - provide new instance DBLinksRepository.
func NewDBLinksRepository(ctx context.Context, dbConf dbconf.DBConfigurer) (*DBLinksRepository, error) {
	db, err := ConnectDB(dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	logger.Sugar.Infof("SUCCESS connecting to db: %v", db.Stats())

	err = migrations.MakeMigrations(ctx, db, dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
		return nil, err
	}

	return &DBLinksRepository{
		db: db,
	}, nil
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
		return nil, repository.ErrCollectingDBConf
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
		return links, repository.ErrExistingLink
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

func (r *DBLinksRepository) AllUserURLs(ctx context.Context, userID string) ([]model.LinkData, error) {
	return []model.LinkData{}, nil
}
