package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) Migrate() error {
	_, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS decks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS cards (
			id TEXT PRIMARY KEY,
			deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			korean TEXT NOT NULL,
			translation TEXT NOT NULL,
			romanization TEXT NOT NULL DEFAULT '',
			example_korean TEXT NOT NULL DEFAULT '',
			example_translation TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			next_review_at TEXT NOT NULL,
			last_review_at TEXT,
			interval_days INTEGER NOT NULL DEFAULT 0,
			ease_factor REAL NOT NULL DEFAULT 2.5,
			review_count INTEGER NOT NULL DEFAULT 0,
			lapse_count INTEGER NOT NULL DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS cards_deck_id_idx ON cards(deck_id);
		CREATE INDEX IF NOT EXISTS cards_next_review_at_idx ON cards(next_review_at);

		CREATE TABLE IF NOT EXISTS reviews (
			id TEXT PRIMARY KEY,
			card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			rating TEXT NOT NULL,
			reviewed_at TEXT NOT NULL,
			previous_state TEXT NOT NULL,
			next_state TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS reviews_card_id_idx ON reviews(card_id);
		CREATE INDEX IF NOT EXISTS reviews_reviewed_at_idx ON reviews(reviewed_at);
	`)
	return err
}

func (store *Store) SeedIfEmpty() error {
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM decks`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	defer rollback(tx)

	deck := core.SeedDeck(now)
	if _, err := tx.Exec(`
		INSERT INTO decks (id, name, description, created_at)
		VALUES (?, ?, ?, ?)
	`, deck.ID, deck.Name, deck.Description, formatTime(deck.CreatedAt)); err != nil {
		return err
	}

	for _, card := range core.SeedCards(now) {
		if err := insertCard(tx, card); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (store *Store) EnsureAdmin(ctx context.Context, user domain.User) error {
	var existingID string
	err := store.db.QueryRowContext(ctx, `
		SELECT id
		FROM users
		WHERE is_admin = 1
		ORDER BY created_at ASC
		LIMIT 1
	`).Scan(&existingID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)
	`, user.ID, user.Name, user.Email, user.PasswordHash, formatTime(user.CreatedAt), formatTime(user.UpdatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) CreateUser(ctx context.Context, user domain.User) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Name, user.Email, user.PasswordHash, boolInt(user.IsAdmin), formatTime(user.CreatedAt), formatTime(user.UpdatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) FindUserByID(ctx context.Context, id string) (domain.User, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id)

	return scanUser(row)
}

func (store *Store) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE email = ?
	`, strings.ToLower(strings.TrimSpace(email)))

	return scanUser(row)
}

func (store *Store) UpdateUser(ctx context.Context, user domain.User) error {
	result, err := store.db.ExecContext(ctx, `
		UPDATE users
		SET name = ?,
			email = ?,
			password_hash = ?,
			is_admin = ?,
			updated_at = ?
		WHERE id = ?
	`, user.Name, user.Email, user.PasswordHash, boolInt(user.IsAdmin), formatTime(user.UpdatedAt), user.ID)
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
		var createdAt string
		if err := rows.Scan(&deck.ID, &deck.Name, &deck.Description, &createdAt); err != nil {
			return nil, err
		}
		deck.CreatedAt = parseTime(createdAt)
		decks = append(decks, deck)
	}

	return decks, rows.Err()
}

func (store *Store) SearchDecks(ctx context.Context, query string) ([]core.Deck, error) {
	query = "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, name, description, created_at
		FROM decks
		WHERE lower(id) LIKE ?
			OR lower(name) LIKE ?
			OR lower(description) LIKE ?
			OR lower(created_at) LIKE ?
		ORDER BY created_at ASC, name ASC
	`, query, query, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDecks(rows)
}

func (store *Store) FindDeckByID(ctx context.Context, id string) (core.Deck, error) {
	row := store.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at
		FROM decks
		WHERE id = ?
	`, id)

	deck, err := scanDeck(row)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Deck{}, repository.ErrNotFound
	}
	return deck, err
}

