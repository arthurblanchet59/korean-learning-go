package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"github.com/arthurblanchet59/korean-learning-go/packages/core"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
	"strings"
	"time"
	"unicode"
)

type LessonWithProgress struct {
	core.Lesson
	Progress core.LessonProgress `json:"progress"`
}

func (service *StudyService) ListLessons(ctx context.Context, userID string) ([]LessonWithProgress, error) {
	lessons, progress, err := service.lessons.ListLessons(ctx, userID)
	if err != nil {
		return nil, err
	}
	byLesson := make(map[string]core.LessonProgress, len(progress))
	for _, item := range progress {
		byLesson[item.LessonID] = item
	}
	result := make([]LessonWithProgress, 0, len(lessons))
	for _, lesson := range lessons {
		item := byLesson[lesson.ID]
		item.UserID = userID
		item.LessonID = lesson.ID
		result = append(result, LessonWithProgress{Lesson: lesson, Progress: item})
	}
	return result, nil
}

func (service *StudyService) LessonByID(ctx context.Context, userID string, id string) (LessonWithProgress, error) {
	lesson, progress, err := service.lessons.FindLessonByID(ctx, userID, strings.TrimSpace(id))
	return LessonWithProgress{Lesson: lesson, Progress: progress}, err
}

func (service *StudyService) UpdateLessonProgress(ctx context.Context, userID string, lessonID string, completed bool, score int) (core.LessonProgress, error) {
	if score < 0 || score > 100 {
		return core.LessonProgress{}, validationErrorf("score must be between 0 and 100")
	}
	if _, _, err := service.lessons.FindLessonByID(ctx, userID, lessonID); err != nil {
		return core.LessonProgress{}, err
	}
	progress := core.LessonProgress{UserID: userID, LessonID: lessonID, Completed: completed, Score: score, UpdatedAt: service.now()}
	if err := service.lessons.UpsertLessonProgress(ctx, progress); err != nil {
		return core.LessonProgress{}, err
	}
	return progress, nil
}

type JournalInput struct {
	Title string
	Text  string
}

func (service *StudyService) ListJournalEntries(ctx context.Context, userID string) ([]core.JournalEntry, error) {
	return service.journal.ListJournalEntries(ctx, userID)
}

func (service *StudyService) JournalEntryByID(ctx context.Context, userID string, id string) (core.JournalEntry, error) {
	return service.journal.FindJournalEntryByID(ctx, userID, strings.TrimSpace(id))
}

func (service *StudyService) CreateJournalEntry(ctx context.Context, userID string, input JournalInput) (core.JournalEntry, error) {
	if strings.TrimSpace(input.Text) == "" {
		return core.JournalEntry{}, validationErrorf("journal text is required")
	}
	corrected, corrections := CorrectKorean(input.Text)
	now := service.now()
	entry := core.JournalEntry{ID: uuid.NewString(), UserID: userID, Title: journalTitle(input.Title, now), OriginalText: strings.TrimSpace(input.Text), CorrectedText: corrected, Corrections: corrections, CreatedAt: now, UpdatedAt: now}
	if err := service.journal.CreateJournalEntry(ctx, entry); err != nil {
		return core.JournalEntry{}, err
	}
	return entry, nil
}

func (service *StudyService) UpdateJournalEntry(ctx context.Context, userID string, id string, input JournalInput) (core.JournalEntry, error) {
	entry, err := service.journal.FindJournalEntryByID(ctx, userID, id)
	if err != nil {
		return core.JournalEntry{}, err
	}
	if strings.TrimSpace(input.Text) == "" {
		return core.JournalEntry{}, validationErrorf("journal text is required")
	}
	entry.Title = journalTitle(input.Title, entry.CreatedAt)
	entry.OriginalText = strings.TrimSpace(input.Text)
	entry.CorrectedText, entry.Corrections = CorrectKorean(entry.OriginalText)
	entry.UpdatedAt = service.now()
	if err := service.journal.UpdateJournalEntry(ctx, entry); err != nil {
		return core.JournalEntry{}, err
	}
	return entry, nil
}

