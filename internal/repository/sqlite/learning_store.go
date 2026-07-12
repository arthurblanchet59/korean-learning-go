package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) seedLessons() error {
	lessons := []core.Lesson{
		{ID: "hangul-1", Title: "Lire les voyelles", Description: "Les voyelles de base du hangeul.", Level: "A0", Order: 1, Content: "ㅏ a · ㅓ eo · ㅗ o · ㅜ u · ㅡ eu · ㅣ i\n\nLis chaque signe puis associe-le a sa romanisation."},
		{ID: "hangul-2", Title: "Former une syllabe", Description: "Assembler consonne et voyelle en blocs.", Level: "A0", Order: 2, Content: "Une syllabe coreenne forme un bloc. ㄱ + ㅏ donne 가 (ga), et ㄴ + ㅏ donne 나 (na)."},
		{ID: "grammar-topic", Title: "La particule de theme", Description: "Choisir 은 ou 는 selon le batchim.", Level: "A1", Order: 3, Content: "Apres une consonne finale, utilise 은. Apres une voyelle, utilise 는. Exemples : 학생은, 저는."},
		{ID: "grammar-object", Title: "La particule d'objet", Description: "Choisir 을 ou 를.", Level: "A1", Order: 4, Content: "Apres une consonne finale, utilise 을. Apres une voyelle, utilise 를. Exemples : 물을, 커피를."},
		{ID: "daily-sentences", Title: "Raconter sa journee", Description: "Construire des phrases simples au present poli.", Level: "A1", Order: 5, Content: "Utilise -아요/-어요 pour un ton poli : 학교에 가요. 한국어를 공부해요. 물을 마셔요."},
	}

	for _, lesson := range lessons {
		if _, err := store.db.Exec(`
			INSERT INTO lessons (id, title, description, level, sort_order, content)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				title = excluded.title,
				description = excluded.description,
				level = excluded.level,
				sort_order = excluded.sort_order,
				content = excluded.content
		`, lesson.ID, lesson.Title, lesson.Description, lesson.Level, lesson.Order, lesson.Content); err != nil {
			return err
		}
	}
	return nil
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

func (store *Store) ListLessons(ctx context.Context, userID string) ([]core.Lesson, []core.LessonProgress, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, title, description, level, sort_order, content
		FROM lessons
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	lessons := make([]core.Lesson, 0)
	for rows.Next() {
		var lesson core.Lesson
		if err := rows.Scan(&lesson.ID, &lesson.Title, &lesson.Description, &lesson.Level, &lesson.Order, &lesson.Content); err != nil {
			return nil, nil, err
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	progressRows, err := store.db.QueryContext(ctx, `
		SELECT user_id, lesson_id, completed, score, updated_at
		FROM lesson_progress
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer progressRows.Close()

	progress := make([]core.LessonProgress, 0)
	for progressRows.Next() {
		var item core.LessonProgress
		var completed int
		var updatedAt string
		if err := progressRows.Scan(&item.UserID, &item.LessonID, &completed, &item.Score, &updatedAt); err != nil {
			return nil, nil, err
		}
		item.Completed = completed == 1
		item.UpdatedAt = parseTime(updatedAt)
		progress = append(progress, item)
	}
	return lessons, progress, progressRows.Err()
}

func (store *Store) FindLessonByID(ctx context.Context, userID string, id string) (core.Lesson, core.LessonProgress, error) {
	var lesson core.Lesson
	err := store.db.QueryRowContext(ctx, `
		SELECT id, title, description, level, sort_order, content FROM lessons WHERE id = ?
	`, id).Scan(&lesson.ID, &lesson.Title, &lesson.Description, &lesson.Level, &lesson.Order, &lesson.Content)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Lesson{}, core.LessonProgress{}, repository.ErrNotFound
	}
	if err != nil {
		return core.Lesson{}, core.LessonProgress{}, err
	}

	progress := core.LessonProgress{UserID: userID, LessonID: id}
	var completed int
	var updatedAt string
	err = store.db.QueryRowContext(ctx, `
		SELECT completed, score, updated_at FROM lesson_progress WHERE user_id = ? AND lesson_id = ?
	`, userID, id).Scan(&completed, &progress.Score, &updatedAt)
	if err == nil {
		progress.Completed = completed == 1
		progress.UpdatedAt = parseTime(updatedAt)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return core.Lesson{}, core.LessonProgress{}, err
	}
	return lesson, progress, nil
}

func (store *Store) UpsertLessonProgress(ctx context.Context, progress core.LessonProgress) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO lesson_progress (user_id, lesson_id, completed, score, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, lesson_id) DO UPDATE SET
			completed = excluded.completed,
			score = excluded.score,
			updated_at = excluded.updated_at
	`, progress.UserID, progress.LessonID, boolInt(progress.Completed), progress.Score, formatTime(progress.UpdatedAt))
	return err
}

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

func requireAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrNotFound
	}
	return nil
}
