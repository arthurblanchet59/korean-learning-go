package repository

import (
	"context"
	"errors"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

var ErrNotFound = errors.New("resource not found")

type DeckRepository interface {
	ListDecks(ctx context.Context) ([]core.Deck, error)
}

type CardRepository interface {
	ListCards(ctx context.Context) ([]core.Card, error)
	ListDueCards(ctx context.Context, now time.Time) ([]core.Card, error)
	FindCardByID(ctx context.Context, id string) (core.Card, error)
	UpdateCard(ctx context.Context, card core.Card) error
}

type ReviewRepository interface {
	CreateReview(ctx context.Context, review core.Review) error
}