func (store *Store) CreateDeck(ctx context.Context, deck core.Deck) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO decks (id, name, description, created_at)
		VALUES (?, ?, ?, ?)
	`, deck.ID, deck.Name, deck.Description, formatTime(deck.CreatedAt))
	return normalizeSQLiteError(err)
}

func (store *Store) UpdateDeck(ctx context.Context, deck core.Deck) error {
	result, err := store.db.ExecContext(ctx, `
		UPDATE decks
		SET name = ?,
			description = ?
		WHERE id = ?
	`, deck.Name, deck.Description, deck.ID)
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

func (store *Store) DeleteDeck(ctx context.Context, id string) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM decks WHERE id = ?`, id)
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

func (store *Store) DeleteDecks(ctx context.Context, ids []string) (int, error) {
	query, args := inQuery(`DELETE FROM decks WHERE id IN `, ids)
	result, err := store.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
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

func (store *Store) SearchCards(ctx context.Context, query string) ([]core.Card, error) {
	query = "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE lower(id) LIKE ?
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
			OR CAST(lapse_count AS TEXT) LIKE ?
		ORDER BY created_at ASC, korean ASC
	`, query, query, query, query, query, query, query, query, query, query, query, query, query, query, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) ListDueCards(ctx context.Context, now time.Time) ([]core.Card, error) {
	rows, err := store.db.QueryContext(ctx, cardSelectSQL()+`
		WHERE next_review_at <= ?
		ORDER BY next_review_at ASC, created_at ASC
	`, formatTime(now))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCards(rows)
}

func (store *Store) FindCardByID(ctx context.Context, id string) (core.Card, error) {
	row := store.db.QueryRowContext(ctx, cardSelectSQL()+`
		WHERE id = ?
	`, id)

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
			id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, card.ID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount)
	return normalizeSQLiteError(err)
}

func (store *Store) UpdateCard(ctx context.Context, card core.Card) error {
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
		WHERE id = ?
	`, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
		card.ReviewState.IntervalDays, card.ReviewState.EaseFactor,
		card.ReviewState.ReviewCount, card.ReviewState.LapseCount, card.ID)
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

func (store *Store) DeleteCard(ctx context.Context, id string) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM cards WHERE id = ?`, id)
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

func (store *Store) DeleteCards(ctx context.Context, ids []string) (int, error) {
	query, args := inQuery(`DELETE FROM cards WHERE id IN `, ids)
	result, err := store.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
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
		VALUES (?, ?, ?, ?, ?, ?)
	`, review.ID, review.CardID, review.Rating, formatTime(review.ReviewedAt), string(previousJSON), string(nextJSON))
	return err
}

func insertCard(tx *sql.Tx, card core.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO cards (
			id, deck_id, kind, korean, translation, romanization,
			example_korean, example_translation, tags, created_at,
			next_review_at, last_review_at, interval_days, ease_factor,
			review_count, lapse_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, card.ID, card.DeckID, card.Kind, card.Korean, card.Translation, card.Romanization,
		card.ExampleKorean, card.ExampleTranslation, string(tagsJSON), formatTime(card.CreatedAt),
		formatTime(card.ReviewState.NextReviewAt), nullableTime(card.ReviewState.LastReviewAt),
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
	err := scanner.Scan(&deck.ID, &deck.Name, &deck.Description, &createdAt)
	if err != nil {
		return core.Deck{}, err
	}

	deck.CreatedAt = parseTime(createdAt)
	return deck, nil
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

func scanUser(scanner rowScanner) (domain.User, error) {
	var user domain.User
	var isAdmin int
	var createdAt string
	var updatedAt string

	err := scanner.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&isAdmin,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	user.IsAdmin = isAdmin == 1
	user.CreatedAt = parseTime(createdAt)
	user.UpdatedAt = parseTime(updatedAt)

	return user, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}

	return parsed
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return formatTime(value)
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizeSQLiteError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
		return repository.ErrConflict
	}
	return err
}

func inQuery(prefix string, ids []string) (string, []any) {
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	return prefix + "(" + strings.Join(placeholders, ",") + ")", args
}
