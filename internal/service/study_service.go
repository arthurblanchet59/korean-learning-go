package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/arthurblanchet59/korean-learning-go/internal/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type StudyService struct {
	decks     repository.DeckRepository
	cards     repository.CardRepository
	reviews   repository.ReviewRepository
	scheduler core.Scheduler
	now       func() time.Time
}

type DeckInput struct {
	Name        string
	Description string
}

type CardInput struct {
	DeckID             string
	Kind               core.CardKind
	Korean             string
	Translation        string
	Romanization       string
	ExampleKorean      string
	ExampleTranslation string
	Tags               []string
}

type DeckPatchInput struct {
	Name        *string
	Description *string
}

type CardPatchInput struct {
	DeckID             *string
	Kind               *core.CardKind
	Korean             *string
	Translation        *string
	Romanization       *string
	ExampleKorean      *string
	ExampleTranslation *string
	Tags               *[]string
}

type SearchResult struct {
	Decks []core.Deck `json:"decks"`
	Cards []core.Card `json:"cards"`
}

func NewStudyService(decks repository.DeckRepository, cards repository.CardRepository, reviews repository.ReviewRepository, scheduler core.Scheduler) *StudyService {
	return &StudyService{
		decks:     decks,
		cards:     cards,
		reviews:   reviews,
		scheduler: scheduler,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (service *StudyService) ListDecks(ctx context.Context) ([]core.Deck, error) {
	return service.decks.ListDecks(ctx)
}

func (service *StudyService) SearchDecks(ctx context.Context, query string) ([]core.Deck, error) {
	return service.decks.SearchDecks(ctx, strings.TrimSpace(query))
}

func (service *StudyService) SearchAll(ctx context.Context, query string) (SearchResult, error) {
	decks, err := service.SearchDecks(ctx, query)
	if err != nil {
		return SearchResult{}, err
	}

	cards, err := service.SearchCards(ctx, query)
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{Decks: decks, Cards: cards}, nil
}

func (service *StudyService) DeckByID(ctx context.Context, id string) (core.Deck, error) {
	return service.decks.FindDeckByID(ctx, strings.TrimSpace(id))
}

func (service *StudyService) CreateDeck(ctx context.Context, input DeckInput) (core.Deck, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return core.Deck{}, fmt.Errorf("deck name is required")
	}

	deck := core.Deck{
		ID:          uuid.NewString(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   service.now(),
	}
	if err := service.decks.CreateDeck(ctx, deck); err != nil {
		return core.Deck{}, err
	}

	return deck, nil
}

func (service *StudyService) UpdateDeck(ctx context.Context, id string, input DeckPatchInput) (core.Deck, error) {
	deck, err := service.decks.FindDeckByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return core.Deck{}, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return core.Deck{}, fmt.Errorf("deck name cannot be empty")
		}
		deck.Name = name
	}
	if input.Description != nil {
		deck.Description = strings.TrimSpace(*input.Description)
	}

	if err := service.decks.UpdateDeck(ctx, deck); err != nil {
		return core.Deck{}, err
	}

	return deck, nil
}

func (service *StudyService) UpdateDecks(ctx context.Context, ids []string, input DeckPatchInput) ([]core.Deck, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("at least one id is required")
	}

	updated := make([]core.Deck, 0, len(cleaned))
	for _, id := range cleaned {
		deck, err := service.UpdateDeck(ctx, id, input)
		if err != nil {
			return nil, err
		}
		updated = append(updated, deck)
	}

	return updated, nil
}

func (service *StudyService) DeleteDeck(ctx context.Context, id string) error {
	return service.decks.DeleteDeck(ctx, strings.TrimSpace(id))
}

func (service *StudyService) DeleteDecks(ctx context.Context, ids []string) (int, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return 0, fmt.Errorf("at least one id is required")
	}

	return service.decks.DeleteDecks(ctx, cleaned)
}

func (service *StudyService) ListCards(ctx context.Context) ([]core.Card, error) {
	return service.cards.ListCards(ctx)
}

func (service *StudyService) SearchCards(ctx context.Context, query string) ([]core.Card, error) {
	return service.cards.SearchCards(ctx, strings.TrimSpace(query))
}

func (service *StudyService) CardByID(ctx context.Context, id string) (core.Card, error) {
	return service.cards.FindCardByID(ctx, strings.TrimSpace(id))
}

func (service *StudyService) CreateCard(ctx context.Context, input CardInput) (core.Card, error) {
	card, err := service.cardFromInput(ctx, input)
	if err != nil {
		return core.Card{}, err
	}
	card.ID = uuid.NewString()
	card.CreatedAt = service.now()
	card.ReviewState = core.NewState(service.now())

	if err := service.cards.CreateCard(ctx, card); err != nil {
		return core.Card{}, err
	}

	return card, nil
}

