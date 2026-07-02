package store

import (
	"context"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

// GetCard fetches a review card by slug.
func (s *Store) GetCard(ctx context.Context, slug string) (domain.Card, bool, error) {
	var (
		c        domain.Card
		due      string
		last     string
		reps     int
		lapses   int
		stab     float64
		diff     float64
	)
	err := s.db.QueryRowContext(ctx, `
SELECT slug, fsrs_state, due_at, reps, lapses, last_review_at, stability, difficulty
FROM cards WHERE slug=?`, slug).
		Scan(&c.Slug, &c.FSRSState, &due, &reps, &lapses, &last, &stab, &diff)
	if err != nil {
		return domain.Card{}, false, ignoreNoRows(err)
	}
	c.DueAt = ptime(due)
	c.LastReviewAt = ptime(last)
	c.Reps = reps
	c.Lapses = lapses
	c.Stability = stab
	c.Difficulty = diff
	return c, true, nil
}

// UpsertCard saves (creating or replacing) a review card.
func (s *Store) UpsertCard(ctx context.Context, c domain.Card) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO cards(slug, fsrs_state, due_at, reps, lapses, last_review_at, stability, difficulty)
VALUES(?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  fsrs_state=excluded.fsrs_state,
  due_at=excluded.due_at,
  reps=excluded.reps,
  lapses=excluded.lapses,
  last_review_at=excluded.last_review_at,
  stability=excluded.stability,
  difficulty=excluded.difficulty`,
		c.Slug, c.FSRSState, tstr(c.DueAt), c.Reps, c.Lapses, tstr(c.LastReviewAt), c.Stability, c.Difficulty)
	return err
}

// DueCards returns up to limit cards due on or before now, ordered by due time.
func (s *Store) DueCards(ctx context.Context, now time.Time, limit int) ([]domain.Card, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT slug, fsrs_state, due_at, reps, lapses, last_review_at, stability, difficulty
FROM cards WHERE due_at != '' AND due_at <= ?
ORDER BY due_at ASC LIMIT ?`, now.UTC().Format(time.RFC3339), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Card
	for rows.Next() {
		var (
			c      domain.Card
			due    string
			last   string
			stab   float64
			diff   float64
		)
		if err := rows.Scan(&c.Slug, &c.FSRSState, &due, &c.Reps, &c.Lapses, &last, &stab, &diff); err != nil {
			return nil, err
		}
		c.DueAt = ptime(due)
		c.LastReviewAt = ptime(last)
		c.Stability = stab
		c.Difficulty = diff
		out = append(out, c)
	}
	return out, rows.Err()
}
