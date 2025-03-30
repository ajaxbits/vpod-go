package db

import (
	"context"
	"database/sql"
	// "fmt"
)

func CreateFeed(ctx context.Context, db *sql.DB, channel_id *string, title *string, desc *string, link *string, xml *string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO Feeds (id, created_at, description, title, updated_at, link, xml) VALUES (?, CURRENT_TIMESTAMP, ?, ?, CURRENT_TIMESTAMP, ?, ?)`
	_, err = tx.ExecContext(ctx, query, channel_id, desc, title, link, xml)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetFeed(ctx context.Context, db *sql.DB, channel_id *string) (*string, error) {
	var _channelId string
	var xml string
	query := `SELECT id, xml FROM Feeds WHERE id = ?`
	err := db.QueryRowContext(ctx, query, channel_id).Scan(&_channelId, &xml)
	if err != nil {
		return nil, err
	} else {
		return &xml, nil
	}
}

func UpdateFeed(ctx context.Context, db *sql.DB, channelId *string, title *string, desc *string, link *string, xml *string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE Feeds SET (description, title, updated_at, link, xml) VALUES (?, ?, CURRENT_TIMESTAMP, ?, ?) WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, desc, title, link, xml, channelId)
	if err != nil {
		return err
	}

	return tx.Commit()
}
