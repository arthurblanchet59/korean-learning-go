package core

import "time"

type StudyStats struct {
	TotalCards     int `json:"totalCards"`
	DueCards       int `json:"dueCards"`
	NewCards       int `json:"newCards"`
	DifficultCards int `json:"difficultCards"`
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
	}

	return stats
}
