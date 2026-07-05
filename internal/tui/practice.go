package tui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"

	"github.com/DuckInAShirt/leetmate/internal/domain"
)

const (
	focusCode = iota
	focusCoach
)

// practiceModel is the per-problem view: statement summary, the learner's code
// (scrollable), and the keys to edit / test / submit. The coaching panel (M1)
// will be added alongside this view.
type practiceModel struct {
	problem       domain.Problem
	code          string
	statement     string
	fullStatement string

	d      dict
	gaveUp bool
	status string

	viewport viewport.Model // highlighted code preview
	editor   textarea.Model // in-app code editor
	coachVP  viewport.Model // coaching reply
	expandVP viewport.Model // full-screen detail (error / coach reply)
	ready    bool
	focus    int // focusCode / focusCoach: which pane scroll keys act on

	editing       bool
	dirty         bool
	saving        bool
	exitAfterSave bool

	coachText     string
	coachTier     domain.Tier
	coaching      bool
	coachErr      string
	answerConfirm bool
	planCtx       *planCtx // non-nil when launched from a study plan

	// expand-mode state
	expanded   bool
	expandKind string // "error" or "coach"
	fullErr    string // complete last error text (for expand)
}

func newPracticeModel(deps Deps, p domain.Problem) practiceModel {
	m := practiceModel{problem: p, d: loadDict(deps.Config.Language)}
	if p.CodePath != "" {
		if b, err := os.ReadFile(p.CodePath); err == nil {
			m.code = string(b)
		}
	}
	// Clean the statement (strip markdown noise / title heading). Keep a short
	// summary above the code viewport and the full text for expand mode.
	m.fullStatement = cleanStatement(p.Content, 0)
	m.statement = truncateLines(m.fullStatement, 10)
	m.viewport = viewport.New(80, 20)
	m.editor = newCodeEditor(m.code)
	m.coachVP = viewport.New(80, 8)
	m.expandVP = viewport.New(80, 20)
	m.setCodeContent(deps)
	m.ready = true
	return m
}

func (m *practiceModel) qid() string {
	if m.problem.FrontendID != "" {
		return m.problem.FrontendID
	}
	return m.problem.Slug
}

func (m *practiceModel) resize(w, h int) {
	if w < 8 || h < 8 {
		return
	}
	vw := w - 4
	if vw < 10 {
		vw = 10
	}
	// Reserve header(1) + blank(1) + statement lines + separator(1) + status(1)
	// + hint(1) + a small margin. Statement lines are counted *after* wrapping
	// to the current width so the code viewport never overflows the screen.
	stmtLines := countLines(wrapWidth(m.statement, vw))
	// header(1) + blank(1) + statement + code-sep(1) + coach-sep(1) + status(1)
	// + hint(1); remaining height splits between code and coach viewports.
	reserved := 6 + stmtLines
	avail := h - reserved
	if avail < 8 {
		avail = 8
	}
	codeH := avail * 3 / 5
	coachH := avail - codeH
	// Folded coach preview: keep it small (a few lines) so the code pane
	// dominates; expand the full reply with `o`.
	coachH = 4
	if coachH > avail-3 {
		coachH = avail / 2
	}
	codeH = avail - coachH
	if codeH < 3 {
		codeH = 3
	}
	if coachH < 3 {
		coachH = 3
	}
	m.viewport.Width = vw
	m.viewport.Height = codeH
	m.editor.SetWidth(vw)
	m.editor.SetHeight(codeH)
	m.coachVP.Width = vw
	m.coachVP.Height = coachH
	// Expand pane covers nearly the whole screen.
	m.expandVP.Width = w - 2
	if m.expandVP.Width < 10 {
		m.expandVP.Width = 80
	}
	m.expandVP.Height = h - 4
	if m.expandVP.Height < 3 {
		m.expandVP.Height = 3
	}
}

