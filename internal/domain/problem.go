// Package domain holds the core entities of LeetMate. It is deliberately free
// of infrastructure concerns (no SQL, no LLM SDKs) so that every other layer
// (leetgo adapter, store, coach, review, TUI) can depend on these types
// without pulling in each other's dependencies.
package domain

// Difficulty mirrors leetgo's difficulty classification.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
)

// ProblemMeta is the lightweight metadata for a LeetCode problem, sourced from
// `leetgo info --format json`. It does not include the problem statement.
type ProblemMeta struct {
	FrontendID string     `json:"frontend_id"`
	Slug       string     `json:"slug"`
	Title      string     `json:"title"`
	Difficulty Difficulty `json:"difficulty"`
	Tags       []string   `json:"tags"`
	IsPaidOnly bool       `json:"is_paid_only"`
	// TopicTags is an alias kept for compatibility with leetgo's JSON shape.
	TopicTags []string `json:"topic_tags,omitempty"`
}

// Problem is a full problem: metadata plus the statement and the path to the
// generated code skeleton on disk.
type Problem struct {
	ProblemMeta
	Content  string `json:"content"` // problem statement (markdown/html)
	CodePath string `json:"code_path"`
}

// DisplayName returns "1. Two Sum" style label for the TUI.
func (p Problem) DisplayName() string {
	return p.FrontendID + ". " + p.Title
}
