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

const curriculumSeedVersion = 2

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
	err := scanner.Scan(&deck.ID, &deck.UserID, &deck.Name, &deck.Description, &createdAt)
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

func execDelete(ctx context.Context, tx *sql.Tx, query string, args ...any) (int, error) {
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	return int(affected), err
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
