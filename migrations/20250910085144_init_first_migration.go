package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitFirstMigration, downInitFirstMigration)
}

func upInitFirstMigration(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS links (
			id UUID PRIMARY KEY,
			short_url VARCHAR(10),
			long_url VARCHAR(255));`)
	if err != nil {
		return fmt.Errorf("up create table error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`CREATE UNIQUE INDEX idx_short_url ON links (short_url);`)
	if err != nil {
		return fmt.Errorf("up create index idx_short_url error: %w", err)
	}

	_, _ = tx.ExecContext(ctx,
		`INSERT INTO links (id, short_url, long_url)
		 VALUES ('019906ca-14b2-7589-b77d-32e3fe12402a', '398f0ca4', 'http://ya.ru')`)

	return nil
}

func downInitFirstMigration(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.ExecContext(ctx,
		`DROP TABLE IF EXISTS links CASCADE;`)
	if err != nil {
		return fmt.Errorf("down drop schema error: %w", err)
	}

	return nil
}
