package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

const (
	knowledgeChunkVersion = "lessons-v1"
	knowledgeBatchSize    = 64
	knowledgeResultCount  = 4
	knowledgeChunkRunes   = 1200
	sourceExcerptRunes    = 280
)

type EmbeddingProvider interface {
	Model() string
	Dimensions() int
	EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error)
	EmbedQuery(ctx context.Context, text string) ([]float64, error)
}

type ContextualKoreanCorrector interface {
	CorrectWithContext(ctx context.Context, input string, sources []core.CorrectionSource) (core.CorrectionResult, error)
}

type KnowledgeIndexStatus struct {
	Enabled        bool      `json:"enabled"`
	Ready          bool      `json:"ready"`
	EmbeddingModel string    `json:"embeddingModel,omitempty"`
	Dimensions     int       `json:"dimensions,omitempty"`
	ChunkCount     int       `json:"chunkCount"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty"`
}

type KnowledgeIndexer interface {
	EnsureIndex(ctx context.Context) (KnowledgeIndexStatus, error)
	Reindex(ctx context.Context) (KnowledgeIndexStatus, error)
	Status(ctx context.Context) (KnowledgeIndexStatus, error)
}

type RAGCorrector struct {
	repository repository.KnowledgeRepository
	embedder   EmbeddingProvider
	generator  ContextualKoreanCorrector
	indexMu    sync.Mutex
}

func NewRAGCorrector(repository repository.KnowledgeRepository, embedder EmbeddingProvider, generator ContextualKoreanCorrector) *RAGCorrector {
	return &RAGCorrector{repository: repository, embedder: embedder, generator: generator}
}

func (corrector *RAGCorrector) Correct(ctx context.Context, input string) (core.CorrectionResult, error) {
	if !containsHangul(input) {
		return corrector.generator.CorrectWithContext(ctx, input, nil)
	}

	chunks, err := corrector.repository.ListKnowledgeChunks(ctx, corrector.embedder.Model())
	if err != nil {
		return core.CorrectionResult{}, fmt.Errorf("load pedagogical index: %w", err)
	}
	if len(chunks) == 0 {
		return corrector.generator.CorrectWithContext(ctx, input, nil)
	}

	queryEmbedding, err := corrector.embedder.EmbedQuery(ctx, input)
	if err != nil {
		return core.CorrectionResult{}, err
	}
	sources := closestSources(queryEmbedding, chunks, knowledgeResultCount)
	result, err := corrector.generator.CorrectWithContext(ctx, input, sources)
	if err != nil {
		return core.CorrectionResult{}, err
	}
	result.Sources = compactSources(sources)
	return result, nil
}

func (corrector *RAGCorrector) EnsureIndex(ctx context.Context) (KnowledgeIndexStatus, error) {
	return corrector.buildIndex(ctx, false)
}

func (corrector *RAGCorrector) Reindex(ctx context.Context) (KnowledgeIndexStatus, error) {
	return corrector.buildIndex(ctx, true)
}

func (corrector *RAGCorrector) Status(ctx context.Context) (KnowledgeIndexStatus, error) {
	status := KnowledgeIndexStatus{
		Enabled:        true,
		EmbeddingModel: corrector.embedder.Model(),
		Dimensions:     corrector.embedder.Dimensions(),
	}
	index, err := corrector.repository.FindKnowledgeIndex(ctx, corrector.embedder.Model())
	if errors.Is(err, repository.ErrNotFound) {
		return status, nil
	}
	if err != nil {
		return KnowledgeIndexStatus{}, err
	}
	status.Ready = index.ChunkCount > 0
	status.ChunkCount = index.ChunkCount
	status.UpdatedAt = index.UpdatedAt
	return status, nil
}

