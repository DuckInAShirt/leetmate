// Package coach assembles coaching prompts and streams the LLM's reply. It is
// the product's soul: tiered help (Hint → Nudge → Review → Answer) gated by a
// system prompt that refuses to leak the full solution except at Answer tier.
package coach

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"leetmate/internal/domain"
	"leetmate/internal/llm"
)

//go:embed prompts/system.md
var systemPrompt string

//go:embed prompts/hint.md
var hintPrompt string

//go:embed prompts/nudge.md
var nudgePrompt string

//go:embed prompts/review.md
var reviewPrompt string

//go:embed prompts/answer.md
var answerPrompt string

// Coach wraps an llm.Provider with LeetMate's coaching policy.
type Coach struct {
	llm llm.Provider
}

// New builds a Coach over the given provider.
func New(p llm.Provider) *Coach { return &Coach{llm: p} }

// Request describes one coaching turn.
type Request struct {
	Tier    domain.Tier
	Problem domain.Problem
	Code    string
	Lang    string
	History []domain.Conversation
}

// Stream sends the assembled prompt and returns a channel of reply chunks.
func (c *Coach) Stream(ctx context.Context, req Request) (<-chan llm.Chunk, error) {
	return c.llm.Chat(ctx, c.buildMessages(req), llm.Options{Temperature: 0.3})
}

// buildMessages assembles the system + context + history into chat messages.
func (c *Coach) buildMessages(req Request) []llm.Message {
	sys := systemPrompt + "\n\n" + levelPrompt(req.Tier)

	var b strings.Builder
	b.WriteString("题目：" + req.Problem.DisplayName() + "\n")
	b.WriteString("难度：" + string(req.Problem.Difficulty) + "\n\n")
	b.WriteString("【题面】\n")
	b.WriteString(strings.TrimSpace(req.Problem.Content))
	b.WriteString("\n")
	if strings.TrimSpace(req.Code) != "" {
		lang := req.Lang
		if lang == "" {
			lang = "go"
		}
		b.WriteString("\n【我目前的代码】\n```" + lang + "\n" + req.Code + "\n```\n")
	}
	switch req.Tier {
	case domain.TierHint:
		b.WriteString("\n请给我一个 Hint。")
	case domain.TierNudge:
		b.WriteString("\n我卡住了，请给我一个 Nudge。")
	case domain.TierReview:
		b.WriteString("\n请审查我上面的代码。")
	case domain.TierAnswer:
		b.WriteString("\n我已确认要看答案，请给出完整解答。")
	}

	msgs := []llm.Message{{Role: llm.RoleSystem, Content: sys}}
	for _, h := range req.History {
		// Only carry user/assistant turns into context.
		if h.Role == domain.RoleUser || h.Role == domain.RoleAssistant {
			msgs = append(msgs, llm.Message{Role: llm.Role(h.Role), Content: h.Content})
		}
	}
	msgs = append(msgs, llm.Message{Role: llm.RoleUser, Content: b.String()})
	return msgs
}

func levelPrompt(t domain.Tier) string {
	switch t {
	case domain.TierHint:
		return hintPrompt
	case domain.TierNudge:
		return nudgePrompt
	case domain.TierReview:
		return reviewPrompt
	case domain.TierAnswer:
		return answerPrompt
	default:
		return hintPrompt
	}
}

// CodeFence extracts the contents of the first fenced code block in text, if
// any. Used so an Answer-tier reply can be offered to the user verbatim.
func CodeFence(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	in := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "```") {
			if in {
				return strings.Join(out, "\n")
			}
			in = true
			continue
		}
		if in {
			out = append(out, l)
		}
	}
	if len(out) > 0 {
		return strings.Join(out, "\n")
	}
	return ""
}

// Summary returns a short label for a tier, for UI display.
func Summary(t domain.Tier) string {
	switch t {
	case domain.TierHint:
		return "Hint"
	case domain.TierNudge:
		return "Nudge"
	case domain.TierReview:
		return "Review"
	case domain.TierAnswer:
		return "Answer"
	}
	return fmt.Sprintf("%s", t)
}
