package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func (store *Store) ListKnowledgeLessons(ctx context.Context) ([]core.Lesson, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, title, description, level, sort_order, content
		FROM lessons
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := make([]core.Lesson, 0)
	for rows.Next() {
		var lesson core.Lesson
		if err := rows.Scan(&lesson.ID, &lesson.Title, &lesson.Description, &lesson.Level, &lesson.Order, &lesson.Content); err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, rows.Err()
}

func (store *Store) FindKnowledgeIndex(ctx context.Context, embeddingModel string) (domain.KnowledgeIndex, error) {
	var index domain.KnowledgeIndex
	var updatedAt string
	err := store.db.QueryRowContext(ctx, `
		SELECT embedding_model, corpus_hash, chunk_count, updated_at
		FROM knowledge_indexes
		WHERE embedding_model = ?
	`, embeddingModel).Scan(&index.EmbeddingModel, &index.CorpusHash, &index.ChunkCount, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.KnowledgeIndex{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.KnowledgeIndex{}, err
	}
	index.UpdatedAt = parseTime(updatedAt)
	return index, nil
}

func (store *Store) ListKnowledgeChunks(ctx context.Context, embeddingModel string) ([]domain.KnowledgeChunk, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT id, source_id, title, level, content, embedding, embedding_model, corpus_hash, updated_at
		FROM knowledge_chunks
		WHERE embedding_model = ?
		ORDER BY source_id, id
	`, embeddingModel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]domain.KnowledgeChunk, 0)
	for rows.Next() {
		var chunk domain.KnowledgeChunk
		var embeddingRaw, updatedAt string
		if err := rows.Scan(
			&chunk.ID,
			&chunk.SourceID,
			&chunk.Title,
			&chunk.Level,
			&chunk.Content,
			&embeddingRaw,
			&chunk.EmbeddingModel,
			&chunk.CorpusHash,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(embeddingRaw), &chunk.Embedding); err != nil {
			return nil, err
		}
		chunk.UpdatedAt = parseTime(updatedAt)
		chunks = append(chunks, chunk)
	}
	return chunks, rows.Err()
}

func (store *Store) ReplaceKnowledgeIndex(ctx context.Context, index domain.KnowledgeIndex, chunks []domain.KnowledgeChunk) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM knowledge_chunks WHERE embedding_model = ?`, index.EmbeddingModel); err != nil {
		return err
	}
	for _, chunk := range chunks {
		embedding, err := json.Marshal(chunk.Embedding)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO knowledge_chunks (
				id, source_id, title, level, content, embedding, embedding_model, corpus_hash, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			chunk.ID,
			chunk.SourceID,
			chunk.Title,
			chunk.Level,
			chunk.Content,
			string(embedding),
			chunk.EmbeddingModel,
			chunk.CorpusHash,
			formatTime(chunk.UpdatedAt),
		); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO knowledge_indexes (embedding_model, corpus_hash, chunk_count, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(embedding_model) DO UPDATE SET
			corpus_hash = excluded.corpus_hash,
			chunk_count = excluded.chunk_count,
			updated_at = excluded.updated_at
	`, index.EmbeddingModel, index.CorpusHash, index.ChunkCount, formatTime(index.UpdatedAt)); err != nil {
		return err
	}
	return tx.Commit()
}
