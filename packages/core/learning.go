package core

import "time"

type Lesson struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Level       string `json:"level"`
	Order       int    `json:"order"`
	Content     string `json:"content"`
}

type LessonProgress struct {
	UserID    string    `json:"-"`
	LessonID  string    `json:"lessonId"`
	Completed bool      `json:"completed"`
	Score     int       `json:"score"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Correction struct {
	Original    string `json:"original"`
	Replacement string `json:"replacement"`
	Reason      string `json:"reason"`
}

type JournalEntry struct {
	ID            string       `json:"id"`
	UserID        string       `json:"-"`
	Title         string       `json:"title"`
	OriginalText  string       `json:"originalText"`
	CorrectedText string       `json:"correctedText"`
	Corrections   []Correction `json:"corrections"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type AnswerCheck struct {
	Correct   bool   `json:"correct"`
	Expected  string `json:"expected"`
	Submitted string `json:"submitted"`
	Direction string `json:"direction"`
}
