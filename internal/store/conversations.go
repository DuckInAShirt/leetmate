package store

import (
	"context"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

// InsertConversation appends a coaching message.
func (s *Store) InsertConversation(ctx context.Context, conv domain.Conversation) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO conversations(slug, tier, role, content, created_at)
VALUES(?,?,?,?,?)`,
		conv.Slug, string(conv.Tier), string(conv.Role), conv.Content, tstr(conv.CreatedAt))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// RecentConversations returns the last n coaching messages for a slug, oldest
// first within that window (ready to feed into an LLM prompt as history).
func (s *Store) RecentConversations(ctx context.Context, slug string, n int) ([]domain.Conversation, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, slug, tier, role, content, created_at FROM (
  SELECT * FROM conversations WHERE slug=? ORDER BY id DESC LIMIT ?
) ORDER BY id ASC`, slug, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Conversation
	for rows.Next() {
		var (
			c    domain.Conversation
			tier string
			role string
			ts   string
		)
		if err := rows.Scan(&c.ID, &c.Slug, &tier, &role, &c.Content, &ts); err != nil {
			return nil, err
		}
		c.Tier = domain.Tier(tier)
		c.Role = domain.Role(role)
		c.CreatedAt = ptime(ts)
		out = append(out, c)
	}
	return out, rows.Err()
}
