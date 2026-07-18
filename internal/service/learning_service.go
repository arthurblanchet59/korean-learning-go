package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/arthurblanchet59/korean-learning-go/packages/core"
)

type LessonWithProgress struct {
	core.Lesson
	Progress core.LessonProgress `json:"progress"`
}

type JournalInput struct {
	Title string
	Text  string
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
		return core.AnswerCheck{}, fmt.Errorf("direction must be korean-to-french or french-to-korean")
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
		return nil, fmt.Errorf("invalid csv: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv must contain a header and at least one card")
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
			return nil, fmt.Errorf("csv row %d: %w", rowIndex+2, err)
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
		return core.LessonProgress{}, fmt.Errorf("score must be between 0 and 100")
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

func (service *StudyService) ListJournalEntries(ctx context.Context, userID string) ([]core.JournalEntry, error) {
	entries, err := service.journal.ListJournalEntries(ctx, userID)
	if err != nil {
		return nil, err
	}
	for index := range entries {
		entries[index] = hideIrrelevantJournalSources(entries[index])
	}
	return entries, nil
}

func (service *StudyService) JournalEntryByID(ctx context.Context, userID string, id string) (core.JournalEntry, error) {
	entry, err := service.journal.FindJournalEntryByID(ctx, userID, strings.TrimSpace(id))
	if err != nil {
		return core.JournalEntry{}, err
	}
	return hideIrrelevantJournalSources(entry), nil
}

func (service *StudyService) CreateJournalEntry(ctx context.Context, userID string, input JournalInput) (core.JournalEntry, error) {
	if strings.TrimSpace(input.Text) == "" {
		return core.JournalEntry{}, fmt.Errorf("journal text is required")
	}
	correction, err := service.CorrectJournalText(ctx, input.Text)
	if err != nil {
		return core.JournalEntry{}, err
	}
	now := service.now()
	entry := core.JournalEntry{
		ID:            uuid.NewString(),
		UserID:        userID,
		Title:         journalTitle(input.Title, now),
		OriginalText:  strings.TrimSpace(input.Text),
		CorrectedText: correction.CorrectedText,
		Corrections:   correction.Corrections,
		Sources:       correction.Sources,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
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
		return core.JournalEntry{}, fmt.Errorf("journal text is required")
	}
	entry.Title = journalTitle(input.Title, entry.CreatedAt)
	entry.OriginalText = strings.TrimSpace(input.Text)
	correction, err := service.CorrectJournalText(ctx, entry.OriginalText)
	if err != nil {
		return core.JournalEntry{}, err
	}
	entry.CorrectedText = correction.CorrectedText
	entry.Corrections = correction.Corrections
	entry.Sources = correction.Sources
	entry.UpdatedAt = service.now()
	if err := service.journal.UpdateJournalEntry(ctx, entry); err != nil {
		return core.JournalEntry{}, err
	}
	return entry, nil
}

func (service *StudyService) CorrectJournalText(ctx context.Context, text string) (core.CorrectionResult, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return core.CorrectionResult{}, fmt.Errorf("journal text is required")
	}
	return service.corrector.Correct(ctx, text)
}

func hideIrrelevantJournalSources(entry core.JournalEntry) core.JournalEntry {
	if !containsHangul(entry.OriginalText) {
		entry.Sources = []core.CorrectionSource{}
	}
	return entry
}

func (service *StudyService) KnowledgeIndexStatus(ctx context.Context) (KnowledgeIndexStatus, error) {
	indexer, ok := service.corrector.(KnowledgeIndexer)
	if !ok {
		return KnowledgeIndexStatus{Enabled: false}, nil
	}
	return indexer.Status(ctx)
}

func (service *StudyService) EnsureKnowledgeIndex(ctx context.Context) (KnowledgeIndexStatus, error) {
	indexer, ok := service.corrector.(KnowledgeIndexer)
	if !ok {
		return KnowledgeIndexStatus{Enabled: false}, nil
	}
	return indexer.EnsureIndex(ctx)
}

func (service *StudyService) ReindexKnowledge(ctx context.Context) (KnowledgeIndexStatus, error) {
	indexer, ok := service.corrector.(KnowledgeIndexer)
	if !ok {
		return KnowledgeIndexStatus{Enabled: false}, fmt.Errorf("pedagogical RAG is not configured")
	}
	return indexer.Reindex(ctx)
}

func (service *StudyService) DeleteJournalEntry(ctx context.Context, userID string, id string) error {
	return service.journal.DeleteJournalEntry(ctx, userID, strings.TrimSpace(id))
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

func journalTitle(title string, date time.Time) string {
	if title = strings.TrimSpace(title); title != "" {
		return title
	}
	return "Journal du " + date.Format("02/01/2006")
}
