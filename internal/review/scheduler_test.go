package review

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

func TestRatingForSubmit(t *testing.T) {
	cases := []struct {
		accepted bool
		gaveUp   bool
		want     domain.Rating
	}{
		{accepted: true, gaveUp: false, want: domain.RatingGood},
		{accepted: true, gaveUp: true, want: domain.RatingAgain},
		{accepted: false, gaveUp: false, want: domain.RatingAgain},
	}
	for _, tc := range cases {
		if got := RatingForSubmit(tc.accepted, tc.gaveUp); got != tc.want {
			t.Fatalf("RatingForSubmit(%v,%v) = %v, want %v", tc.accepted, tc.gaveUp, got, tc.want)
		}
	}
}

func TestRateNewGoodCard(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	card := Rate(domain.Card{Slug: "two-sum"}, domain.RatingGood, now)

	if card.Slug != "two-sum" || card.Reps != 1 || card.Lapses != 0 {
		t.Fatalf("unexpected card counters: %+v", card)
	}
	if !card.LastReviewAt.Equal(now) {
		t.Fatalf("LastReviewAt = %v, want %v", card.LastReviewAt, now)
	}
	if got, want := card.DueAt.Sub(now), 24*time.Hour; got != want {
		t.Fatalf("DueAt interval = %v, want %v", got, want)
	}
	if card.Stability < 1 || card.Difficulty <= 0 {
		t.Fatalf("expected derived stability/difficulty, got %+v", card)
	}
	var st state
	if err := json.Unmarshal(card.FSRSState, &st); err != nil {
		t.Fatalf("state JSON: %v", err)
	}
	if st.Version != stateVersion || st.LastRating != "Good" || st.IntervalDays != 1 {
		t.Fatalf("state = %+v", st)
	}
}

func TestRateAgainSchedulesSoonAndRecordsLapse(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	card := Rate(domain.Card{Slug: "x", Reps: 3, Lapses: 1, Stability: 4, Difficulty: 5}, domain.RatingAgain, now)

	if card.Reps != 0 || card.Lapses != 2 {
		t.Fatalf("Again should reset reps and increment lapses: %+v", card)
	}
	if got, want := card.DueAt.Sub(now), 10*time.Minute; got != want {
		t.Fatalf("Again interval = %v, want %v", got, want)
	}
	if card.Stability != 0.1 {
		t.Fatalf("Again stability = %v, want 0.1", card.Stability)
	}
}

func TestRateEasyGrowsIntervalAndClampsDifficulty(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	card := Rate(domain.Card{Slug: "x", LastReviewAt: now.Add(-48 * time.Hour), Stability: 2, Difficulty: 1.2}, domain.RatingEasy, now)

	if got, want := card.DueAt.Sub(now), 8*24*time.Hour; got != want {
		t.Fatalf("Easy interval = %v, want %v", got, want)
	}
	if card.Difficulty < 1 || card.Difficulty > 10 {
		t.Fatalf("difficulty should be clamped, got %v", card.Difficulty)
	}
}
