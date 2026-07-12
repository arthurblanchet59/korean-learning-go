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

func (store *Store) Migrate() error {
	if _, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS decks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
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

		CREATE TABLE IF NOT EXISTS reviews (
			id TEXT PRIMARY KEY,
			card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			rating TEXT NOT NULL,
			reviewed_at TEXT NOT NULL,
			previous_state TEXT NOT NULL,
			next_state TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS lessons (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			level TEXT NOT NULL,
			sort_order INTEGER NOT NULL,
			content TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS lesson_progress (
			user_id TEXT NOT NULL,
			lesson_id TEXT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
			completed INTEGER NOT NULL DEFAULT 0,
			score INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (user_id, lesson_id)
		);

		CREATE TABLE IF NOT EXISTS journal_entries (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			original_text TEXT NOT NULL,
			corrected_text TEXT NOT NULL,
			corrections TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_seed_versions (
			user_id TEXT PRIMARY KEY,
			version INTEGER NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		return err
	}

	for _, migration := range []struct {
		table      string
		column     string
		definition string
	}{
		{"decks", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
		{"cards", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
		{"reviews", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
	} {
		if err := store.ensureColumn(migration.table, migration.column, migration.definition); err != nil {
			return err
		}
	}

	if _, err := store.db.Exec(`
		CREATE INDEX IF NOT EXISTS decks_user_id_idx ON decks(user_id);
		CREATE INDEX IF NOT EXISTS cards_user_id_idx ON cards(user_id);
		CREATE INDEX IF NOT EXISTS cards_deck_id_idx ON cards(deck_id);
		CREATE INDEX IF NOT EXISTS cards_next_review_at_idx ON cards(next_review_at);
		CREATE INDEX IF NOT EXISTS reviews_user_id_idx ON reviews(user_id);
		CREATE INDEX IF NOT EXISTS reviews_card_id_idx ON reviews(card_id);
		CREATE INDEX IF NOT EXISTS reviews_reviewed_at_idx ON reviews(reviewed_at);
		CREATE INDEX IF NOT EXISTS journal_user_id_idx ON journal_entries(user_id);
	`); err != nil {
		return err
	}

	return store.seedLessons()
}

func (store *Store) SeedIfEmpty() error {
	return store.SeedUser(context.Background(), "admin")
}

func (store *Store) SeedUser(ctx context.Context, userID string) error {
	var version int
	err := store.db.QueryRowContext(ctx, `SELECT version FROM user_seed_versions WHERE user_id = ?`, userID).Scan(&version)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if version >= curriculumSeedVersion {
		return nil
	}

	now := time.Now().UTC()
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	deckIDs := make(map[string]string)
	for _, deck := range core.SeedDecks(now) {
		templateID := deck.ID
		deck.ID = userID + "-" + templateID
		deck.UserID = userID
		deckIDs[templateID] = deck.ID
		if _, err := tx.Exec(`
			INSERT INTO decks (id, user_id, name, description, created_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(id) DO NOTHING
		`, deck.ID, deck.UserID, deck.Name, deck.Description, formatTime(deck.CreatedAt)); err != nil {
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

	if _, err := tx.Exec(`
		INSERT INTO user_seed_versions (user_id, version, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			version = excluded.version,
			updated_at = excluded.updated_at
	`, userID, curriculumSeedVersion, formatTime(now)); err != nil {
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

func (store *Store) ensureColumn(table string, column string, definition string) error {
	rows, err := store.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = store.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
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

func (store *Store) Reset(ctx context.Context) (repository.ResetResult, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ResetResult{}, err
	}
	defer rollback(tx)

	result := repository.ResetResult{}
	var resetErr error

	if result.DeletedReviews, resetErr = execDelete(ctx, tx, `DELETE FROM reviews`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedCards, resetErr = execDelete(ctx, tx, `DELETE FROM cards`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedDecks, resetErr = execDelete(ctx, tx, `DELETE FROM decks`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedJournal, resetErr = execDelete(ctx, tx, `DELETE FROM journal_entries`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedProgress, resetErr = execDelete(ctx, tx, `DELETE FROM lesson_progress`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if _, resetErr = execDelete(ctx, tx, `DELETE FROM user_seed_versions`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedUsers, resetErr = execDelete(ctx, tx, `DELETE FROM users WHERE is_admin = 0`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}

	if err := tx.Commit(); err != nil {
		return repository.ResetResult{}, err
	}

	return result, nil
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
