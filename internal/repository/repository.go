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
	ListDecks(ctx context.Context, userID string) ([]core.Deck, error)
	SearchDecks(ctx context.Context, userID string, query string) ([]core.Deck, error)
	FindDeckByID(ctx context.Context, userID string, id string) (core.Deck, error)
	CreateDeck(ctx context.Context, deck core.Deck) error
	UpdateDeck(ctx context.Context, userID string, deck core.Deck) error
	UpdateDecks(ctx context.Context, userID string, decks []core.Deck) error
	DeleteDeck(ctx context.Context, userID string, id string) error
	DeleteDecks(ctx context.Context, userID string, ids []string) (int, error)
}

type CardRepository interface {
	ListCards(ctx context.Context, userID string) ([]core.Card, error)
	SearchCards(ctx context.Context, userID string, query string) ([]core.Card, error)
	ListDueCards(ctx context.Context, userID string, now time.Time) ([]core.Card, error)
	FindCardByID(ctx context.Context, userID string, id string) (core.Card, error)
	CreateCard(ctx context.Context, card core.Card) error
	CreateCards(ctx context.Context, userID string, cards []core.Card) error
	UpdateCard(ctx context.Context, userID string, card core.Card) error
	UpdateCards(ctx context.Context, userID string, cards []core.Card) error
	DeleteCard(ctx context.Context, userID string, id string) error
	DeleteCards(ctx context.Context, userID string, ids []string) (int, error)
}

type ReviewRepository interface {
	SaveReview(ctx context.Context, userID string, card core.Card, review core.Review) error
	ListReviewsSince(ctx context.Context, userID string, since time.Time) ([]core.Review, error)
}

type LessonRepository interface {
	ListLessons(ctx context.Context, userID string) ([]core.Lesson, []core.LessonProgress, error)
	FindLessonByID(ctx context.Context, userID string, id string) (core.Lesson, core.LessonProgress, error)
	UpsertLessonProgress(ctx context.Context, progress core.LessonProgress) error
}

type JournalRepository interface {
	ListJournalEntries(ctx context.Context, userID string) ([]core.JournalEntry, error)
	FindJournalEntryByID(ctx context.Context, userID string, id string) (core.JournalEntry, error)
	CreateJournalEntry(ctx context.Context, entry core.JournalEntry) error
	UpdateJournalEntry(ctx context.Context, entry core.JournalEntry) error
	DeleteJournalEntry(ctx context.Context, userID string, id string) error
}

type KnowledgeRepository interface {
	ListKnowledgeLessons(ctx context.Context) ([]core.Lesson, error)
	FindKnowledgeIndex(ctx context.Context, embeddingModel string) (domain.KnowledgeIndex, error)
	ListKnowledgeChunks(ctx context.Context, embeddingModel string) ([]domain.KnowledgeChunk, error)
	ReplaceKnowledgeIndex(ctx context.Context, index domain.KnowledgeIndex, chunks []domain.KnowledgeChunk) error
}

type ClientBackupRepository interface {
	FindClientBackup(ctx context.Context, userID string) (domain.ClientBackup, error)
	UpsertClientBackup(ctx context.Context, backup domain.ClientBackup) error
}

type UserDataSeeder interface {
	SeedUser(ctx context.Context, userID string) error
}

type UserRegistrationRepository interface {
	CreateUserWithSeed(ctx context.Context, user domain.User) error
}

type ResetResult struct {
	DeletedReviews  int `json:"deletedReviews"`
	DeletedCards    int `json:"deletedCards"`
	DeletedDecks    int `json:"deletedDecks"`
	DeletedUsers    int `json:"deletedUsers"`
	DeletedJournal  int `json:"deletedJournal"`
	DeletedProgress int `json:"deletedProgress"`
}

type AdminRepository interface {
	Reset(ctx context.Context) (ResetResult, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	ListUsers(ctx context.Context) ([]domain.User, error)
	FindUserByID(ctx context.Context, id string) (domain.User, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	EnsureAdmin(ctx context.Context, user domain.User) error
}
