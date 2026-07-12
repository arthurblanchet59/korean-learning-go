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

func TestCreateCardsRollsBackWholeBatch(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := store.CreateDeck(ctx, core.Deck{ID: "deck", UserID: "user", Name: "Deck", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	card := core.Card{ID: "duplicate", UserID: "user", DeckID: "deck", Kind: core.CardKindVocabulary, Korean: "집", Translation: "maison", CreatedAt: now, ReviewState: core.NewState(now)}
	if err := store.CreateCards(ctx, "user", []core.Card{card, card}); err == nil {
		t.Fatal("expected duplicate card batch to fail")
	}
	cards, err := store.ListCards(ctx, "user")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Fatalf("failed batch left %d card(s) behind", len(cards))
	}
}

func TestSaveReviewRollsBackCardWhenReviewInsertFails(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := store.CreateDeck(ctx, core.Deck{ID: "deck", UserID: "user", Name: "Deck", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	card := core.Card{ID: "card", UserID: "user", DeckID: "deck", Kind: core.CardKindVocabulary, Korean: "집", Translation: "maison", CreatedAt: now, ReviewState: core.NewState(now)}
	if err := store.CreateCard(ctx, card); err != nil {
		t.Fatal(err)
	}
	review := core.Review{ID: "review", UserID: "user", CardID: card.ID, Rating: core.RatingGood, ReviewedAt: now, Previous: card.ReviewState, Next: card.ReviewState}
	if err := store.CreateReview(ctx, review); err != nil {
		t.Fatal(err)
	}
	card.ReviewState.ReviewCount = 1
	card.ReviewState.IntervalDays = 10
	if err := store.SaveReview(ctx, "user", card, review); err == nil {
		t.Fatal("expected duplicate review id to fail")
	}
	stored, err := store.FindCardByID(ctx, "user", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ReviewState.ReviewCount != 0 || stored.ReviewState.IntervalDays != 0 {
		t.Fatalf("card state was not rolled back: %#v", stored.ReviewState)
	}
}
