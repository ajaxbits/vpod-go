package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
)

func CreateFeed(ctx context.Context, db *sql.DB, title *string, channel_id *string, desc *string, link *string) error {
	uuid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO Feeds (id, created_at, channel_id, description, title, updated_at, link) VALUES (?, CURRENT_TIMESTAMP, ?, ?, ?, CURRENT_TIMESTAMP, ?) RETURNING created_at`
	_, err = tx.ExecContext(ctx, query, uuid, channel_id, desc, title, link)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
