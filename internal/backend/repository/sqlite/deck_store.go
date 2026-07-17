package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) ListDecks(ctx context.Context, userID string) ([]core.Deck, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, created_at
		FROM decks
		WHERE user_id = ?
		ORDER BY created_at ASC, name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decks := make([]core.Deck, 0)
	for rows.Next() {
		var deck core.Deck
		var createdAt string
		if err := rows.Scan(&deck.ID, &deck.UserID, &deck.Name, &deck.Description, &createdAt); err != nil {
			return nil, err
		}
		deck.CreatedAt = parseTime(createdAt)
		decks = append(decks, deck)
	}

	return decks, rows.Err()
}

func (store *Store) SearchDecks(ctx context.Context, userID string, query string) ([]core.Deck, error) {
	query = "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, created_at
		FROM decks
		WHERE user_id = ? AND (
			lower(id) LIKE ?
			OR lower(name) LIKE ?
			OR lower(description) LIKE ?
			OR lower(created_at) LIKE ?)
		ORDER BY created_at ASC, name ASC
	`, userID, query, query, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDecks(rows)
}

func (store *Store) FindDeckByID(ctx context.Context, userID string, id string) (core.Deck, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, created_at
		FROM decks
		WHERE id = ? AND user_id = ?
	`, id, userID)

	deck, err := scanDeck(row)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Deck{}, repository.ErrNotFound
	}
	return deck, err
}

func (store *Store) CreateDeck(ctx context.Context, deck core.Deck) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO decks (id, user_id, name, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, deck.ID, deck.UserID, deck.Name, deck.Description, formatTime(deck.CreatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) UpdateDeck(ctx context.Context, userID string, deck core.Deck) error {
	result, err := store.db.ExecContext(ctx, `
		UPDATE decks
		SET name = ?,
			description = ?
		WHERE id = ? AND user_id = ?
	`, deck.Name, deck.Description, deck.ID, userID)
	if err != nil {
		return normalizeSQLiteError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (store *Store) DeleteDeck(ctx context.Context, userID string, id string) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM decks WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (store *Store) DeleteDecks(ctx context.Context, userID string, ids []string) (int, error) {
	query, args := inQuery(`DELETE FROM decks WHERE user_id = ? AND id IN `, ids)
	args = append([]any{userID}, args...)
	result, err := store.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func (store *Store) UpdateDecks(ctx context.Context, userID string, decks []core.Deck) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	for _, deck := range decks {
		result, err := tx.ExecContext(ctx, `UPDATE decks SET name = ?, description = ? WHERE id = ? AND user_id = ?`, deck.Name, deck.Description, deck.ID, userID)
		if err != nil {
			return err
		}
		if err := requireAffected(result); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func scanDecks(rows *sql.Rows) ([]core.Deck, error) {
	decks := make([]core.Deck, 0)
	for rows.Next() {
		deck, err := scanDeck(rows)
		if err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}
	return decks, rows.Err()
}

func scanDeck(scanner rowScanner) (core.Deck, error) {
	var deck core.Deck
	var createdAt string
	err := scanner.Scan(&deck.ID, &deck.UserID, &deck.Name, &deck.Description, &createdAt)
	if err != nil {
		return core.Deck{}, err
	}
	deck.CreatedAt = parseTime(createdAt)
	return deck, nil
}