func (service *StudyService) DeleteJournalEntry(ctx context.Context, userID string, id string) error {
	return service.journal.DeleteJournalEntry(ctx, userID, strings.TrimSpace(id))
}

func journalTitle(title string, date time.Time) string {
	if title = strings.TrimSpace(title); title != "" {
		return title
	}
	return "Journal du " + date.Format("02/01/2006")
}

func (service *StudyService) CheckAnswer(ctx context.Context, userID string, cardID string, answer string, direction string) (core.AnswerCheck, error) {
	card, err := service.cards.FindCardByID(ctx, userID, cardID)
	if err != nil {
		return core.AnswerCheck{}, err
	}

	direction = strings.TrimSpace(direction)
	if direction == "" {
		direction = "korean-to-french"
	}
	var expected string
	switch direction {
	case "korean-to-french":
		expected = card.Translation
	case "french-to-korean":
		expected = card.Korean
	default:
		return core.AnswerCheck{}, validationErrorf("direction must be korean-to-french or french-to-korean")
	}

	return core.AnswerCheck{
		Correct:   normalizeAnswer(answer) == normalizeAnswer(expected),
		Expected:  expected,
		Submitted: strings.TrimSpace(answer),
		Direction: direction,
	}, nil
}

func (service *StudyService) ExportCardsCSV(ctx context.Context, userID string) (string, error) {
	cards, err := service.cards.ListCards(ctx, userID)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	_ = writer.Write([]string{"deckId", "kind", "korean", "translation", "romanization", "exampleKorean", "exampleTranslation", "tags"})
	for _, card := range cards {
		_ = writer.Write([]string{card.DeckID, string(card.Kind), card.Korean, card.Translation, card.Romanization, card.ExampleKorean, card.ExampleTranslation, strings.Join(card.Tags, "|")})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return output.String(), nil
}

func (service *StudyService) ImportCardsCSV(ctx context.Context, userID string, defaultDeckID string, content string) ([]core.Card, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, validationErrorf("invalid csv: %v", err)
	}
	if len(records) < 2 {
		return nil, validationErrorf("csv must contain a header and at least one card")
	}

	headers := map[string]int{}
	for index, header := range records[0] {
		headers[strings.ToLower(strings.TrimSpace(header))] = index
	}
	value := func(record []string, name string) string {
		index, ok := headers[strings.ToLower(name)]
		if !ok || index >= len(record) {
			return ""
		}
		return record[index]
	}

	created := make([]core.Card, 0, len(records)-1)
	for rowIndex, record := range records[1:] {
		deckID := strings.TrimSpace(value(record, "deckId"))
		if deckID == "" {
			deckID = defaultDeckID
		}
		kind := strings.TrimSpace(value(record, "kind"))
		if kind == "" {
			kind = string(core.CardKindVocabulary)
		}
		card, err := service.cardFromInput(ctx, userID, CardInput{
			DeckID:             deckID,
			Kind:               core.CardKind(kind),
			Korean:             value(record, "korean"),
			Translation:        value(record, "translation"),
			Romanization:       value(record, "romanization"),
			ExampleKorean:      value(record, "exampleKorean"),
			ExampleTranslation: value(record, "exampleTranslation"),
			Tags:               strings.FieldsFunc(value(record, "tags"), func(r rune) bool { return r == '|' || r == ',' }),
		})
		if err != nil {
			return nil, validationErrorf("csv row %d: %v", rowIndex+2, err)
		}
		card.ID = uuid.NewString()
		card.UserID = userID
		card.CreatedAt = service.now()
		card.ReviewState = core.NewState(service.now())
		created = append(created, card)
	}
	if err := service.cards.CreateCards(ctx, userID, created); err != nil {
		return nil, err
	}
	return created, nil
}

func normalizeAnswer(value string) string {
	value = norm.NFD.String(strings.ToLower(strings.TrimSpace(value)))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.In(r, unicode.Hangul) {
			return r
		}
		return -1
	}, value)
}