func (m *practiceModel) reloadCode(deps Deps) {
	if m.problem.CodePath == "" || m.editing {
		return
	}
	if b, err := os.ReadFile(m.problem.CodePath); err == nil {
		m.code = string(b)
		m.editor.SetValue(m.code)
		m.dirty = false
		m.setCodeContent(deps)
	}
}

func (m *practiceModel) setCodeContent(deps Deps) {
	m.viewport.SetContent(highlightCode(m.code, codeLanguage(deps, m.problem.CodePath)))
}

func (m *practiceModel) startEditing() tea.Cmd {
	m.editing = true
	m.focus = focusCode
	return m.editor.Focus()
}

func (m *practiceModel) stopEditing() {
	m.editing = false
	m.saving = false
	m.exitAfterSave = false
	m.editor.Blur()
}

func (m *practiceModel) editorValue() string { return m.editor.Value() }

func (m *practiceModel) saveEditor(exitAfter bool) tea.Cmd {
	if m.saving {
		return nil
	}
	m.saving = true
	m.exitAfterSave = exitAfter
	return saveCodeCmd(m.problem.CodePath, m.editorValue())
}

func (m *practiceModel) applyEditorSaved(deps Deps, msg editorSavedMsg) {
	m.saving = false
	if msg.err != nil {
		m.status = m.d.t("practice.saveError") + summarizeErr(msg.err.Error())
		return
	}
	m.code = msg.content
	m.dirty = false
	m.status = m.d.t("practice.saved")
	m.setCodeContent(deps)
	if m.exitAfterSave {
		m.stopEditing()
	}
	m.exitAfterSave = false
}

func newCodeEditor(code string) textarea.Model {
	ed := textarea.New()
	ed.Prompt = ""
	ed.ShowLineNumbers = true
	ed.EndOfBufferCharacter = ' '
	ed.MaxHeight = 10000
	ed.MaxWidth = 0
	ed.FocusedStyle = codeEditorFocusedStyle()
	ed.BlurredStyle = codeEditorBlurredStyle()
	ed.SetValue(code)
	return ed
}

func (m *practiceModel) insertEditorText(s string) {
	m.editor.InsertString(s)
	m.dirty = m.editorValue() != m.code
}

