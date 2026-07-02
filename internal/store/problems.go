package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

// UpsertProblemMeta inserts or replaces cached problem metadata.
func (s *Store) UpsertProblemMeta(ctx context.Context, m domain.ProblemMeta) error {
	tags := m.Tags
	if tags == nil {
		tags = m.TopicTags
	}
	b, _ := json.Marshal(tags)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO problems(slug, frontend_id, title, difficulty, tags, is_paid_only, updated_at)
VALUES(?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET
  frontend_id=excluded.frontend_id,
  title=excluded.title,
  difficulty=excluded.difficulty,
  tags=excluded.tags,
  is_paid_only=excluded.is_paid_only,
  updated_at=excluded.updated_at`,
		m.Slug, m.FrontendID, m.Title, string(m.Difficulty), string(b), m.IsPaidOnly, time.Now().UTC().Format(time.RFC3339))
	return err
}

// GetProblemMeta fetches cached metadata by slug.
func (s *Store) GetProblemMeta(ctx context.Context, slug string) (domain.ProblemMeta, bool, error) {
	var (
		m          domain.ProblemMeta
		difficulty string
		tagsJSON   string
		paid       int
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT slug, frontend_id, title, difficulty, tags, is_paid_only FROM problems WHERE slug=?`, slug).
		Scan(&m.Slug, &m.FrontendID, &m.Title, &difficulty, &tagsJSON, &paid)
	if err != nil {
		return domain.ProblemMeta{}, false, ignoreNoRows(err)
	}
	m.Difficulty = domain.Difficulty(difficulty)
	m.IsPaidOnly = paid != 0
	_ = json.Unmarshal([]byte(tagsJSON), &m.Tags)
	return m, true, nil
}
