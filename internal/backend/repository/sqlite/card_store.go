package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) ListCards(ctx context.Context, userID string) ([]core.Card, error) {
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE user_id = ?
		ORDER BY created_at ASC, korean ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) SearchCards(ctx context.Context, userID string, query string) ([]core.Card, error) {
	query = "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE user_id = ? AND (
			lower(id) LIKE ?
			OR lower(deck_id) LIKE ?
			OR lower(kind) LIKE ?
			OR lower(korean) LIKE ?
			OR lower(translation) LIKE ?
			OR lower(romanization) LIKE ?
			OR lower(example_korean) LIKE ?
			OR lower(example_translation) LIKE ?
			OR lower(tags) LIKE ?
			OR lower(created_at) LIKE ?
			OR lower(next_review_at) LIKE ?
			OR lower(COALESCE(last_review_at, '')) LIKE ?
			OR CAST(interval_days AS TEXT) LIKE ?
			OR CAST(ease_factor AS TEXT) LIKE ?
			OR CAST(review_count AS TEXT) LIKE ?
			OR CAST(lapse_count AS TEXT) LIKE ?)
		ORDER BY created_at ASC, korean ASC
	`, userID, query, query, query, query, query, query, query, query, query, query, query, query, query, query, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) ListDueCards(ctx context.Context, userID string, now time.Time) ([]core.Card, error) {
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE user_id = ? AND next_review_at <= ?
		ORDER BY next_review_at ASC, created_at ASC
	`, userID, formatTime(now))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) FindCardByID(ctx context.Context, userID string, id string) (core.Card, error) {
	row := store.db.QueryRowContext(ctx, cardSelectSQL()+`
		WHERE id = ? AND user_id = ?
	`, id, userID)

	card, err := scanCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Card{}, repository.ErrNotFound
	}

	return card, err
}

func (store *Store) CreateCard(ctx context.Context, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO cards (
			id, user_id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, card.ID, card.UserID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount)
	return normalizeSQLiteError(err)
}

func (store *Store) UpdateCard(ctx context.Context, userID string, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	result, err := store.db.ExecContext(ctx, `
		UPDATE cards
		SET deck_id = ?,
			kind = ?,
			korean = ?,
			translation = ?,
			romanization = ?,
			example_korean = ?,
			example_translation = ?,
			tags = ?,
			created_at = ?,
			next_review_at = ?,
			last_review_at = ?,
			interval_days = ?,
			ease_factor = ?,
			review_count = ?,
			lapse_count = ?
		WHERE id = ? AND user_id = ?
	`, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount, card.ID, userID)
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

func (store *Store) DeleteCard(ctx context.Context, userID string, id string) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM cards WHERE id = ? AND user_id = ?`, id, userID)
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

func (store *Store) DeleteCards(ctx context.Context, userID string, ids []string) (int, error) {
	query, args := inQuery(`DELETE FROM cards WHERE user_id = ? AND id IN `, ids)
	args = append([]any{userID}, args...)
	result, err := store.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func (store *Store) CreateCards(ctx context.Context, userID string, cards []core.Card) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	for _, card := range cards {
		if card.UserID != userID {
			return repository.ErrNotFound
		}
		if err := insertCard(tx, card); err != nil {
			return normalizeSQLiteError(err)
		}
	}
	return tx.Commit()
}

func (store *Store) UpdateCards(ctx context.Context, userID string, cards []core.Card) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	for _, card := range cards {
		if err := updateCard(ctx, tx, userID, card); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertCard(tx *sql.Tx, card core.Card) error {
	return insertCardWithConflict(tx, card, "")
}

func insertSeedCard(tx *sql.Tx, card core.Card) error {
	return insertCardWithConflict(tx, card, "ON CONFLICT(id) DO NOTHING")
}

func insertCardWithConflict(tx *sql.Tx, card core.Card, conflictClause string) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO cards (
			id, user_id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`+conflictClause+`
	`, card.ID, card.UserID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount)
	return err
}

func updateCard(ctx context.Context, tx *sql.Tx, userID string, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE cards SET deck_id = ?, kind = ?, korean = ?, translation = ?, romanization = ?,
			example_korean = ?, example_translation = ?, tags = ?, created_at = ?, next_review_at = ?,
			last_review_at = ?, interval_days = ?, ease_factor = ?, review_count = ?, lapse_count = ?
		WHERE id = ? AND user_id = ?
	`, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor, card.ReviewState.ReviewCount,
		card.ReviewState.LapseCount, card.ID, userID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func cardSelectSQL() string {
	return `
		SELECT
			id, user_id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		FROM cards
	`
}

func scanCards(rows *sql.Rows) ([]core.Card, error) {
	cards := make([]core.Card, 0)
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func scanCard(scanner rowScanner) (core.Card, error) {
	var card core.Card
	var kind string
	var tagsRaw string
	var createdAt string
	var nextReviewAt string
	var lastReviewAt sql.NullString

	err := scanner.Scan(
		&card.ID,
		&card.UserID,
		&card.DeckID,
		&kind,
		&card.Korean,
		&card.Translation,
		&card.Romanization,
		&card.ExampleKorean,
		&card.ExampleTranslation,
		&tagsRaw,
		&createdAt,
		&nextReviewAt,
		&lastReviewAt,
		&card.ReviewState.IntervalDays,
		&card.ReviewState.EaseFactor,
		&card.ReviewState.ReviewCount,
		&card.ReviewState.LapseCount,
	)
	if err != nil {
		return core.Card{}, err
	}

	if err := json.Unmarshal([]byte(tagsRaw), &card.Tags); err != nil {
		return core.Card{}, fmt.Errorf("decode card tags: %w", err)
	}
	card.Kind = core.CardKind(kind)
	card.CreatedAt = parseTime(createdAt)
	card.ReviewState.NextReviewAt = parseTime(nextReviewAt)
	if lastReviewAt.Valid {
		card.ReviewState.LastReviewAt = parseTime(lastReviewAt.String)
	}
	return card, nil
}
