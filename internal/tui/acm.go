package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/leetgo"
)

// acmModel is the ACM-mode view. The compact view shows the statement summary
// and the code editor; pressing o opens a full-screen expand mode where tab
// cycles statement / code / stdin / output — mirroring practice's expand mode
// so the keybindings stay consistent across modes. The learner writes plain
// input/output from scratch (no leetgo skeleton, no LeetCode serialization); r
// runs the code with the stdin pane's contents.
type acmModel struct {
	problem       domain.Problem
	statement     string // compact summary
	fullStatement string // full statement, for expand mode
	codeEditor    textarea.Model
	stdinEditor   textarea.Model
	expandVP      viewport.Model // full-screen view of statement / output
	acmPath       string         // <problem dir>/acm.<ext>, persisted separately from solution
	lang          string

	status string
	output string

	editing       bool
	expanded      bool
	expandKind    string // "statement" / "code" / "stdin" / "output"
	saving        bool
	exitAfterSave bool
	running       bool
	d             dict
}

// newACMModel builds the ACM view. acmPath sits next to the leetgo solution file
// (acm.py / acm.go) so it persists across sessions without touching the
// core-code skeleton.
func newACMModel(deps Deps, p domain.Problem) acmModel {
	m := acmModel{
		problem: p,
		d:       loadDict(deps.Config.Language),
		lang:    deps.Leetgo.Lang(),
	}
	m.fullStatement = cleanStatement(p.Content, 0)
	m.statement = truncateLines(m.fullStatement, 10)
	initial := ""
	if p.CodePath != "" {
		ext := filepath.Ext(p.CodePath)
		m.acmPath = filepath.Join(filepath.Dir(p.CodePath), "acm"+ext)
		if b, err := os.ReadFile(m.acmPath); err == nil {
			initial = string(b)
		}
	}
	m.codeEditor = newCodeEditor(initial)
	m.stdinEditor = newCodeEditor("")
	m.expandVP = viewport.New(80, 20)
	return m
}

func (m *acmModel) resize(w, h int) {
	if w < 8 || h < 8 {
		return
	}
	vw := w - 4
	if vw < 10 {
		vw = 10
	}
	stmtLines := countLines(wrapWidth(m.statement, vw))
	// header(1) + blank(1) + statement + section(1) + status(1) + hint(1).
	reserved := 5 + stmtLines
	avail := h - reserved
	if avail < 3 {
		avail = 3
	}
	m.codeEditor.SetWidth(vw)
	m.codeEditor.SetHeight(avail)
	// stdin/editor panes fill the screen in expand mode.
	m.stdinEditor.SetWidth(vw)
	m.stdinEditor.SetHeight(maxInt(h-4, 3))
	m.expandVP.Width = vw
	m.expandVP.Height = maxInt(h-4, 3)
}

func (m *acmModel) view() string {
	if m.expanded {
		return m.renderExpand()
	}
	var b strings.Builder
	diff := m.problem.Difficulty
	b.WriteString(titleStyle.Render(m.problem.DisplayName()) + "  " +
		difficultyStyle(string(diff)).Render(m.d.difficultyLabel(string(diff))) + "  " +
		subtleStyle.Render(m.d.t("acm.badge")) + "\n\n")

	width := m.expandVP.Width
	if width < 10 {
		width = 76
	}
	if m.statement != "" {
		b.WriteString(subtleStyle.Render(wrapWidth(m.statement, width)) + "\n")
	}
	b.WriteString(subtleStyle.Render("▸ " + m.d.t("acm.sectionCode")) + "\n")
	b.WriteString(m.codeEditor.View())

	if m.status != "" {
		style := subtleStyle
		switch {
		case strings.HasPrefix(m.status, "✓"):
			style = statusOKStyle
		case strings.HasPrefix(m.status, "✗"), strings.HasPrefix(m.status, "⚠"):
			style = statusErrStyle
		}
		sw := m.expandVP.Width
		if sw < 10 {
			sw = 76
		}
		b.WriteString("\n" + style.Render(truncateWidth(firstLine(m.status), sw)))
	}

	hint := m.d.t("acm.hint")
	if m.editing {
		hint = m.d.t("acm.editorHint")
	}
	b.WriteString(hintStyle.Render("\n" + hint))
	return b.String()
}

// --- expand mode (mirrors practiceModel's expand, for keybinding consistency) ---

var acmExpandKinds = []string{"statement", "code", "stdin", "output"}

func (m *acmModel) expandTitleKey() string {
	switch m.expandKind {
	case "statement":
		return "expand.statement"
	case "stdin":
		return "acm.sectionStdin"
	case "output":
		return "acm.sectionOutput"
	default:
		return "acm.sectionCode"
	}
}

func (m *acmModel) openExpand(kind string) {
	m.expandKind = kind
	m.expanded = true
	m.refreshExpand()
}

func (m *acmModel) cycleExpand() {
	for i, k := range acmExpandKinds {
		if k == m.expandKind {
			m.openExpand(acmExpandKinds[(i+1)%len(acmExpandKinds)])
			return
		}
	}
	m.openExpand("statement")
}

