package sqlite

import (
	"context"
	"encoding/json"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) CreateReview(ctx context.Context, review core.Review) error {
	previousJSON, err := json.Marshal(review.Previous)
	if err != nil {
		return err
	}
	nextJSON, err := json.Marshal(review.Next)
	if err != nil {
		return err
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO reviews (id, user_id, card_id, rating, reviewed_at, previous_state, next_state)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, review.ID, review.UserID, review.CardID, review.Rating, formatTime(review.ReviewedAt), string(previousJSON), string(nextJSON))
	return err
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

func (store *Store) ListReviewsSince(ctx context.Context, userID string, since time.Time) ([]core.Review, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, user_id, card_id, rating, reviewed_at, previous_state, next_state
		FROM reviews
		WHERE user_id = ? AND reviewed_at >= ?
		ORDER BY reviewed_at ASC
	`, userID, formatTime(since))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]core.Review, 0)
	for rows.Next() {
		var review core.Review
		var rating, reviewedAt, previousRaw, nextRaw string
		if err := rows.Scan(&review.ID, &review.UserID, &review.CardID, &rating, &reviewedAt, &previousRaw, &nextRaw); err != nil {
			return nil, err
		}
		review.Rating = core.Rating(rating)
		review.ReviewedAt = parseTime(reviewedAt)
		if err := json.Unmarshal([]byte(previousRaw), &review.Previous); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(nextRaw), &review.Next); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}
