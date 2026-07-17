package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func TestKnowledgeIndexIsReplacedAtomically(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "knowledge.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	now := time.Now().UTC()
	index := domain.KnowledgeIndex{EmbeddingModel: "embed-v-4-0", CorpusHash: "first", ChunkCount: 2, UpdatedAt: now}
	chunks := []domain.KnowledgeChunk{
		{ID: "lesson-01", SourceID: "lesson", Title: "Lesson", Level: "A1", Content: "First", Embedding: []float64{1, 0}, EmbeddingModel: index.EmbeddingModel, CorpusHash: index.CorpusHash, UpdatedAt: now},
		{ID: "lesson-02", SourceID: "lesson", Title: "Lesson", Level: "A1", Content: "Second", Embedding: []float64{0, 1}, EmbeddingModel: index.EmbeddingModel, CorpusHash: index.CorpusHash, UpdatedAt: now},
	}
	if err := store.ReplaceKnowledgeIndex(ctx, index, chunks); err != nil {
		t.Fatal(err)
	}

	index.CorpusHash = "second"
	index.ChunkCount = 1
	chunks = chunks[:1]
	chunks[0].CorpusHash = index.CorpusHash
	if err := store.ReplaceKnowledgeIndex(ctx, index, chunks); err != nil {
		t.Fatal(err)
	}

	savedIndex, err := store.FindKnowledgeIndex(ctx, index.EmbeddingModel)
	if err != nil {
		t.Fatal(err)
	}
	savedChunks, err := store.ListKnowledgeChunks(ctx, index.EmbeddingModel)
	if err != nil {
		t.Fatal(err)
	}
	if savedIndex.CorpusHash != "second" || savedIndex.ChunkCount != 1 || len(savedChunks) != 1 {
		t.Fatalf("unexpected saved index: index=%+v chunks=%+v", savedIndex, savedChunks)
	}
}

func TestJournalSourcesArePersisted(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "journal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	entry := core.JournalEntry{
		ID:            "entry",
		UserID:        "user",
		Title:         "Test",
		OriginalText:  "저는 학생이에요",
		CorrectedText: "저는 학생이에요.",
		Corrections:   []core.Correction{},
		Sources:       []core.CorrectionSource{{ID: "grammar-01", Title: "Les particules", Level: "A1", Excerpt: "은/는 indique le thème.", Score: 0.95}},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := store.CreateJournalEntry(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	saved, err := store.FindJournalEntryByID(context.Background(), "user", entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Sources) != 1 || saved.Sources[0].ID != "grammar-01" {
		t.Fatalf("unexpected persisted sources: %+v", saved.Sources)
	}
}
