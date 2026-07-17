package service

import (
	"context"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	"github.com/google/uuid"
	"strings"
	"time"
)

type StudyService struct {
	decks     repository.DeckRepository
	cards     repository.CardRepository
	reviews   repository.ReviewRepository
	lessons   repository.LessonRepository
	journal   repository.JournalRepository
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

func NewStudyService(decks repository.DeckRepository, cards repository.CardRepository, reviews repository.ReviewRepository, lessons repository.LessonRepository, journal repository.JournalRepository, scheduler core.Scheduler) *StudyService {
	return &StudyService{
		decks:     decks,
		cards:     cards,
		reviews:   reviews,
		lessons:   lessons,
		journal:   journal,
		scheduler: scheduler,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (service *StudyService) cardFromInput(ctx context.Context, userID string, input CardInput) (core.Card, error) {
	deckID := strings.TrimSpace(input.DeckID)
	if deckID == "" {
		return core.Card{}, validationErrorf("deckId is required")
	}
	if _, err := service.decks.FindDeckByID(ctx, userID, deckID); err != nil {
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
		return core.Card{}, validationErrorf("korean and translation are required")
	}

	return card, nil
}

func (service *StudyService) ListDecks(ctx context.Context, userID string) ([]core.Deck, error) {
	return service.decks.ListDecks(ctx, userID)
}

func (service *StudyService) SearchDecks(ctx context.Context, userID string, query string) ([]core.Deck, error) {
	return service.decks.SearchDecks(ctx, userID, strings.TrimSpace(query))
}

func (service *StudyService) DeckByID(ctx context.Context, userID string, id string) (core.Deck, error) {
	return service.decks.FindDeckByID(ctx, userID, strings.TrimSpace(id))
}

func (service *StudyService) CreateDeck(ctx context.Context, userID string, input DeckInput) (core.Deck, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return core.Deck{}, validationErrorf("deck name is required")
	}

	deck := core.Deck{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		CreatedAt:   service.now(),
	}
	if err := service.decks.CreateDeck(ctx, deck); err != nil {
		return core.Deck{}, err
	}

	return deck, nil
}

func (service *StudyService) UpdateDeck(ctx context.Context, userID string, id string, input DeckPatchInput) (core.Deck, error) {
	deck, err := service.decks.FindDeckByID(ctx, userID, strings.TrimSpace(id))
	if err != nil {
		return core.Deck{}, err
	}
	if err := applyDeckPatch(&deck, input); err != nil {
		return core.Deck{}, err
	}

	if err := service.decks.UpdateDeck(ctx, userID, deck); err != nil {
		return core.Deck{}, err
	}

	return deck, nil
}

func (service *StudyService) UpdateDecks(ctx context.Context, userID string, ids []string, input DeckPatchInput) ([]core.Deck, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return nil, validationErrorf("at least one id is required")
	}

	updated := make([]core.Deck, 0, len(cleaned))
	for _, id := range cleaned {
		deck, err := service.decks.FindDeckByID(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		if err := applyDeckPatch(&deck, input); err != nil {
			return nil, err
		}
		updated = append(updated, deck)
	}
	if err := service.decks.UpdateDecks(ctx, userID, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (service *StudyService) DeleteDeck(ctx context.Context, userID string, id string) error {
	return service.decks.DeleteDeck(ctx, userID, strings.TrimSpace(id))
}

func (service *StudyService) DeleteDecks(ctx context.Context, userID string, ids []string) (int, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return 0, validationErrorf("at least one id is required")
	}

	return service.decks.DeleteDecks(ctx, userID, cleaned)
}

func applyDeckPatch(deck *core.Deck, input DeckPatchInput) error {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return validationErrorf("deck name cannot be empty")
		}
		deck.Name = name
	}
	if input.Description != nil {
		deck.Description = strings.TrimSpace(*input.Description)
	}
	return nil
}

func (service *StudyService) SearchAll(ctx context.Context, userID string, query string) (SearchResult, error) {
	decks, err := service.SearchDecks(ctx, userID, query)
	if err != nil {
		return SearchResult{}, err
	}

	cards, err := service.SearchCards(ctx, userID, query)
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{Decks: decks, Cards: cards}, nil
}

func (service *StudyService) ListCards(ctx context.Context, userID string) ([]core.Card, error) {
	return service.cards.ListCards(ctx, userID)
}

func (service *StudyService) SearchCards(ctx context.Context, userID string, query string) ([]core.Card, error) {
	return service.cards.SearchCards(ctx, userID, strings.TrimSpace(query))
}

func (service *StudyService) CardByID(ctx context.Context, userID string, id string) (core.Card, error) {
	return service.cards.FindCardByID(ctx, userID, strings.TrimSpace(id))
}

func (service *StudyService) CreateCard(ctx context.Context, userID string, input CardInput) (core.Card, error) {
	card, err := service.cardFromInput(ctx, userID, input)
	if err != nil {
		return core.Card{}, err
	}
	card.ID = uuid.NewString()
	card.UserID = userID
	card.CreatedAt = service.now()
	card.ReviewState = core.NewState(service.now())

	if err := service.cards.CreateCard(ctx, card); err != nil {
		return core.Card{}, err
	}

	return card, nil
}

func (service *StudyService) UpdateCard(ctx context.Context, userID string, id string, input CardPatchInput) (core.Card, error) {
	card, err := service.cards.FindCardByID(ctx, userID, strings.TrimSpace(id))
	if err != nil {
		return core.Card{}, err
	}

	if err := service.applyCardPatch(ctx, userID, &card, input); err != nil {
		return core.Card{}, err
	}

	if err := service.cards.UpdateCard(ctx, userID, card); err != nil {
		return core.Card{}, err
	}

	return card, nil
}

func (service *StudyService) applyCardPatch(ctx context.Context, userID string, card *core.Card, input CardPatchInput) error {
	if input.DeckID != nil {
		deckID := strings.TrimSpace(*input.DeckID)
		if deckID == "" {
			return validationErrorf("deckId cannot be empty")
		}
		if _, err := service.decks.FindDeckByID(ctx, userID, deckID); err != nil {
			return err
		}
		card.DeckID = deckID
	}
	if input.Kind != nil {
		if err := validateCardKind(*input.Kind); err != nil {
			return err
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
		return validationErrorf("korean and translation are required")
	}
	return nil
}

func (service *StudyService) UpdateCards(ctx context.Context, userID string, ids []string, input CardPatchInput) ([]core.Card, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return nil, validationErrorf("at least one id is required")
	}

	updated := make([]core.Card, 0, len(cleaned))
	for _, id := range cleaned {
		card, err := service.cards.FindCardByID(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		if err := service.applyCardPatch(ctx, userID, &card, input); err != nil {
			return nil, err
		}
		updated = append(updated, card)
	}
	if err := service.cards.UpdateCards(ctx, userID, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (service *StudyService) DeleteCard(ctx context.Context, userID string, id string) error {
	return service.cards.DeleteCard(ctx, userID, strings.TrimSpace(id))
}

func (service *StudyService) DeleteCards(ctx context.Context, userID string, ids []string) (int, error) {
	cleaned := cleanIDs(ids)
	if len(cleaned) == 0 {
		return 0, validationErrorf("at least one id is required")
	}

	return service.cards.DeleteCards(ctx, userID, cleaned)
}

func validateCardKind(kind core.CardKind) error {
	switch kind {
	case core.CardKindVocabulary, core.CardKindPhrase, core.CardKindHangul:
		return nil
	default:
		return validationErrorf("invalid card kind")
	}
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

func (service *StudyService) DueCards(ctx context.Context, userID string) ([]core.Card, error) {
	return service.cards.ListDueCards(ctx, userID, service.now())
}

func (service *StudyService) DifficultCards(ctx context.Context, userID string) ([]core.Card, error) {
	cards, err := service.cards.ListCards(ctx, userID)
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

func (service *StudyService) Stats(ctx context.Context, userID string) (core.StudyStats, error) {
	cards, err := service.cards.ListCards(ctx, userID)
	if err != nil {
		return core.StudyStats{}, err
	}

	stats := core.BuildStats(cards, service.now())
	reviews, err := service.reviews.ListReviewsSince(ctx, userID, service.now().AddDate(0, 0, -90))
	if err != nil {
		return core.StudyStats{}, err
	}
	applyReviewStats(&stats, reviews, service.now())
	return stats, nil
}

func (service *StudyService) AnswerCard(ctx context.Context, userID string, cardID string, ratingValue string) (core.Review, error) {
	rating, err := core.ParseRating(ratingValue)
	if err != nil {
		return core.Review{}, err
	}

	card, err := service.cards.FindCardByID(ctx, userID, cardID)
	if err != nil {
		return core.Review{}, err
	}

	reviewedAt := service.now()
	previous := card.ReviewState
	next := service.scheduler.Schedule(previous, rating, reviewedAt)
	card.ReviewState = next

	review := core.Review{
		ID:         uuid.NewString(),
		UserID:     userID,
		CardID:     cardID,
		Rating:     rating,
		ReviewedAt: next.LastReviewAt,
		Previous:   previous,
		Next:       next,
	}

	if err := service.reviews.SaveReview(ctx, userID, card, review); err != nil {
		return core.Review{}, err
	}

	return review, nil
}

func applyReviewStats(stats *core.StudyStats, reviews []core.Review, now time.Time) {
	days := map[string]*core.DailyReviewStat{}
	correct, total := 0, 0
	for _, review := range reviews {
		date := review.ReviewedAt.UTC().Format("2006-01-02")
		item := days[date]
		if item == nil {
			item = &core.DailyReviewStat{Date: date}
			days[date] = item
		}
		item.Reviews++
		total++
		if review.Rating != core.RatingAgain {
			correct++
			item.Correct++
		}
		if date == now.UTC().Format("2006-01-02") {
			stats.ReviewsToday++
		}
	}
	if total > 0 {
		stats.AccuracyPercent = float64(correct) * 100 / float64(total)
	}
	for day := 89; day >= 0; day-- {
		date := now.UTC().AddDate(0, 0, -day).Format("2006-01-02")
		if item := days[date]; item != nil {
			stats.ReviewHistory = append(stats.ReviewHistory, *item)
		}
	}
	streak, longest := 0, 0
	for day := 0; day < 90; day++ {
		date := now.UTC().AddDate(0, 0, -day).Format("2006-01-02")
		if days[date] == nil {
			if day == 0 {
				continue
			}
			break
		}
		streak++
	}
	running := 0
	for day := 89; day >= 0; day-- {
		date := now.UTC().AddDate(0, 0, -day).Format("2006-01-02")
		if days[date] != nil {
			running++
			if running > longest {
				longest = running
			}
		} else {
			running = 0
		}
	}
	stats.CurrentStreak = streak
	stats.LongestStreak = longest
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
