package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/curriculum"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) seedLessons() error {

	for _, lesson := range curriculum.Lessons() {
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
