package domain

import "time"

// Card is LeetMate's view of an FSRS review card. The opaque scheduler state is
// stored serialized (by the review package) in FSRSState; LeetMate only needs
// the derived scheduling fields for the UI.
type Card struct {
	Slug         string
	FSRSState    []byte    // serialized go-fsrs Card state (JSON)
	DueAt        time.Time // next scheduled review
	Reps         int       // successful reviews in a row
	Lapses       int       // times forgotten
	LastReviewAt time.Time
	Stability    float64   // FSRS stability (days), echoed for convenience
	Difficulty   float64   // FSRS difficulty [0..10], echoed for convenience
}

// IsDue reports whether the card is due on or before the given time.
func (c Card) IsDue(now time.Time) bool {
	return !c.DueAt.After(now)
}
