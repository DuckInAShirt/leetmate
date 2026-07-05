// Package review owns LeetMate's spaced-review scheduling logic.
package review

import (
	"encoding/json"
	"math"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

const stateVersion = 1

type state struct {
	Version      int    `json:"version"`
	LastRating   string `json:"last_rating"`
	IntervalDays int    `json:"interval_days"`
}

// RatingForSubmit maps an observed submission outcome to LeetMate's review
// scale. Revealing the answer means the learner did not independently recall
// the solution, even if the final submission was accepted.
func RatingForSubmit(accepted bool, gaveUp bool) domain.Rating {
	if accepted && !gaveUp {
		return domain.RatingGood
	}
	return domain.RatingAgain
}

// Rate applies a lightweight FSRS-style scheduling step to a card. It keeps the
// algorithm behind this package so a future full FSRS implementation can replace
// it without changing the TUI or store layers.
func Rate(card domain.Card, rating domain.Rating, now time.Time) domain.Card {
	if rating == domain.RatingManual {
		rating = domain.RatingGood
	}
	if card.Difficulty == 0 {
		card.Difficulty = 5
	}
	if card.Stability == 0 {
		card.Stability = initialStability(rating)
	}

	intervalDays := interval(card, rating)
	card.LastReviewAt = now.UTC()
	card.DueAt = now.UTC().Add(intervalDuration(intervalDays, rating))
	card.Stability = nextStability(card, rating, intervalDays)
	card.Difficulty = nextDifficulty(card.Difficulty, rating)

	if rating == domain.RatingAgain {
		card.Lapses++
		card.Reps = 0
	} else {
		card.Reps++
	}

	b, _ := json.Marshal(state{Version: stateVersion, LastRating: rating.String(), IntervalDays: intervalDays})
	card.FSRSState = b
	return card
}

func interval(card domain.Card, rating domain.Rating) int {
	if isNewCard(card) {
		switch rating {
		case domain.RatingAgain:
			return 0
		case domain.RatingEasy:
			return 3
		default:
			return 1
		}
	}
	s := card.Stability
	if s <= 0 {
		s = initialStability(rating)
	}
	switch rating {
	case domain.RatingAgain:
		return 0
	case domain.RatingHard:
		return maxInt(1, int(math.Ceil(s*1.2)))
	case domain.RatingEasy:
		return maxInt(3, int(math.Ceil(s*4)))
	default: // Good and Manual
		return maxInt(1, int(math.Ceil(s*2.5)))
	}
}

func isNewCard(card domain.Card) bool {
	return card.Reps == 0 && card.Lapses == 0 && card.LastReviewAt.IsZero() && len(card.FSRSState) == 0
}

func intervalDuration(days int, rating domain.Rating) time.Duration {
	if rating == domain.RatingAgain || days <= 0 {
		return 10 * time.Minute
	}
	return time.Duration(days) * 24 * time.Hour
}

func initialStability(rating domain.Rating) float64 {
	switch rating {
	case domain.RatingAgain:
		return 0.1
	case domain.RatingHard:
		return 1
	case domain.RatingEasy:
		return 3
	default:
		return 1
	}
}

func nextStability(card domain.Card, rating domain.Rating, intervalDays int) float64 {
	if rating == domain.RatingAgain {
		return 0.1
	}
	return math.Max(float64(intervalDays), initialStability(rating))
}

func nextDifficulty(d float64, rating domain.Rating) float64 {
	switch rating {
	case domain.RatingAgain:
		d += 1
	case domain.RatingHard:
		d += 0.3
	case domain.RatingEasy:
		d -= 0.8
	case domain.RatingGood:
		d -= 0.2
	}
	return clamp(d, 1, 10)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