func (service *StudyService) UpdateCard(ctx context.Context, id string, input CardPatchInput) (core.Card, error) {
	card, err := service.cards.FindCardByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return core.Card{}, err
	}

	if input.DeckID != nil {
		deckID := strings.TrimSpace(*input.DeckID)
		if deckID == "" {
			return core.Card{}, fmt.Errorf("deckId cannot be empty")
		}
		if _, err := service.decks.FindDeckByID(ctx, deckID); err != nil {
			return core.Card{}, err
		}
		card.DeckID = deckID
	}
	if input.Kind != nil {
		if err := validateCardKind(*input.Kind); err != nil {
			return core.Card{}, err
		}
		card.Kind = *input.Kind
	}
	if input.Korean != nil {
		card.Korean = strings.TrimSpace(*input.Korean)
	}
	if input.Translation != nil {
		card.Translation = strings.TrimSpace(*input.Translation)
	}
	if input.Romanization != nil {
		card.Romanization = strings.TrimSpace(*input.Romanization)
	}
	if input.ExampleKorean != nil {
		card.ExampleKorean = strings.TrimSpace(*input.ExampleKorean)
	}
	if input.ExampleTranslation != nil {
		card.ExampleTranslation = strings.TrimSpace(*input.ExampleTranslation)
	}
	if input.Tags != nil {
		card.Tags = cleanTags(*input.Tags)
	}
	if card.Korean == "" || card.Translation == "" {
		return core.Card{}, fmt.Errorf("korean and translation are required")
	}

	if err := service.cards.UpdateCard(ctx, card); err != nil {
		return core.Card{}, err
	}

	return card, nil
}

func (service *StudyService) UpdateCards(ctx context.Context, ids []string, input CardPatchInput) ([]core.Card, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("at least one id is required")
	}

	updated := make([]core.Card, 0, len(cleaned))
	for _, id := range cleaned {
		card, err := service.UpdateCard(ctx, id, input)
		if err != nil {
			return nil, err
		}
		updated = append(updated, card)
	}

	return updated, nil
}

func (service *StudyService) DeleteCard(ctx context.Context, id string) error {
	return service.cards.DeleteCard(ctx, strings.TrimSpace(id))
}

func (service *StudyService) DeleteCards(ctx context.Context, ids []string) (int, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return 0, fmt.Errorf("at least one id is required")
	}

	return service.cards.DeleteCards(ctx, cleaned)
}

func (service *StudyService) DueCards(ctx context.Context) ([]core.Card, error) {
	return service.cards.ListDueCards(ctx, service.now())
}

func (service *StudyService) DifficultCards(ctx context.Context) ([]core.Card, error) {
	cards, err := service.cards.ListCards(ctx)
	if err != nil {
		return nil, err
	}

	difficult := make([]core.Card, 0)
	for _, card := range cards {
		if card.ReviewState.LapseCount >= 2 {
			difficult = append(difficult, card)
		}
	}

	return difficult, nil
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

	reviewedAt := service.now()
	previous := card.ReviewState
	next := service.scheduler.Schedule(previous, rating, reviewedAt)
	card.ReviewState = next

	if err := service.cards.UpdateCard(ctx, card); err != nil {
		return core.Review{}, err
	}

	review := core.Review{
		ID:         fmt.Sprintf("review-%s-%d", cardID, reviewedAt.UnixNano()),
		CardID:     cardID,
		Rating:     rating,
		ReviewedAt: next.LastReviewAt,
		Previous:   previous,
		Next:       next,
	}

	if err := service.reviews.CreateReview(ctx, review); err != nil {
		return core.Review{}, err
	}

	return review, nil
}

func (service *StudyService) cardFromInput(ctx context.Context, input CardInput) (core.Card, error) {
	deckID := strings.TrimSpace(input.DeckID)
	if deckID == "" {
		return core.Card{}, fmt.Errorf("deckId is required")
	}
	if _, err := service.decks.FindDeckByID(ctx, deckID); err != nil {
		return core.Card{}, err
	}
	if err := validateCardKind(input.Kind); err != nil {
		return core.Card{}, err
	}

	card := core.Card{
		DeckID:             deckID,
		Kind:               input.Kind,
		Korean:             strings.TrimSpace(input.Korean),
		Translation:        strings.TrimSpace(input.Translation),
		Romanization:       strings.TrimSpace(input.Romanization),
		ExampleKorean:      strings.TrimSpace(input.ExampleKorean),
		ExampleTranslation: strings.TrimSpace(input.ExampleTranslation),
		Tags:               cleanTags(input.Tags),
	}
	if card.Korean == "" || card.Translation == "" {
		return core.Card{}, fmt.Errorf("korean and translation are required")
	}

	return card, nil
}

func validateCardKind(kind core.CardKind) error {
	switch kind {
	case core.CardKindVocabulary, core.CardKindPhrase, core.CardKindHangul:
		return nil
	default:
		return fmt.Errorf("invalid card kind")
	}
}

func cleanIDs(ids []string) []string {
	cleaned := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			cleaned = append(cleaned, id)
		}
	}
	return cleaned
}

func cleanTags(tags []string) []string {
	cleaned := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !seen[tag] {
			seen[tag] = true
			cleaned = append(cleaned, tag)
		}
	}
	return cleaned
}
