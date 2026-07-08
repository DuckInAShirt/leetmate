package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/leetgo"
	"github.com/DuckInAShirt/leetmate/internal/llm"
	"github.com/DuckInAShirt/leetmate/internal/store"
	"github.com/DuckInAShirt/leetmate/internal/studyplan"
)

func cfg(lang string) *config.Config { return &config.Config{Language: lang} }

func keypress(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func enter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }
func esc() tea.KeyMsg   { return tea.KeyMsg{Type: tea.KeyEsc} }

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

func TestLogoRenderPadsLinesToSameWidth(t *testing.T) {
	rendered := renderLogo(118)
	if rendered == "" {
		t.Fatal("logo should render in the home card width")
	}
	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("logo should not be empty")
	}
	wantWidth := lipgloss.Width(lines[0])
	for i, line := range lines {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Fatalf("rendered logo line %d width = %d, want %d: %q", i+1, got, wantWidth, line)
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

func TestMenuNavReviewEmpty(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()
	m := New(Deps{Config: cfg("zh"), Store: st})
	m = apply(t, m, keypress("j")) // 1 plans
	m = apply(t, m, keypress("j")) // 2 review
	res, cmd := m.Update(enter())
	m = res.(Model)
	if !m.busy || cmd == nil {
		t.Fatalf("enter on review should start loading, busy=%v cmd=%v", m.busy, cmd)
	}
	m = runCmds(t, m, cmd)
	if !strings.Contains(m.notice, "暂无到期复习") {
		t.Errorf("empty review queue should set notice, got %q", m.notice)
	}
}

func TestBackFromPlanItemsRestoresPlanListCursor(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()
	plans := []*studyplan.Plan{
		{ID: "hot100", Title: "Hot 100", Items: []string{"1", "2"}},
		{ID: "interview150", Title: "Interview 150", Items: []string{"3", "4"}},
	}
	deps := Deps{Config: cfg("zh"), Store: st, Plans: studyplan.NewService(st, plans)}
	m := New(deps)
	m.planList = plans
	m.planCursor = 1
	m.view = viewPlanList

	m = apply(t, m, enter())
	if m.view != viewPlanItems || m.curPlanID != "interview150" {
		t.Fatalf("expected to enter interview150 items, view=%d curPlanID=%q", m.view, m.curPlanID)
	}
	if m.planCursor != 0 {
		t.Fatalf("item cursor = %d, want 0", m.planCursor)
	}

	m = apply(t, m, keypress("b"))
	if m.view != viewPlanList {
		t.Fatalf("b should return to plan list, view=%d", m.view)
	}
	if m.planCursor != 1 {
		t.Fatalf("plan list cursor = %d, want 1", m.planCursor)
	}
}

func TestPlanListEnterClampsStaleCursor(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()
	plans := []*studyplan.Plan{
		{ID: "hot100", Title: "Hot 100", Items: []string{"1", "2"}},
		{ID: "interview150", Title: "Interview 150", Items: []string{"3", "4"}},
	}
	deps := Deps{Config: cfg("zh"), Store: st, Plans: studyplan.NewService(st, plans)}
	m := New(deps)
	m.planList = plans
	m.planCursor = len(plans) // stale item cursor used to panic on enter
	m.view = viewPlanList

	m = apply(t, m, enter())
	if m.view != viewPlanItems {
		t.Fatalf("enter with stale cursor should still open a plan, view=%d", m.view)
	}
	if m.curPlanID != "interview150" {
		t.Fatalf("curPlanID = %q, want interview150", m.curPlanID)
	}
	if m.planCursor != 0 {
		t.Fatalf("item cursor = %d, want 0", m.planCursor)
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

func TestPracticeEditorAutoCompletesPairs(t *testing.T) {
	m := New(Deps{Config: cfg("zh")})
	pm := newPracticeModel(Deps{Config: cfg("zh")}, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}})
	m.practice = &pm
	m.view = viewPractice

	res, _ := m.Update(keypress("i"))
	m = res.(Model)
	if !m.practice.editing {
		t.Fatal("e should enter in-app editing mode")
	}

	m = apply(t, m, keypress("("))
	if got := m.practice.editorValue(); got != "()" {
		t.Fatalf("editor value after (: %q", got)
	}
	m = apply(t, m, keypress("x"))
	if got := m.practice.editorValue(); got != "(x)" {
		t.Fatalf("editor value after x: %q", got)
	}
	m = apply(t, m, keypress(")"))
	if got := m.practice.editorValue(); got != "(x)" {
		t.Fatalf("closing pair should be skipped, got %q", got)
	}
}

func TestPracticeEditorTabInsertsSpaces(t *testing.T) {
	m := New(Deps{Config: cfg("zh")})
	pm := newPracticeModel(Deps{Config: cfg("zh")}, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}})
	m.practice = &pm
	m.view = viewPractice

	res, _ := m.Update(keypress("i"))
	m = res.(Model)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.practice.editorValue(); got != "    " {
		t.Fatalf("tab should insert four spaces, got %q", got)
	}
}

