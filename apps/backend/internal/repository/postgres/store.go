package postgres

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Migrate(ctx context.Context) error {
	migration, err := migrationFiles.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = store.db.ExecContext(ctx, string(migration))
	return err
}

func (store *Store) SeedIfEmpty(ctx context.Context) error {
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM decks`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	deck := core.SeedDeck(now)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO decks (id, name, description, created_at)
		VALUES ($1, $2, $3, $4)
	`, deck.ID, deck.Name, deck.Description, deck.CreatedAt); err != nil {
		return err
	}

	for _, card := range core.SeedCards(now) {
		if err := insertCard(ctx, tx, card); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (store *Store) ListDecks(ctx context.Context) ([]core.Deck, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, name, description, created_at
		FROM decks
		ORDER BY created_at ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decks := make([]core.Deck, 0)
	for rows.Next() {
		var deck core.Deck
		if err := rows.Scan(&deck.ID, &deck.Name, &deck.Description, &deck.CreatedAt); err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}

	return decks, rows.Err()
}

func (store *Store) ListCards(ctx context.Context) ([]core.Card, error) {
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		ORDER BY created_at ASC, korean ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) ListDueCards(ctx context.Context, now time.Time) ([]core.Card, error) {
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE next_review_at <= $1
		ORDER BY next_review_at ASC, created_at ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) FindCardByID(ctx context.Context, id string) (core.Card, error) {
	row := store.db.QueryRowContext(ctx, cardSelectSQL()+`
		WHERE id = $1
	`, id)

	card, err := scanCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Card{}, repository.ErrNotFound
	}

	return card, err
}

func (store *Store) UpdateCard(ctx context.Context, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	result, err := store.db.ExecContext(ctx, `
		UPDATE cards
		SET deck_id = $2,
			kind = $3,
			korean = $4,
			translation = $5,
			romanization = $6,
			example_korean = $7,
			example_translation = $8,
			tags = $9,
			created_at = $10,
			next_review_at = $11,
			last_review_at = $12,
			interval_days = $13,
			ease_factor = $14,
			review_count = $15,
			lapse_count = $16
		WHERE id = $1
	`, card.ID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, tagsJSON, card.CreatedAt,
		card.ReviewState.NextReviewAt, nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount)
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
		INSERT INTO reviews (id, card_id, rating, reviewed_at, previous_state, next_state)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, review.ID, review.CardID, review.Rating, review.ReviewedAt, previousJSON, nextJSON)
	return err
}

func insertCard(ctx context.Context, tx *sql.Tx, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO cards (
			id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, card.ID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, tagsJSON, card.CreatedAt,
		card.ReviewState.NextReviewAt, nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount)
	return err
}

func cardSelectSQL() string {
	return `
		SELECT
			id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		FROM cards
	`
}

type rowScanner interface {
	Scan(dest ...any) error
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
	var tagsRaw []byte
	var lastReviewAt sql.NullTime

	err := scanner.Scan(
		&card.ID,
		&card.DeckID,
		&kind,
		&card.Korean,
		&card.Translation,
		&card.Romanization,
		&card.ExampleKorean,
		&card.ExampleTranslation,
		&tagsRaw,
		&card.CreatedAt,
		&card.ReviewState.NextReviewAt,
		&lastReviewAt,
		&card.ReviewState.IntervalDays,
		&card.ReviewState.EaseFactor,
		&card.ReviewState.ReviewCount,
		&card.ReviewState.LapseCount,
	)
	if err != nil {
		return core.Card{}, err
	}

	if err := json.Unmarshal(tagsRaw, &card.Tags); err != nil {
		return core.Card{}, fmt.Errorf("decode card tags: %w", err)
	}

	card.Kind = core.CardKind(kind)
	if lastReviewAt.Valid {
		card.ReviewState.LastReviewAt = lastReviewAt.Time
	}

	return card, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
