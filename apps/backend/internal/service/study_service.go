package service

import (
	"context"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/apps/backend/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type StudyService struct {
	decks     repository.DeckRepository
	cards     repository.CardRepository
	scheduler core.Scheduler
	now       func() time.Time
}

func NewStudyService(decks repository.DeckRepository, cards repository.CardRepository, scheduler core.Scheduler) *StudyService {
	return &StudyService{
		decks:     decks,
		cards:     cards,
		scheduler: scheduler,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (service *StudyService) ListDecks(ctx context.Context) ([]core.Deck, error) {
	return service.decks.ListDecks(ctx)
}

func (service *StudyService) ListCards(ctx context.Context) ([]core.Card, error) {
	return service.cards.ListCards(ctx)
}

func (service *StudyService) DueCards(ctx context.Context) ([]core.Card, error) {
	return service.cards.ListDueCards(ctx, service.now())
}

func (service *StudyService) Stats(ctx context.Context) (core.StudyStats, error) {
	cards, err := service.cards.ListCards(ctx)
	if err != nil {
		return core.StudyStats{}, err
	}

	return core.BuildStats(cards, service.now()), nil
}

func (service *StudyService) AnswerCard(ctx context.Context, cardID string, ratingValue string) (core.Review, error) {
	rating, err := core.ParseRating(ratingValue)
	if err != nil {
		return core.Review{}, err
	}

	card, err := service.cards.FindCardByID(ctx, cardID)
	if err != nil {
		return core.Review{}, err
	}

	previous := card.ReviewState
	next := service.scheduler.Schedule(previous, rating, service.now())
	card.ReviewState = next

	if err := service.cards.UpdateCard(ctx, card); err != nil {
		return core.Review{}, err
	}

	return core.Review{
		ID:         "review-" + cardID,
		CardID:     cardID,
		Rating:     rating,
		ReviewedAt: next.LastReviewAt,
		Previous:   previous,
		Next:       next,
	}, nil
}
