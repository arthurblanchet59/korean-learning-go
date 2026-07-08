package memory

import (
	"context"
	"sync"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type Store struct {
	mu    sync.RWMutex
	decks []core.Deck
	cards []core.Card
}

func NewStore(decks []core.Deck, cards []core.Card) *Store {
	return &Store{
		decks: append([]core.Deck(nil), decks...),
		cards: append([]core.Card(nil), cards...),
	}
}

func (store *Store) ListDecks(ctx context.Context) ([]core.Deck, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return append([]core.Deck(nil), store.decks...), nil
}

func (store *Store) ListCards(ctx context.Context) ([]core.Card, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return append([]core.Card(nil), store.cards...), nil
}

func (store *Store) ListDueCards(ctx context.Context, now time.Time) ([]core.Card, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	due := make([]core.Card, 0)
	for _, card := range store.cards {
		if card.Due(now) {
			due = append(due, card)
		}
	}

	return due, nil
}

func (store *Store) FindCardByID(ctx context.Context, id string) (core.Card, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, card := range store.cards {
		if card.ID == id {
			return card, nil
		}
	}

	return core.Card{}, repository.ErrNotFound
}

func (store *Store) UpdateCard(ctx context.Context, card core.Card) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	for index := range store.cards {
		if store.cards[index].ID == card.ID {
			store.cards[index] = card
			return nil
		}
	}

	return repository.ErrNotFound
}
