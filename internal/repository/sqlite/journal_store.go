package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) ListJournalEntries(ctx context.Context, userID string) ([]core.JournalEntry, error) {
	rows, err := store.db.QueryContext(ctx, journalSelectSQL()+` WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]core.JournalEntry, 0)
	for rows.Next() {
		entry, err := scanJournalEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (store *Store) FindJournalEntryByID(ctx context.Context, userID string, id string) (core.JournalEntry, error) {
	entry, err := scanJournalEntry(store.db.QueryRowContext(ctx, journalSelectSQL()+` WHERE id = ? AND user_id = ?`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return core.JournalEntry{}, repository.ErrNotFound
	}
	return entry, err
}

func (store *Store) CreateJournalEntry(ctx context.Context, entry core.JournalEntry) error {
	corrections, err := json.Marshal(entry.Corrections)
	if err != nil {
		return err
	}
	_, err = store.db.ExecContext(ctx, `
		INSERT INTO journal_entries (id, user_id, title, original_text, corrected_text, corrections, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.UserID, entry.Title, entry.OriginalText, entry.CorrectedText, string(corrections), formatTime(entry.CreatedAt), formatTime(entry.UpdatedAt))
	return err
}

func (store *Store) UpdateJournalEntry(ctx context.Context, entry core.JournalEntry) error {
	corrections, err := json.Marshal(entry.Corrections)
	if err != nil {
		return err
	}
	result, err := store.db.ExecContext(ctx, `
		UPDATE journal_entries SET title = ?, original_text = ?, corrected_text = ?, corrections = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, entry.Title, entry.OriginalText, entry.CorrectedText, string(corrections), formatTime(entry.UpdatedAt), entry.ID, entry.UserID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (store *Store) DeleteJournalEntry(ctx context.Context, userID string, id string) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM journal_entries WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func journalSelectSQL() string {
	return `SELECT id, user_id, title, original_text, corrected_text, corrections, created_at, updated_at FROM journal_entries`
}

func scanJournalEntry(scanner rowScanner) (core.JournalEntry, error) {
	var entry core.JournalEntry
	var correctionsRaw, createdAt, updatedAt string
	if err := scanner.Scan(&entry.ID, &entry.UserID, &entry.Title, &entry.OriginalText, &entry.CorrectedText, &correctionsRaw, &createdAt, &updatedAt); err != nil {
		return core.JournalEntry{}, err
	}
	if err := json.Unmarshal([]byte(correctionsRaw), &entry.Corrections); err != nil {
		return core.JournalEntry{}, err
	}
	entry.CreatedAt = parseTime(createdAt)
	entry.UpdatedAt = parseTime(updatedAt)
	return entry, nil
}