// refreshExpand loads viewport content for the statement/output panes.
func (m *acmModel) refreshExpand() {
	w := m.expandVP.Width
	if w < 10 {
		w = 76
	}
	switch m.expandKind {
	case "statement":
		m.expandVP.SetContent(wrapWidth(m.fullStatement, w))
		m.expandVP.GotoTop()
	case "output":
		content := m.output
		if content == "" {
			content = subtleStyle.Render(m.d.t("acm.outputEmpty"))
		}
		m.expandVP.SetContent(wrapWidth(content, w))
		m.expandVP.GotoTop()
	}
}

func (m *acmModel) renderExpand() string {
	title := titleStyle.Render(m.d.t(m.expandTitleKey()))
	hint := subtleStyle.Render(m.d.t("acm.expandHint"))
	var body string
	switch m.expandKind {
	case "code":
		body = m.codeEditor.View()
	case "stdin":
		body = m.stdinEditor.View()
	default: // statement / output
		body = m.expandVP.View()
	}
	return title + "  " + hint + "\n\n" + body
}

func (m *acmModel) scrollExpand(dir int, page bool) {
	if m.expandKind != "statement" && m.expandKind != "output" {
		return
	}
	switch {
	case page && dir < 0:
		m.expandVP.HalfViewUp()
	case page:
		m.expandVP.HalfViewDown()
	case dir < 0:
		m.expandVP.LineUp(1)
	default:
		m.expandVP.LineDown(1)
	}
}

// applyRun renders one run's stdout/stderr into the output pane + status.
func (m *acmModel) applyRun(msg runResultMsg) {
	m.running = false
	var b strings.Builder
	if strings.TrimSpace(msg.stderr) != "" {
		b.WriteString(statusErrStyle.Render(m.d.t("acm.stderr")) + "\n" +
			strings.TrimSpace(msg.stderr) + "\n\n")
	}
	b.WriteString(subtleStyle.Render(m.d.t("acm.stdout")) + "\n")
	out := strings.TrimSpace(msg.stdout)
	if out == "" {
		b.WriteString(subtleStyle.Render(m.d.t("acm.emptyOut")))
	} else {
		b.WriteString(out)
	}
	m.output = strings.TrimRight(b.String(), "\n")

	switch {
	case msg.err == nil:
		m.status = m.d.t("acm.done")
	case errors.Is(msg.err, leetgo.ErrACMTimeout):
		m.status = m.d.t("acm.timeout")
	default:
		m.status = "⚠ " + m.d.t("acm.runError")
	}
	// If the output pane is already open, refresh it with the new result.
	if m.expanded && m.expandKind == "output" {
		m.refreshExpand()
	}
}

// --- editor ---

func (m *acmModel) codeValue() string  { return m.codeEditor.Value() }
func (m *acmModel) stdinValue() string { return m.stdinEditor.Value() }

// focusedEditor: compact mode → code editor; expand mode → the current pane's
// editor (code or stdin).
func (m *acmModel) focusedEditor() *textarea.Model {
	if m.expanded && m.expandKind == "stdin" {
		return &m.stdinEditor
	}
	return &m.codeEditor
}

func (m *acmModel) startEditing() tea.Cmd {
	m.editing = true
	return m.focusedEditor().Focus()
}

func (m *acmModel) stopEditing() {
	m.editing = false
	m.saving = false
	m.exitAfterSave = false
	m.codeEditor.Blur()
	m.stdinEditor.Blur()
}

func (m *acmModel) saveEditor(exitAfter bool) tea.Cmd {
	if m.acmPath == "" {
		m.exitAfterSave = exitAfter
		return nil
	}
	if m.saving {
		return nil
	}
	m.saving = true
	m.exitAfterSave = exitAfter
	return saveCodeCmd(m.acmPath, m.codeValue())
}

func (m *acmModel) applyEditorSaved(msg editorSavedMsg) {
	m.saving = false
	if msg.err != nil {
		m.status = m.d.t("practice.saveError") + summarizeErr(msg.err.Error())
		return
	}
	m.status = m.d.t("practice.saved")
	if m.exitAfterSave {
		m.stopEditing()
	}
	m.exitAfterSave = false
}

func (m *acmModel) insertEditorText(s string) {
	m.focusedEditor().InsertString(s)
}

func (m *acmModel) updateEditorKey(key tea.KeyMsg) tea.Cmd {
	ed := m.focusedEditor()
	if len(key.Runes) == 1 {
		r := key.Runes[0]
		if skipClosingPair(ed, r) || insertPair(ed, r) {
			return nil
		}
	}
	var cmd tea.Cmd
	*ed, cmd = ed.Update(key)
	return cmd
}

func insertPair(ed *textarea.Model, r rune) bool {
	close, ok := pairClose(r)
	if !ok {
		return false
	}
	col := editorCursorColumn(ed)
	ed.InsertString(string([]rune{r, close}))
	ed.SetCursor(col + 1)
	return true
}

func skipClosingPair(ed *textarea.Model, r rune) bool {
	if !isPairCloser(r) {
		return false
	}
	line, col := editorLineAndColumn(ed)
	lines := strings.Split(ed.Value(), "\n")
	if line < 0 || line >= len(lines) {
		return false
	}
	runes := []rune(lines[line])
	if col >= len(runes) || runes[col] != r {
		return false
	}
	ed.SetCursor(col + 1)
	return true
}

func editorLineAndColumn(ed *textarea.Model) (int, int) {
	li := ed.LineInfo()
	return ed.Line(), li.StartColumn + li.ColumnOffset
}

func editorCursorColumn(ed *textarea.Model) int {
	_, col := editorLineAndColumn(ed)
	return col
}
