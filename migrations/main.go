package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Pklerik/urlshortener/internal/config/dbconf"
	"github.com/Pklerik/urlshortener/internal/logger"
	"github.com/pressly/goose/v3"
)

// ErrEmptyDB db is nil.
var ErrEmptyDB = errors.New("db is nil: %w")

// MakeMigrations makes migrations in current dir.
func MakeMigrations(ctx context.Context, db *sql.DB, DBConf dbconf.DBConfigurer) error {
	if db == nil {
		logger.Sugar.Error("db is nil: %w")
		return ErrEmptyDB
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("crating tx error: %w", err)
	}
	scheme := DBConf.GetOptions()["search_path"]
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s AUTHORIZATION %s;", scheme, DBConf.GetUser())
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("up create schema error: %w", err)
	}
	tx.Commit()

	goose.SetTableName(fmt.Sprintf("%s.goose_db_version", scheme))
	err = goose.Up(db, ".")
	if err != nil {
		logger.Sugar.Errorf("Can't make migrations to db server: %w", err)
		return fmt.Errorf("can't make migrations to db server: %w", err)
	}
	logger.Sugar.Infof("SUCCESS making migration to db: %v", db.Stats())
	return nil
}
