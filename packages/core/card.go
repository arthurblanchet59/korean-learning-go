package core

import "time"

type CardKind string

const (
	CardKindVocabulary CardKind = "vocabulary"
	CardKindPhrase     CardKind = "phrase"
	CardKindHangul     CardKind = "hangul"
)

type Deck struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Card struct {
	ID                 string    `json:"id"`
	DeckID             string    `json:"deckId"`
	Kind               CardKind  `json:"kind"`
	Korean             string    `json:"korean"`
	Translation        string    `json:"translation"`
	Romanization       string    `json:"romanization"`
	ExampleKorean      string    `json:"exampleKorean"`
	ExampleTranslation string    `json:"exampleTranslation"`
	Tags               []string  `json:"tags"`
	CreatedAt          time.Time `json:"createdAt"`
	ReviewState        State     `json:"reviewState"`
}

type State struct {
	NextReviewAt time.Time `json:"nextReviewAt"`
	LastReviewAt time.Time `json:"lastReviewAt,omitempty"`
	IntervalDays int       `json:"intervalDays"`
	EaseFactor   float64   `json:"easeFactor"`
	ReviewCount  int       `json:"reviewCount"`
	LapseCount   int       `json:"lapseCount"`
}

func NewState(now time.Time) State {
	return State{
		NextReviewAt: now,
		IntervalDays: 0,
		EaseFactor:   2.5,
	}
}

func (card Card) Due(now time.Time) bool {
	return !card.ReviewState.NextReviewAt.After(now)
}
