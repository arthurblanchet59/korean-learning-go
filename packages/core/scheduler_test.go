package core

import (
	"testing"
	"time"
)

func TestScheduleAgainReturnsSoon(t *testing.T) {
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	scheduler := NewScheduler()
	state := NewState(now)

	next := scheduler.Schedule(state, RatingAgain, now)

	if next.LapseCount != 1 {
		t.Fatalf("expected one lapse, got %d", next.LapseCount)
	}
	if next.NextReviewAt.Sub(now) != scheduler.AgainDelay {
		t.Fatalf("expected again delay %s, got %s", scheduler.AgainDelay, next.NextReviewAt.Sub(now))
	}
}

func TestScheduleGoodIncreasesInterval(t *testing.T) {
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	scheduler := NewScheduler()
	state := NewState(now)

	first := scheduler.Schedule(state, RatingGood, now)
	second := scheduler.Schedule(first, RatingGood, now.AddDate(0, 0, 1))

	if first.IntervalDays != 1 {
		t.Fatalf("expected first good interval to be 1 day, got %d", first.IntervalDays)
	}
	if second.IntervalDays <= first.IntervalDays {
		t.Fatalf("expected interval to increase, got first=%d second=%d", first.IntervalDays, second.IntervalDays)
	}
}
