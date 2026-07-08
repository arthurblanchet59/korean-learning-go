package core

import "fmt"

type Rating string

const (
	RatingAgain Rating = "again"
	RatingHard  Rating = "hard"
	RatingGood  Rating = "good"
	RatingEasy  Rating = "easy"
)

func ParseRating(value string) (Rating, error) {
	switch Rating(value) {
	case RatingAgain, RatingHard, RatingGood, RatingEasy:
		return Rating(value), nil
	default:
		return "", fmt.Errorf("unknown rating %q", value)
	}
}
