package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

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

func (store *Store) SaveReview(ctx context.Context, userID string, card core.Card, review core.Review) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := updateCard(ctx, tx, userID, card); err != nil {
		return err
	}
	if review.UserID != userID || review.CardID != card.ID {
		return repository.ErrNotFound
	}
	previousJSON, err := json.Marshal(review.Previous)
	if err != nil {
		return err
	}
	nextJSON, err := json.Marshal(review.Next)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO reviews (id, user_id, card_id, rating, reviewed_at, previous_state, next_state)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, review.ID, review.UserID, review.CardID, review.Rating, formatTime(review.ReviewedAt), string(previousJSON), string(nextJSON)); err != nil {
		return err
	}
	return tx.Commit()
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
