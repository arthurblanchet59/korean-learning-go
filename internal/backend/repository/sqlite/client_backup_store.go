package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
)

func (store *Store) FindClientBackup(ctx context.Context, userID string) (domain.ClientBackup, error) {
	var backup domain.ClientBackup
	var configJSON, stateJSON, updatedAt string
	err := store.db.QueryRowContext(ctx, `
		SELECT user_id, config_json, state_json, updated_at
		FROM client_backups
		WHERE user_id = ?
	`, userID).Scan(&backup.UserID, &configJSON, &stateJSON, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ClientBackup{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.ClientBackup{}, err
	}

	backup.Config = []byte(configJSON)
	backup.State = []byte(stateJSON)
	backup.UpdatedAt = parseTime(updatedAt)
	return backup, nil
}

func (store *Store) UpsertClientBackup(ctx context.Context, backup domain.ClientBackup) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO client_backups (user_id, config_json, state_json, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			config_json = excluded.config_json,
			state_json = excluded.state_json,
			updated_at = excluded.updated_at
	`, backup.UserID, string(backup.Config), string(backup.State), formatTime(backup.UpdatedAt))
	return err
}
