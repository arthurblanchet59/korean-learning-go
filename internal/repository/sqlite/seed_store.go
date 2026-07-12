package sqlite

import (
	"context"
	"fmt"
)

func (store *Store) SeedIfEmpty() error {
	return store.SeedUser(context.Background(), "admin")
}

func (store *Store) SeedUser(ctx context.Context, userID string) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := seedUserData(ctx, tx, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (store *Store) SeedAllUsers(ctx context.Context) error {
	rows, err := store.db.QueryContext(ctx, `SELECT id FROM users ORDER BY created_at ASC`)
	if err != nil {
		return err
	}

	userIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			_ = rows.Close()
			return err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := store.SeedUser(ctx, userID); err != nil {
			return fmt.Errorf("seed curriculum for user %s: %w", userID, err)
		}
	}
	return nil
}
