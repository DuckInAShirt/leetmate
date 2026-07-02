package domain

import "time"

// Role identifies who produced a coaching message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Tier is the coaching tier a conversation is happening under. See coach package.
type Tier string

const (
	TierHint   Tier = "hint"
	TierNudge  Tier = "nudge"
	TierReview Tier = "review"
	TierAnswer Tier = "answer"
)

// Conversation is a coaching exchange tied to a problem. History is persisted
// so multi-turn coaching can carry context.
type Conversation struct {
	ID        int64
	Slug      string
	Tier      Tier
	Role      Role
	Content   string
	CreatedAt time.Time
}
