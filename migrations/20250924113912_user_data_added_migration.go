package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upUserDataAddedMigration, downUserDataAddedMigration)
}

func upUserDataAddedMigration(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links ADD COLUMN IF NOT EXISTS user_id integer;`)
	if err != nil {
		return fmt.Errorf("up add column user_id error: %w", err)
	}

	return nil
}

func downUserDataAddedMigration(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links DROP COLUMN IF EXISTS user_id;`)
	if err != nil {
		return fmt.Errorf("down drop column user_id error: %w", err)
	}

	return nil
}
