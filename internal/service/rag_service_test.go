package service

import (
	"context"
	"testing"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type knowledgeRepositoryStub struct {
	lessons      []core.Lesson
	index        domain.KnowledgeIndex
	chunks       []domain.KnowledgeChunk
	replaceCalls int
}

func (repositoryStub *knowledgeRepositoryStub) ListKnowledgeLessons(context.Context) ([]core.Lesson, error) {
	return repositoryStub.lessons, nil
}

func (repositoryStub *knowledgeRepositoryStub) FindKnowledgeIndex(context.Context, string) (domain.KnowledgeIndex, error) {
	if repositoryStub.index.EmbeddingModel == "" {
		return domain.KnowledgeIndex{}, repository.ErrNotFound
	}
	return repositoryStub.index, nil
}

func (repositoryStub *knowledgeRepositoryStub) ListKnowledgeChunks(context.Context, string) ([]domain.KnowledgeChunk, error) {
	return repositoryStub.chunks, nil
}

func (repositoryStub *knowledgeRepositoryStub) ReplaceKnowledgeIndex(_ context.Context, index domain.KnowledgeIndex, chunks []domain.KnowledgeChunk) error {
	repositoryStub.index = index
	repositoryStub.chunks = chunks
	repositoryStub.replaceCalls++
	return nil
}

type embeddingProviderStub struct {
	documentCalls int
	queryVector   []float64
}

func (*embeddingProviderStub) Model() string   { return "embed-v-4-0" }
func (*embeddingProviderStub) Dimensions() int { return 2 }

func (provider *embeddingProviderStub) EmbedDocuments(_ context.Context, texts []string) ([][]float64, error) {
	provider.documentCalls++
	result := make([][]float64, len(texts))
	for index := range texts {
		result[index] = []float64{1, float64(index)}
	}
	return result, nil
}

func (provider *embeddingProviderStub) EmbedQuery(context.Context, string) ([]float64, error) {
	return provider.queryVector, nil
}

type contextualCorrectorStub struct {
	sources []core.CorrectionSource
}

func (corrector *contextualCorrectorStub) CorrectWithContext(_ context.Context, input string, sources []core.CorrectionSource) (core.CorrectionResult, error) {
	corrector.sources = sources
	return core.CorrectionResult{CorrectedText: input, Corrections: []core.Correction{}, Sources: sources}, nil
}

func TestChunkLessonsKeepsLessonMetadata(t *testing.T) {
	lessons := []core.Lesson{{
		ID:          "grammar-topic",
		Title:       "Les particules",
		Description: "Choisir le theme.",
		Level:       "A1",
		Content:     "OBJECTIF\nComprendre les particules.\n\nREGLE\n은 et 는 indiquent le theme.",
	}}
	chunks := chunkLessons(lessons)
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}
	if chunks[0].SourceID != "grammar-topic" || chunks[0].Title != "Les particules" || chunks[0].Level != "A1" {
		t.Fatalf("unexpected chunk metadata: %+v", chunks[0])
	}
}

func TestEnsureIndexSkipsUnchangedCorpus(t *testing.T) {
	repositoryStub := &knowledgeRepositoryStub{lessons: []core.Lesson{{
		ID: "lesson", Title: "Lesson", Level: "A1", Content: "A short lesson.",
	}}}
	embedder := &embeddingProviderStub{}
	corrector := NewRAGCorrector(repositoryStub, embedder, &contextualCorrectorStub{})

	first, err := corrector.EnsureIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := corrector.EnsureIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !first.Ready || !second.Ready || repositoryStub.replaceCalls != 1 || embedder.documentCalls != 1 {
		t.Fatalf("unexpected indexing state: first=%+v second=%+v replacements=%d embeddings=%d", first, second, repositoryStub.replaceCalls, embedder.documentCalls)
	}
}

func TestRAGCorrectorRetrievesClosestLesson(t *testing.T) {
	repositoryStub := &knowledgeRepositoryStub{
		index: domain.KnowledgeIndex{EmbeddingModel: "embed-v-4-0", ChunkCount: 2, UpdatedAt: time.Now()},
		chunks: []domain.KnowledgeChunk{
			{ID: "particles-01", Title: "Particules", Level: "A1", Content: "은/는 indique le theme.", Embedding: []float64{1, 0}, EmbeddingModel: "embed-v-4-0"},
			{ID: "past-01", Title: "Passe", Level: "A1", Content: "Le passe utilise 았/었어요.", Embedding: []float64{0, 1}, EmbeddingModel: "embed-v-4-0"},
		},
	}
	embedder := &embeddingProviderStub{queryVector: []float64{0.9, 0.1}}
	generator := &contextualCorrectorStub{}
	corrector := NewRAGCorrector(repositoryStub, embedder, generator)

	result, err := corrector.Correct(context.Background(), "저 는 학생이에요")
	if err != nil {
		t.Fatal(err)
	}
	if len(generator.sources) != 2 || generator.sources[0].ID != "particles-01" {
		t.Fatalf("unexpected retrieval order: %+v", generator.sources)
	}
	if len(result.Sources) != 2 || result.Sources[0].Score <= result.Sources[1].Score {
		t.Fatalf("unexpected public sources: %+v", result.Sources)
	}
}

func TestCosineSimilarityHandlesInvalidVectors(t *testing.T) {
	if cosineSimilarity([]float64{1}, []float64{1, 2}) != 0 {
		t.Fatal("vectors with different dimensions must not be compared")
	}
	if score := cosineSimilarity([]float64{1, 0}, []float64{1, 0}); score != 1 {
		t.Fatalf("expected identical vectors to score 1, got %f", score)
	}
}
