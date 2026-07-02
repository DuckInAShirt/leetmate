package store

import (
	"context"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

// InsertAttempt records one practice session and returns its id.
func (s *Store) InsertAttempt(ctx context.Context, a domain.Attempt) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO attempts(slug, started_at, finished_at, ac, runtime_ms, memory_kb, rating, gave_up)
VALUES(?,?,?,?,?,?,?,?)`,
		a.Slug, tstr(a.StartedAt), tstr(a.FinishedAt), btoi(a.AC), a.RuntimeMS, a.MemoryKB, int(a.Rating), btoi(a.GaveUp))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListAttempts returns attempts for a slug, newest first.
func (s *Store) ListAttempts(ctx context.Context, slug string) ([]domain.Attempt, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, slug, started_at, finished_at, ac, runtime_ms, memory_kb, rating, gave_up
		 FROM attempts WHERE slug=? ORDER BY id DESC`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Attempt
	for rows.Next() {
		var (
			a         domain.Attempt
			started   string
			finished  string
			ac, gave  int
			ratingInt int
		)
		if err := rows.Scan(&a.ID, &a.Slug, &started, &finished, &ac, &a.RuntimeMS, &a.MemoryKB, &ratingInt, &gave); err != nil {
			return nil, err
		}
		a.StartedAt = ptime(started)
		a.FinishedAt = ptime(finished)
		a.AC = ac != 0
		a.GaveUp = gave != 0
		a.Rating = domain.Rating(ratingInt)
		out = append(out, a)
	}
	return out, rows.Err()
}

// --- helpers ---

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func tstr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func ptime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func ignoreNoRows(err error) error {
	if err.Error() == "sql: no rows" {
		return nil
	}
	return err
}
