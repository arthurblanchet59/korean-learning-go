package repository

import (
	"context"
	"errors"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/domain"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

var ErrNotFound = errors.New("resource not found")
var ErrConflict = errors.New("resource already exists")

type DeckRepository interface {
	ListDecks(ctx context.Context) ([]core.Deck, error)
	SearchDecks(ctx context.Context, query string) ([]core.Deck, error)
	FindDeckByID(ctx context.Context, id string) (core.Deck, error)
	CreateDeck(ctx context.Context, deck core.Deck) error
	UpdateDeck(ctx context.Context, deck core.Deck) error
	DeleteDeck(ctx context.Context, id string) error
	DeleteDecks(ctx context.Context, ids []string) (int, error)
}

type CardRepository interface {
	ListCards(ctx context.Context) ([]core.Card, error)
	SearchCards(ctx context.Context, query string) ([]core.Card, error)
	ListDueCards(ctx context.Context, now time.Time) ([]core.Card, error)
	FindCardByID(ctx context.Context, id string) (core.Card, error)
	CreateCard(ctx context.Context, card core.Card) error
	UpdateCard(ctx context.Context, card core.Card) error
	DeleteCard(ctx context.Context, id string) error
	DeleteCards(ctx context.Context, ids []string) (int, error)
}

type ReviewRepository interface {
	CreateReview(ctx context.Context, review core.Review) error
}

type ResetResult struct {
	DeletedReviews int `json:"deletedReviews"`
	DeletedCards   int `json:"deletedCards"`
	DeletedDecks   int `json:"deletedDecks"`
	DeletedUsers   int `json:"deletedUsers"`
}

type AdminRepository interface {
	Reset(ctx context.Context) (ResetResult, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	FindUserByID(ctx context.Context, id string) (domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	EnsureAdmin(ctx context.Context, user domain.User) error
}
