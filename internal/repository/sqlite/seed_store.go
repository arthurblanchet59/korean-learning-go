package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

const curriculumSeedVersion = 2

func (store *Store) CreateUserWithSeed(ctx context.Context, user domain.User) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Name, user.Email, user.PasswordHash, boolInt(user.IsAdmin), formatTime(user.CreatedAt), formatTime(user.UpdatedAt)); err != nil {
		return normalizeSQLiteError(err)
	}
	if err := seedUserData(ctx, tx, user.ID); err != nil {
		return err
	}
	return tx.Commit()
}

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

func seedUserData(ctx context.Context, tx *sql.Tx, userID string) error {
	var version int
	err := tx.QueryRowContext(ctx, `SELECT version FROM user_seed_versions WHERE user_id = ?`, userID).Scan(&version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if version >= curriculumSeedVersion {
		return nil
	}

	now := time.Now().UTC()
	deckIDs := make(map[string]string)
	for _, deck := range core.SeedDecks(now) {
		templateID := deck.ID
		deck.ID = userID + "-" + templateID
		deck.UserID = userID
		deckIDs[templateID] = deck.ID
		if _, err := tx.ExecContext(ctx, `INSERT INTO decks (id, user_id, name, description, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING`, deck.ID, deck.UserID, deck.Name, deck.Description, formatTime(deck.CreatedAt)); err != nil {
			return err
		}
	}
	for _, card := range core.SeedCards(now) {
		card.ID = userID + "-" + card.ID
		card.UserID = userID
		card.DeckID = deckIDs[card.DeckID]
		if err := insertSeedCard(tx, card); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO user_seed_versions (user_id, version, updated_at) VALUES (?, ?, ?) ON CONFLICT(user_id) DO UPDATE SET version = excluded.version, updated_at = excluded.updated_at`, userID, curriculumSeedVersion, formatTime(now))
	return err
}