func (m *practiceModel) updateEditorKey(key tea.KeyMsg) tea.Cmd {
	before := m.editorValue()
	if len(key.Runes) == 1 {
		r := key.Runes[0]
		if m.skipClosingPair(r) || m.insertPair(r) {
			m.dirty = m.editorValue() != m.code
			return nil
		}
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(key)
	m.dirty = m.editorValue() != before || m.dirty
	return cmd
}

func (m *practiceModel) insertPair(r rune) bool {
	close, ok := pairClose(r)
	if !ok {
		return false
	}
	col := m.editorCursorColumn()
	m.editor.InsertString(string([]rune{r, close}))
	m.editor.SetCursor(col + 1)
	return true
}

func (m *practiceModel) skipClosingPair(r rune) bool {
	if !isPairCloser(r) {
		return false
	}
	line, col := m.editorLineAndColumn()
	if line < 0 || line >= len(strings.Split(m.editorValue(), "\n")) {
		return false
	}
	runes := []rune(strings.Split(m.editorValue(), "\n")[line])
	if col >= len(runes) || runes[col] != r {
		return false
	}
	m.editor.SetCursor(col + 1)
	return true
}

func (m *practiceModel) editorLineAndColumn() (int, int) {
	li := m.editor.LineInfo()
	return m.editor.Line(), li.StartColumn + li.ColumnOffset
}

func (m *practiceModel) editorCursorColumn() int {
	_, col := m.editorLineAndColumn()
	return col
}

func pairClose(r rune) (rune, bool) {
	switch r {
	case '(':
		return ')', true
	case '[':
		return ']', true
	case '{':
		return '}', true
	case '"':
		return '"', true
	case '\'':
		return '\'', true
	default:
		return 0, false
	}
}

func isPairCloser(r rune) bool {
	switch r {
	case ')', ']', '}', '"', '\'':
		return true
	default:
		return false
	}
}

func (m *practiceModel) applySubmit(msg submitResultMsg) {
	if msg.err != nil {
		m.fullErr = msg.err.Error()
		m.status = m.d.t("practice.submitError") + summarizeErr(m.fullErr)
		return
	}
	if msg.result.Accepted {
		m.fullErr = ""
		m.status = m.d.t("practice.accepted") + runtimeSuffix(msg.result)
		return
	}
	// Wrong answer: keep the raw verdict (case diff etc.) for expand.
	m.fullErr = strings.TrimSpace(msg.result.Raw)
	v := firstLineVerdict(msg.result)
	if v == "" {
		v = m.d.t("practice.notAccepted")
	}
	m.status = "✗ " + v
}

func (m *practiceModel) applyTest(msg testResultMsg) {
	if msg.err != nil {
		m.fullErr = msg.err.Error()
		m.status = m.d.t("practice.testError") + summarizeErr(m.fullErr)
		return
	}
	if msg.result.Passed {
		m.fullErr = ""
		m.status = m.d.t("practice.testPassed")
		return
	}
	// Some case failed (but command succeeded): keep raw output for expand.
	m.fullErr = strings.TrimSpace(msg.result.Raw)
	m.status = m.d.t("practice.testFailed")
}

// --- coaching ---

func (m *practiceModel) applyCoachChunk(text string) {
	m.coachText += text
	m.setCoachContent()
	m.coachVP.GotoBottom()
	if m.expanded && m.expandKind == "coach" {
		m.setExpandContent()
		m.expandVP.GotoBottom()
	}
}

func (m *practiceModel) applyCoachDone() {
	m.coaching = false
	if m.coachTier == domain.TierAnswer {
		m.gaveUp = true
	}
	m.setCoachContent()
	m.coachVP.GotoBottom()
}

func (m *practiceModel) applyCoachErr(err error) {
	m.coaching = false
	m.coachErr = err.Error()
	m.setCoachContent()
}

// setCoachContent wraps the coach display to the panel width and loads it.
func (m *practiceModel) setCoachContent() {
	w := m.coachVP.Width
	if w < 10 {
		w = 76
	}
	m.coachVP.SetContent(wrapWidth(m.coachDisplay(), w))
}

// sectionHeader renders a pane title, highlighting the focused one with ▸.
func (m *practiceModel) sectionHeader(key string, focused bool) string {
	if focused {
		return selectedStyle.Render("▸ " + m.d.t(key))
	}
	return subtleStyle.Render("  " + m.d.t(key))
}

// openExpandForFocus toggles into a full-screen detail view. When both test
// output and coach output exist, `o` opens the focused pane and expand mode can
// switch between available detail targets.
func (m *practiceModel) openExpandForFocus() {
	kind := "statement"
	if m.fullStatement == "" && m.fullErr != "" {
		kind = "error"
	} else if m.fullStatement == "" && (m.coachText != "" || m.coaching) {
		kind = "coach"
	}
	m.openExpandKind(kind)
}

func (m *practiceModel) cycleExpand() {
	kinds := m.expandKinds()
	if len(kinds) < 2 {
		return
	}
	for i, kind := range kinds {
		if kind == m.expandKind {
			m.openExpandKind(kinds[(i+1)%len(kinds)])
			return
		}
	}
	m.openExpandKind(kinds[0])
}

func (m *practiceModel) expandKinds() []string {
	var kinds []string
	if m.fullStatement != "" {
		kinds = append(kinds, "statement")
	}
	if m.fullErr != "" {
		kinds = append(kinds, "error")
	}
	if m.coachText != "" || m.coaching {
		kinds = append(kinds, "coach")
	}
	if len(kinds) == 0 {
		kinds = append(kinds, "statement")
	}
	return kinds
}

func (m *practiceModel) openExpandKind(kind string) {
	m.expandKind = "statement"
	for _, k := range m.expandKinds() {
		if k == kind {
			m.expandKind = kind
			break
		}
	}
	m.expanded = true
	m.setExpandContent()
	m.expandVP.GotoTop()
}

func (m *practiceModel) setExpandContent() {
	content := wrapWidth(m.fullStatement, m.expandContentWidth())
	switch m.expandKind {
	case "error":
		content = m.fullErr
	case "coach":
		content = wrapWidth(m.coachText, m.expandContentWidth())
	}
	m.expandVP.SetContent(content)
}

func (m *practiceModel) expandContentWidth() int {
	w := m.expandVP.Width
	if w < 10 {
		w = 76
	}
	return w
}

func (m *practiceModel) renderExpand() string {
	title := m.d.t("expand.statement")
	switch m.expandKind {
	case "error":
		title = m.d.t("expand.error")
	case "coach":
		title = m.d.t("expand.coach")
	}
	return titleStyle.Render(title) + subtleStyle.Render("  · "+m.d.t("expand.hint")) + "\n\n" +
		m.expandVP.View()
}

// coachDisplay decides what the coach panel shows given the current state.
func (m *practiceModel) coachDisplay() string {
	switch {
	case m.answerConfirm:
		return m.d.t("coach.confirm")
	case m.coachErr != "":
		return statusErrStyle.Render("⚠ " + m.coachErr)
	case m.coachText != "":
		s := m.coachText
		if m.coachTier == domain.TierAnswer && m.gaveUp && !m.coaching {
			s += "\n\n" + subtleStyle.Render(m.d.t("coach.gaveup"))
		}
		return s
	case m.coaching:
		return subtleStyle.Render(m.d.t("coach.thinking"))
	default:
		return subtleStyle.Render(m.d.t("coach.empty"))
	}
}

func (m *practiceModel) view() string {
	if m.expanded {
		return m.renderExpand()
	}
	var b strings.Builder

	// Header: title + difficulty.
	diff := m.problem.Difficulty
	b.WriteString(titleStyle.Render(m.problem.DisplayName()) + "  " +
		difficultyStyle(string(diff)).Render(m.d.difficultyLabel(string(diff))) + "\n\n")

	// Statement summary (cleaned, wrapped to viewport width).
	width := m.viewport.Width
	if width < 10 {
		width = 76
	}
	if m.statement != "" {
		b.WriteString(subtleStyle.Render(wrapWidth(m.statement, width)) + "\n")
	}

	// Code section.
	b.WriteString(m.sectionHeader("section.code", false) + "\n")
	if m.editing {
		b.WriteString(m.editor.View())
	} else {
		b.WriteString(m.viewport.View())
	}

	// Coach section.
	b.WriteString("\n" + m.sectionHeader("section.coach", false) + "\n")
	m.setCoachContent()
	b.WriteString(m.coachVP.View())

	// Status line (single line, truncated — never let it push the layout).
	if m.status != "" {
		style := subtleStyle
		if strings.HasPrefix(m.status, "✓") {
			style = statusOKStyle
		} else if strings.HasPrefix(m.status, "✗") || strings.HasPrefix(m.status, "⚠") {
			style = statusErrStyle
		}
		sw := m.viewport.Width
		if sw < 10 {
			sw = 76
		}
		b.WriteString("\n" + style.Render(truncateWidth(firstLine(m.status), sw)))
	}

	hint := m.d.t("practice.hint")
	if m.editing {
		hint = m.d.t("practice.editorHint")
	}
	b.WriteString(hintStyle.Render("\n" + hint))
	return b.String()
}

// openEditor suspends the TUI and launches $EDITOR on the code file.
func openEditor(deps Deps, path string) tea.Cmd {
	if path == "" {
		// Nothing to edit; no-op.
		return nil
	}
	editor := deps.Config.EditorPath()
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

// --- statement cleaning helpers ---

var (
	mdRefDef   = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*\S.*$`) // [id]: url reference defs (allow indent)
	mdRefLink  = regexp.MustCompile(`\[([^\]]+)\]\[[^\]]*\]`)      // [text][id]
	mdInlLink  = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)       // [text](url)
	mdBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdBacktick = regexp.MustCompile("`([^`]+)`")
)

// cleanStatement strips HTML/markdown noise and the redundant title heading,
// returning at most maxLines non-empty lines joined with newlines.
func cleanStatement(s string, maxLines int) string {
	if s == "" {
		return ""
	}
	if strings.Contains(s, "<") {
		s = stripHTML(s)
	}
	s = stripMarkdown(s)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		// Skip the redundant markdown title line (e.g. "# 3286. ... (Medium)").
		if strings.HasPrefix(l, "#") {
			continue
		}
		out = append(out, l)
		if maxLines > 0 && len(out) >= maxLines {
			break
		}
	}
	return strings.Join(out, "\n")
}

func stripMarkdown(s string) string {
	s = mdRefDef.ReplaceAllString(s, "")
	s = mdRefLink.ReplaceAllString(s, "$1")
	s = mdInlLink.ReplaceAllString(s, "$1")
	s = mdBold.ReplaceAllString(s, "$1")
	s = mdBacktick.ReplaceAllString(s, "$1")
	return s
}

func truncateLines(s string, maxLines int) string {
	if maxLines <= 0 || s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// wrapWidth hard-wraps s so no line exceeds maxCells display columns (accounting
// for double-wide CJK runes), preserving existing newlines.
func wrapWidth(s string, maxCells int) string {
	if maxCells < 4 {
		maxCells = 76
	}
	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		col := 0
		var cur strings.Builder
		for _, r := range line {
			w := runewidth.RuneWidth(r)
			if col+w > maxCells && cur.Len() > 0 {
				out.WriteString(cur.String())
				out.WriteByte('\n')
				cur.Reset()
				col = 0
			}
			cur.WriteRune(r)
			col += w
		}
		out.WriteString(cur.String())
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}

// summarizeErr extracts the most informative single line from a (possibly
// multi-line, leetgo-flavored) error, so the status bar never multi-lines and
// pushes the problem statement off screen.
func summarizeErr(s string) string {
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimLeft(strings.TrimSpace(line), "●×· ")
		low := strings.ToLower(l)
		if strings.Contains(low, "build failed") || strings.Contains(low, "failed to run") ||
			strings.Contains(low, "http ") || strings.Contains(low, "disabled") || strings.Contains(low, "quota") {
			return l
		}
	}
	for _, line := range strings.Split(s, "\n") {
		if l := strings.TrimLeft(strings.TrimSpace(line), "●×· "); l != "" {
			return l
		}
	}
	return s
}

// firstLine returns s up to the first newline.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// truncateWidth truncates a single line to maxCells display columns, appending
// … if shortened.
func truncateWidth(s string, maxCells int) string {
	if maxCells < 4 {
		return s
	}
	var b strings.Builder
	col := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if col+w > maxCells-1 {
			b.WriteString("…")
			return b.String()
		}
		b.WriteRune(r)
		col += w
	}
	return b.String()
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case r == '<' && looksLikeHTMLTagStart(runes, i+1):
			inTag = true
		case r == '>' && inTag:
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func looksLikeHTMLTagStart(runes []rune, i int) bool {
	if i >= len(runes) {
		return false
	}
	r := runes[i]
	return r == '/' || r == '!' || unicode.IsLetter(r)
}

func runtimeSuffix(r domain.SubmitResult) string {
	if r.RuntimeMS > 0 {
		return fmt.Sprintf("  (%d ms)", r.RuntimeMS)
	}
	return ""
}

func firstLineVerdict(r domain.SubmitResult) string {
	raw := strings.TrimSpace(r.Raw)
	if raw == "" {
		return ""
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
