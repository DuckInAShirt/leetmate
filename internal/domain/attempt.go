package domain

import "time"

// Rating is the learner's subjective recall quality after a review, mirroring
// the FSRS scale. It is defined here (rather than importing go-fsrs) so the
// domain layer stays infrastructure-free; the review package maps this to
// fsrs.Rating.
type Rating int

const (
	// RatingManual is the zero value and means "not yet rated".
	RatingManual Rating = iota
	RatingAgain
	RatingHard
	RatingGood
	RatingEasy
)

func (r Rating) String() string {
	switch r {
	case RatingAgain:
		return "Again"
	case RatingHard:
		return "Hard"
	case RatingGood:
		return "Good"
	case RatingEasy:
		return "Easy"
	default:
		return "Manual"
	}
}

// Attempt records a single practice session on a problem, from opening it to
// submitting (or giving up).
type Attempt struct {
	ID         int64
	Slug       string
	StartedAt  time.Time
	FinishedAt time.Time
	AC         bool       // accepted by LeetCode
	RuntimeMS  int        // runtime in milliseconds (0 if unknown/failed)
	MemoryKB   int        // memory in KB (0 if unknown/failed)
	Rating     Rating     // learner's self-assessment (Manual until rated)
	GaveUp     bool       // true if the learner revealed the answer (Answer tier)
}
