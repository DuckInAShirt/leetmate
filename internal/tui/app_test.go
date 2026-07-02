package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"leetmate/internal/coach"
	"leetmate/internal/config"
	"leetmate/internal/domain"
	"leetmate/internal/leetgo"
	"leetmate/internal/llm"
	"leetmate/internal/store"
)

func cfg(lang string) *config.Config { return &config.Config{Language: lang} }

func keypress(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func enter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }

// apply sends a message to m and asserts the result stays a Model.
func apply(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	res, _ := m.Update(msg)
	mm, ok := res.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", res)
	}
	return mm
}

func TestMenuRenderZH(t *testing.T) {
	m := New(Deps{Config: cfg("zh")})
	v := m.View()
	for _, want := range []string{"今日题目", "待复习", "退出", "选择"} {
		if !strings.Contains(v, want) {
			t.Errorf("zh view missing %q\n%s", want, v)
		}
	}
}

func TestMenuRenderEN(t *testing.T) {
	m := New(Deps{Config: cfg("en")})
	v := m.View()
	for _, want := range []string{"Today's problem", "Due for review", "Quit", "select"} {
		if !strings.Contains(v, want) {
			t.Errorf("en view missing %q\n%s", want, v)
		}
	}
}

func TestMenuNavPlansSafeWithoutService(t *testing.T) {
	// cursor 1 is now "Study plans"; with no Plans service it must not crash.
	m := New(Deps{Config: cfg("zh")})
	m = apply(t, m, keypress("j"))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	m = apply(t, m, enter())
	if m.view != viewMenu {
		t.Errorf("nil Plans should not change view, view=%d", m.view)
	}
}

func TestMenuNavReviewNotice(t *testing.T) {
	m := New(Deps{Config: cfg("zh")})
	m = apply(t, m, keypress("j")) // 1 plans
	m = apply(t, m, keypress("j")) // 2 review
	m = apply(t, m, enter())
	if !strings.Contains(m.notice, "M3") {
		t.Errorf("enter on review should set notice, got %q", m.notice)
	}
}

func TestPickTransitionsToPracticeAndRendersTitle(t *testing.T) {
	m := New(Deps{Config: cfg("zh")})
	prob := domain.Problem{ProblemMeta: domain.ProblemMeta{
		FrontendID: "3286", Slug: "find-a-safe-walk-through-a-grid",
		Title: "穿越网格图的安全路径", Difficulty: domain.DifficultyMedium,
	}}
	m = apply(t, m, pickResultMsg{problem: prob})
	if m.view != viewPractice || m.practice == nil {
		t.Fatalf("expected practice view, view=%d practice=%v", m.view, m.practice)
	}
	v := m.View()
	if !strings.Contains(v, "3286. 穿越网格图的安全路径") {
		t.Errorf("practice view missing title\n%s", v)
	}
	if !strings.Contains(v, "中等") {
		t.Errorf("practice view missing zh difficulty\n%s", v)
	}
}

func TestPracticeDifficultyEN(t *testing.T) {
	m := New(Deps{Config: cfg("en")})
	prob := domain.Problem{ProblemMeta: domain.ProblemMeta{
		FrontendID: "1", Slug: "two-sum", Title: "Two Sum", Difficulty: domain.DifficultyEasy,
	}}
	m = apply(t, m, pickResultMsg{problem: prob})
	if !strings.Contains(m.View(), "Easy") {
		t.Errorf("en practice view missing Easy\n%s", m.View())
	}
}

// fakeLLM implements llm.Provider for testing the coaching pipeline.
type fakeLLM struct{ chunks []string }

func (f *fakeLLM) Chat(_ context.Context, _ []llm.Message, _ llm.Options) (<-chan llm.Chunk, error) {
	ch := make(chan llm.Chunk, len(f.chunks))
	for _, c := range f.chunks {
		ch <- llm.Chunk{Text: c}
	}
	close(ch)
	return ch, nil
}

// runCmds drains a tea.Cmd chain (each cmd produces one msg fed back through
// Update) until no cmd remains. This simulates the bubbletea loop in tests.
func runCmds(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		res, next := m.Update(msg)
		mm, ok := res.(Model)
		if !ok {
			t.Fatalf("expected Model, got %T", res)
		}
		m = mm
		cmd = next
	}
	return m
}

// TestCoachingFlowStreamsIntoPanel drives the full pipeline with a fake LLM:
// pressing "1" (Hint) → stream opens → chunks accumulate into the coach panel.
func TestCoachingFlowStreamsIntoPanel(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()
	lc, err := leetgo.New(config.LeetgoConfig{Workspace: dir})
	if err != nil {
		t.Fatalf("leetgo.New: %v", err)
	}
	deps := Deps{
		Config: cfg("zh"), Store: st, Leetgo: lc,
		Coach: coach.New(&fakeLLM{chunks: []string{"想想", "单调栈"}}),
	}

	m := New(deps)
	prob := domain.Problem{ProblemMeta: domain.ProblemMeta{
		FrontendID: "1", Slug: "two-sum", Title: "Two Sum", Difficulty: domain.DifficultyEasy,
	}}
	pm := newPracticeModel(deps, prob)
	m.practice = &pm
	m.view = viewPractice

	// Press "1" (Hint). Update returns the coachStartCmd that begins the chain.
	res, cmd := m.Update(keypress("1"))
	m = res.(Model)
	m = runCmds(t, m, cmd)

	if m.practice.coachText != "想想单调栈" {
		t.Errorf("coachText = %q, want %q", m.practice.coachText, "想想单调栈")
	}
	if m.practice.coaching {
		t.Error("coaching flag should be cleared after stream completes")
	}
	// The reply must have been persisted as an assistant conversation turn.
	hist, err := st.RecentConversations(context.Background(), "two-sum", 10)
	if err != nil || len(hist) != 1 || hist[0].Content != "想想单调栈" {
		t.Errorf("conversation not persisted: %v (err %v)", hist, err)
	}
}

// TestAnswerConfirmGate verifies pressing "4" requires a y/n confirmation.
func TestAnswerConfirmGate(t *testing.T) {
	dir := t.TempDir()
	st, _ := store.Open(filepath.Join(dir, "t.db"))
	defer st.Close()
	lc, _ := leetgo.New(config.LeetgoConfig{Workspace: dir})
	deps := Deps{Config: cfg("zh"), Store: st, Leetgo: lc,
		Coach: coach.New(&fakeLLM{chunks: []string{"ans"}})}

	m := New(deps)
	pm := newPracticeModel(deps, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}})
	m.practice = &pm
	m.view = viewPractice

	// "4" arms confirmation but does NOT stream yet.
	res, _ := m.Update(keypress("4"))
	m = res.(Model)
	if !m.practice.answerConfirm || m.coachStream != nil {
		t.Error(`"4" should arm confirm without starting stream`)
	}
	if !strings.Contains(m.View(), "确认") {
		t.Errorf("confirm prompt not shown:\n%s", m.View())
	}

	// "n" cancels.
	res, _ = m.Update(keypress("n"))
	m = res.(Model)
	if m.practice.answerConfirm {
		t.Error(`"n" should cancel confirmation`)
	}
}
