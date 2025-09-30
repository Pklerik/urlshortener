// Package dbrepo provide realization of repository for db storage.
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

// LinksRepositoryPostgres provide base struct for db implementation.
type LinksRepositoryPostgres struct {
	db *sql.DB
}

// NewDBLinksRepository - provide new instance DBLinksRepository.
func NewDBLinksRepository(ctx context.Context, dbConf dbconf.DBConfigurer) (*LinksRepositoryPostgres, error) {
	db, err := ConnectDB(dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	logger.Sugar.Infof("SUCCESS connecting to db: %v", db.Stats())

	err = migrations.MakeMigrations(ctx, db, dbConf)
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
		return nil, fmt.Errorf("NewDBLinksRepository: %w", err)
	}

	return &LinksRepositoryPostgres{
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

// SetLinks - writes linkData pointer to internal DBLinksRepository map Shorts.
func (r *LinksRepositoryPostgres) SetLinks(ctx context.Context, links []model.LinkData) ([]model.LinkData, error) {
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

func (r *LinksRepositoryPostgres) insertBatch(ctx context.Context, links []model.LinkData) ([]model.LinkData, error) {
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

	linksData, err := collectLinks(rows)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", pgi*4+1, pgi*4+2, pgi*4+3, pgi*4+4))
		queryArgs = append(queryArgs, linkData.UUID, linkData.ShortURL, linkData.LongURL, linkData.UserID)
		logger.Sugar.Infof("Short url: %s sets for long: %s by userID: %d", linkData.ShortURL, linkData.LongURL, linkData.UserID)
	}

	return fmt.Sprintf("INSERT INTO links (id, short_url, long_url, user_id) VALUES %s ON CONFLICT (short_url) DO NOTHING RETURNING id, short_url, long_url, user_id", strings.Join(placeholders, ", ")), queryArgs
}

// TODO  collectLinks(rows *sql.Rows, data *any, items ...any) (int, error)
// provide unification for rows scanning
// return number of inserted rows and error
func collectLinks(rows *sql.Rows) ([]model.LinkData, error) {
	linksData := make([]model.LinkData, 0, 1)
	for rows.Next() {
		linkData := model.LinkData{}

		err := rows.Scan(&linkData.UUID, &linkData.ShortURL, &linkData.LongURL, &linkData.UserID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Sugar.Errorf("error scanning QUERY result: %w", err)
				return linksData, fmt.Errorf("error scanning QUERY result: %w", err)
			}

			return linksData, fmt.Errorf("collectLinks: %w", err)
		}

		linksData = append(linksData, linkData)
	}

	return linksData, nil
}

// FindShort - provide model.LinkData and error
// If shortURL is absent returns nil.
func (r *LinksRepositoryPostgres) FindShort(ctx context.Context, short string) (model.LinkData, error) {
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
func (r *LinksRepositoryPostgres) PingDB(_ context.Context) error {
	err := r.db.Ping()
	if err != nil {
		logger.Sugar.Errorf("Cant connect to db server: %w", err)
	}

	return nil
}

func (r *LinksRepositoryPostgres) getShort(ctx context.Context, tx *sql.Tx, short string) (*model.LinkData, error) {
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

// SelectUserLinks selects user links by userID.
func (r *LinksRepositoryPostgres) SelectUserLinks(ctx context.Context, userID model.UserID) ([]model.LinkData, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return []model.LinkData{}, fmt.Errorf("error creating tx error: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT id, short_url, long_url, user_id FROM links WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("error selecting link data: %w", err)
	}

	lds, err := collectLinks(rows)
	if err != nil {
		return nil, fmt.Errorf("error collecting link data: %w", err)
	}

	return lds, nil
}

// CreateUser creates user.
func (r *LinksRepositoryPostgres) CreateUser(ctx context.Context, userID model.UserID) (model.User, error) {
	user := model.User{ID: userID}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return user, fmt.Errorf("error creating tx error: %w", err)
	}
	defer tx.Commit()

	rows, err := tx.QueryContext(ctx, `INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING RETURNING id`, user.ID)
	if err != nil {
		return user, fmt.Errorf("error selecting link data: %w", err)
	}
	for rows.Next() {
		err := rows.Scan(&user.ID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Sugar.Errorf("error scanning QUERY result: %w", err)

				return user, fmt.Errorf("error scanning QUERY result: %w", err)
			}

			return user, fmt.Errorf("collect users: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return user, fmt.Errorf("committing insertion error: %w", err)
	}

	return user, nil
}

func (r *LinksRepositoryPostgres) BatchMarkAsDeleted(ctx context.Context, links []model.LinkData) (int, error) {
	return 0, nil
}