func TestPracticeEditorSaveOnEsc(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "solution.go")
	if err := os.WriteFile(path, []byte("func f() {}"), 0o644); err != nil {
		t.Fatalf("write code: %v", err)
	}
	deps := Deps{Config: cfg("zh")}
	m := New(deps)
	pm := newPracticeModel(deps, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}, CodePath: path})
	m.practice = &pm
	m.view = viewPractice

	res, _ := m.Update(keypress("i"))
	m = res.(Model)
	m = apply(t, m, keypress("("))
	res, cmd := m.Update(esc())
	m = res.(Model)
	m = runCmds(t, m, cmd)

	if m.practice.editing {
		t.Fatal("esc should save and exit editing mode")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved code: %v", err)
	}
	if got := string(b); got != "func f() {}()" {
		t.Fatalf("saved code = %q", got)
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

func TestExpandCanSwitchStatementErrorAndCoach(t *testing.T) {
	deps := Deps{Config: cfg("zh")}
	m := New(deps)
	prob := domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}, Content: "题面第一行\n题面第二行"}
	pm := newPracticeModel(deps, prob)
	pm.fullErr = "compile error"
	pm.coachText = "review text"
	m.practice = &pm
	m.view = viewPractice

	res, _ := m.Update(keypress("o"))
	m = res.(Model)
	if !m.practice.expanded || m.practice.expandKind != "statement" {
		t.Fatalf("o should open statement detail, expanded=%v kind=%q", m.practice.expanded, m.practice.expandKind)
	}
	if got := m.practice.expandVP.View(); !strings.Contains(got, "题面第二行") {
		t.Fatalf("statement detail missing content: %q", got)
	}

	res, _ = m.Update(keypress("tab"))
	m = res.(Model)
	if m.practice.expandKind != "error" {
		t.Fatalf("tab should switch to error detail, got %q", m.practice.expandKind)
	}
	if got := m.practice.expandVP.View(); !strings.Contains(got, "compile error") {
		t.Fatalf("error detail missing content: %q", got)
	}

	res, _ = m.Update(keypress("tab"))
	m = res.(Model)
	if m.practice.expandKind != "coach" {
		t.Fatalf("tab should switch to coach detail, got %q", m.practice.expandKind)
	}
	if got := m.practice.expandVP.View(); !strings.Contains(got, "review text") {
		t.Fatalf("coach detail missing content: %q", got)
	}
}

func TestExpandedCoachDetailStreamsChunks(t *testing.T) {
	deps := Deps{Config: cfg("zh")}
	m := New(deps)
	pm := newPracticeModel(deps, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}})
	pm.coaching = true
	m.practice = &pm
	m.view = viewPractice

	m.practice.openExpandKind("coach")
	if !m.practice.expanded || m.practice.expandKind != "coach" {
		t.Fatalf("expected expanded coach detail, expanded=%v kind=%q", m.practice.expanded, m.practice.expandKind)
	}

	m = apply(t, m, coachChunkMsg{text: "first "})
	m = apply(t, m, coachChunkMsg{text: "second"})
	if m.practice.coachText != "first second" {
		t.Fatalf("coachText = %q", m.practice.coachText)
	}
	if got := m.practice.expandVP.View(); !strings.Contains(got, "first second") {
		t.Fatalf("expanded coach detail did not stream chunks: %q", got)
	}
}

func TestExpandedCoachDetailWrapsLongLines(t *testing.T) {
	pm := newPracticeModel(Deps{Config: cfg("zh")}, domain.Problem{ProblemMeta: domain.ProblemMeta{Slug: "s"}})
	pm.expandVP.Width = 20
	pm.expandVP.Height = 10
	pm.coachText = "abcdefghijklmnopqrstuvwxyz"

	pm.openExpandKind("coach")

	if !pm.expanded || pm.expandKind != "coach" {
		t.Fatalf("expected coach detail, expanded=%v kind=%q", pm.expanded, pm.expandKind)
	}
	if got := pm.expandVP.View(); !strings.Contains(got, "abcdefghijklmnopqrst\nuvwxyz") {
		t.Fatalf("expanded coach detail should wrap long lines, got %q", got)
	}
}

func TestApplyTestErrorPrefersRawLeetgoOutput(t *testing.T) {
	pm := newPracticeModel(Deps{Config: cfg("zh")}, domain.Problem{})
	raw := `✘ Wrong Answer

Passed cases:  ✘✔
Input:         [0,1,0,2,1,0,1,3,2,1,2,1]
Output:        -8
Expected:      6`
	pm.applyTest(testResultMsg{
		result: domain.TestResult{Raw: raw},
		err:    assertErr("leetgo test 42: exit status 1"),
	})
	if pm.fullErr != raw {
		t.Fatalf("fullErr = %q, want raw output", pm.fullErr)
	}
	if !strings.Contains(pm.status, "Wrong Answer") {
		t.Fatalf("status = %q, want Wrong Answer summary", pm.status)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
