package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

func TestDecksAreScopedByUser(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := store.CreateDeck(ctx, core.Deck{ID: "a-deck", UserID: "user-a", Name: "A", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateDeck(ctx, core.Deck{ID: "b-deck", UserID: "user-b", Name: "B", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	decksA, err := store.ListDecks(ctx, "user-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(decksA) != 1 || decksA[0].ID != "a-deck" {
		t.Fatalf("user-a received another user's decks: %#v", decksA)
	}
	if _, err := store.FindDeckByID(ctx, "user-a", "b-deck"); err == nil {
		t.Fatal("user-a should not access user-b deck")
	}
}

func TestSeedUserCreatesIndependentStarterDeck(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}
	if err := store.SeedUser(context.Background(), "learner"); err != nil {
		t.Fatal(err)
	}
	decks, _ := store.ListDecks(context.Background(), "learner")
	cards, _ := store.ListCards(context.Background(), "learner")
	if len(decks) != len(core.SeedDecks(time.Now())) || len(cards) != len(core.SeedCards(time.Now())) {
		t.Fatalf("expected starter content, got %d decks and %d cards", len(decks), len(cards))
	}
	if err := store.SeedUser(context.Background(), "learner"); err != nil {
		t.Fatal(err)
	}
	decks, _ = store.ListDecks(context.Background(), "learner")
	cards, _ = store.ListCards(context.Background(), "learner")
	if len(decks) != len(core.SeedDecks(time.Now())) || len(cards) != len(core.SeedCards(time.Now())) {
		t.Fatalf("seed should be idempotent, got %d decks and %d cards", len(decks), len(cards))
	}
}
