package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitAddUserIsDeleteFlag, downInitAddUserIsDeleteFlag)
}

func upInitAddUserIsDeleteFlag(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links ADD COLUMN IF NOT EXISTS is_deleted boolean DEFAULT False;`)
	if err != nil {
		return fmt.Errorf("up add column is_deleted error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links
		ALTER COLUMN user_id 
		SET DATA TYPE UUID USING gen_random_uuid();`)
	if err != nil {
		return fmt.Errorf("up alter column user_id error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY);`)
	if err != nil {
		return fmt.Errorf("up create table users error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (id) SELECT user_id FROM links;`)
	if err != nil {
		return fmt.Errorf("up transfer existing users error: %w", err)
	}

	// _, err = tx.ExecContext(ctx,
	// 	`ALTER TABLE links
	// 		ADD CONSTRAINT fk_user_id
	// 		FOREIGN KEY (user_id)
	// 		REFERENCES users (id)
	// 		ON DELETE CASCADE
	// 		ON UPDATE NO ACTION;`)
	// if err != nil {
	// 	return fmt.Errorf("up added fk users error: %w", err)
	// }

	return nil
}

func downInitAddUserIsDeleteFlag(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links DROP COLUMN IF EXISTS is_deleted;`)
	if err != nil {
		return fmt.Errorf("down drop column is_deleted error: %w", err)
	}

	// _, err = tx.ExecContext(ctx,
	// 	`ALTER TABLE links
	// 	 DROP CONSTRAINT fk_user_id;`)
	// if err != nil {
	// 	return fmt.Errorf("down CONSTRAINT fk_user_id links error: %w", err)
	// }

	_, err = tx.ExecContext(ctx,
		`ALTER TABLE IF EXISTS links
		ALTER COLUMN user_id 
		TYPE INT USING SUBSTRING(MD5(user_id::text), 1, 9)::INT;`)
	if err != nil {
		return fmt.Errorf("down alter column user_id error: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`DROP TABLE IF EXISTS users CASCADE;`)
	if err != nil {
		return fmt.Errorf("down create table users error: %w", err)
	}

	return nil
}
