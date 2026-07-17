package core

import "time"

type Scheduler struct {
	AgainDelay time.Duration
	HardDelay  time.Duration
}

func NewScheduler() Scheduler {
	return Scheduler{
		AgainDelay: 10 * time.Minute,
		HardDelay:  1 * time.Hour,
	}
}

func (scheduler Scheduler) Schedule(previous State, rating Rating, now time.Time) State {
	next := previous
	next.LastReviewAt = now
	next.ReviewCount++

	if next.EaseFactor == 0 {
		next.EaseFactor = 2.5
	}

	switch rating {
	case RatingAgain:
		next.LapseCount++
		next.IntervalDays = 0
		next.EaseFactor = max(1.3, next.EaseFactor-0.2)
		next.NextReviewAt = now.Add(scheduler.AgainDelay)
	case RatingHard:
		next.EaseFactor = max(1.3, next.EaseFactor-0.15)
		if previous.IntervalDays <= 0 {
			next.IntervalDays = 0
			next.NextReviewAt = now.Add(scheduler.HardDelay)
		} else {
			next.IntervalDays = max(1, int(float64(previous.IntervalDays)*0.8))
			next.NextReviewAt = now.AddDate(0, 0, next.IntervalDays)
		}
	case RatingGood:
		next.IntervalDays = nextGoodInterval(previous.IntervalDays, next.EaseFactor)
		next.NextReviewAt = now.AddDate(0, 0, next.IntervalDays)
	case RatingEasy:
		next.EaseFactor += 0.15
		next.IntervalDays = nextEasyInterval(previous.IntervalDays, next.EaseFactor)
		next.NextReviewAt = now.AddDate(0, 0, next.IntervalDays)
	}

	return next
}

func nextGoodInterval(previous int, ease float64) int {
	if previous <= 0 {
		return 1
	}

	return max(previous+1, int(float64(previous)*ease))
}

func nextEasyInterval(previous int, ease float64) int {
	if previous <= 0 {
		return 4
	}

	return max(previous+3, int(float64(previous)*ease*1.3))
}
