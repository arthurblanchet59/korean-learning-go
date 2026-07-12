package core

import "time"

type StudyStats struct {
	TotalCards      int               `json:"totalCards"`
	DueCards        int               `json:"dueCards"`
	NewCards        int               `json:"newCards"`
	DifficultCards  int               `json:"difficultCards"`
	MasteredCards   int               `json:"masteredCards"`
	ReviewsToday    int               `json:"reviewsToday"`
	AccuracyPercent float64           `json:"accuracyPercent"`
	CurrentStreak   int               `json:"currentStreak"`
	LongestStreak   int               `json:"longestStreak"`
	ReviewHistory   []DailyReviewStat `json:"reviewHistory"`
}

type DailyReviewStat struct {
	Date    string `json:"date"`
	Reviews int    `json:"reviews"`
	Correct int    `json:"correct"`
}

func BuildStats(cards []Card, now time.Time) StudyStats {
	stats := StudyStats{TotalCards: len(cards)}

	for _, card := range cards {
		if card.Due(now) {
			stats.DueCards++
		}
		if card.ReviewState.ReviewCount == 0 {
			stats.NewCards++
		}
		if card.ReviewState.LapseCount >= 2 {
			stats.DifficultCards++
		}
		if card.ReviewState.IntervalDays >= 21 {
			stats.MasteredCards++
		}
	}

	return stats
}