func (corrector *RAGCorrector) buildIndex(ctx context.Context, force bool) (KnowledgeIndexStatus, error) {
	corrector.indexMu.Lock()
	defer corrector.indexMu.Unlock()

	lessons, err := corrector.repository.ListKnowledgeLessons(ctx)
	if err != nil {
		return KnowledgeIndexStatus{}, fmt.Errorf("load lessons for pedagogical index: %w", err)
	}
	chunks := chunkLessons(lessons)
	if len(chunks) == 0 {
		return KnowledgeIndexStatus{}, fmt.Errorf("cannot build pedagogical index from an empty lesson corpus")
	}
	corpusHash := knowledgeHash(chunks, corrector.embedder.Model(), corrector.embedder.Dimensions())
	if !force {
		current, err := corrector.repository.FindKnowledgeIndex(ctx, corrector.embedder.Model())
		switch {
		case err == nil && current.CorpusHash == corpusHash && current.ChunkCount == len(chunks):
			return KnowledgeIndexStatus{
				Enabled:        true,
				Ready:          true,
				EmbeddingModel: current.EmbeddingModel,
				Dimensions:     corrector.embedder.Dimensions(),
				ChunkCount:     current.ChunkCount,
				UpdatedAt:      current.UpdatedAt,
			}, nil
		case err != nil && !errors.Is(err, repository.ErrNotFound):
			return KnowledgeIndexStatus{}, err
		}
	}

	now := time.Now().UTC()
	for start := 0; start < len(chunks); start += knowledgeBatchSize {
		end := min(start+knowledgeBatchSize, len(chunks))
		texts := make([]string, 0, end-start)
		for _, chunk := range chunks[start:end] {
			texts = append(texts, chunk.Content)
		}
		embeddings, err := corrector.embedder.EmbedDocuments(ctx, texts)
		if err != nil {
			return KnowledgeIndexStatus{}, err
		}
		if len(embeddings) != len(texts) {
			return KnowledgeIndexStatus{}, fmt.Errorf("%w: invalid embedding count", ErrEmbeddingUnavailable)
		}
		for offset := range embeddings {
			chunk := &chunks[start+offset]
			chunk.Embedding = embeddings[offset]
			chunk.EmbeddingModel = corrector.embedder.Model()
			chunk.CorpusHash = corpusHash
			chunk.UpdatedAt = now
		}
	}

	index := domain.KnowledgeIndex{
		EmbeddingModel: corrector.embedder.Model(),
		CorpusHash:     corpusHash,
		ChunkCount:     len(chunks),
		UpdatedAt:      now,
	}
	if err := corrector.repository.ReplaceKnowledgeIndex(ctx, index, chunks); err != nil {
		return KnowledgeIndexStatus{}, fmt.Errorf("save pedagogical index: %w", err)
	}
	return KnowledgeIndexStatus{
		Enabled:        true,
		Ready:          true,
		EmbeddingModel: index.EmbeddingModel,
		Dimensions:     corrector.embedder.Dimensions(),
		ChunkCount:     index.ChunkCount,
		UpdatedAt:      index.UpdatedAt,
	}, nil
}

func chunkLessons(lessons []core.Lesson) []domain.KnowledgeChunk {
	chunks := make([]domain.KnowledgeChunk, 0, len(lessons)*3)
	for _, lesson := range lessons {
		chunkIndex := 0
		paragraphs := strings.Split(strings.ReplaceAll(lesson.Content, "\r\n", "\n"), "\n\n")
		current := strings.Builder{}
		flush := func() {
			content := strings.TrimSpace(current.String())
			current.Reset()
			if content == "" {
				return
			}
			chunkIndex++
			chunks = append(chunks, domain.KnowledgeChunk{
				ID:       fmt.Sprintf("%s-%02d", lesson.ID, chunkIndex),
				SourceID: lesson.ID,
				Title:    lesson.Title,
				Level:    lesson.Level,
				Content:  fmt.Sprintf("%s\nNiveau %s\n%s\n\n%s", lesson.Title, lesson.Level, lesson.Description, content),
			})
		}
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}
			if current.Len() > 0 && utf8.RuneCountInString(current.String())+utf8.RuneCountInString(paragraph) > knowledgeChunkRunes {
				flush()
			}
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(paragraph)
		}
		flush()
	}
	return chunks
}

func knowledgeHash(chunks []domain.KnowledgeChunk, model string, dimensions int) string {
	hash := sha256.New()
	fmt.Fprintf(hash, "%s\n%s\n%d\n", knowledgeChunkVersion, model, dimensions)
	for _, chunk := range chunks {
		fmt.Fprintf(hash, "%s\n%s\n%s\n", chunk.ID, chunk.SourceID, chunk.Content)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func closestSources(query []float64, chunks []domain.KnowledgeChunk, limit int) []core.CorrectionSource {
	sources := make([]core.CorrectionSource, 0, len(chunks))
	for _, chunk := range chunks {
		score := cosineSimilarity(query, chunk.Embedding)
		sources = append(sources, core.CorrectionSource{
			ID:      chunk.ID,
			Title:   chunk.Title,
			Level:   chunk.Level,
			Excerpt: chunk.Content,
			Score:   score,
		})
	}
	sort.SliceStable(sources, func(left, right int) bool {
		return sources[left].Score > sources[right].Score
	})
	if limit > 0 && len(sources) > limit {
		sources = sources[:limit]
	}
	return sources
}

func cosineSimilarity(left []float64, right []float64) float64 {
	if len(left) == 0 || len(left) != len(right) {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		dot += left[index] * right[index]
		leftNorm += left[index] * left[index]
		rightNorm += right[index] * right[index]
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func compactSources(sources []core.CorrectionSource) []core.CorrectionSource {
	compacted := make([]core.CorrectionSource, len(sources))
	copy(compacted, sources)
	for index := range compacted {
		compacted[index].Excerpt = truncateRunes(compacted[index].Excerpt, sourceExcerptRunes)
		compacted[index].Score = math.Round(compacted[index].Score*1000) / 1000
	}
	return compacted
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}
