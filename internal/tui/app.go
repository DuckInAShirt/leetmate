// Package tui implements LeetMate's terminal interface (bubbletea). The MVP
// keeps a small state machine: a top menu and a practice view. Coaching (M1+)
// and review (M3) plug in as additional views driven by the same root model.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/leetgo"
	"github.com/DuckInAShirt/leetmate/internal/llm"
	"github.com/DuckInAShirt/leetmate/internal/store"
	"github.com/DuckInAShirt/leetmate/internal/studyplan"
)

// Deps bundles the services the TUI needs.
type Deps struct {
	Leetgo *leetgo.Client
	Store  *store.Store
	Config *config.Config
	Coach  *coach.Coach        // nil when no LLM key is configured — coaching is then disabled
	Plans  *studyplan.Service
}

const (
	viewMenu = iota
	viewPractice
	viewPlanList  // choose a study plan
	viewPlanItems // choose a problem within a plan
)

type menuItem struct {
	label string
	desc  string
}

// Model is the root bubbletea model.
type Model struct {
	deps   Deps
	d      dict
	view   int
	cursor int
	menu   []menuItem

	practice *practiceModel

	// coachStream is the live LLM stream while coaching is in progress; nil
	// otherwise. See coachStartCmd / listenCoach.
	coachStream <-chan llm.Chunk

	// study-plan navigation state
	planList   []*studyplan.Plan
	planItems  []string
	planDone   map[string]bool
	curPlanID  string
	planCursor int

	width  int
	height int
	err    string
	notice string
	busy   bool // true while an async leetgo command is in flight
}

// New returns the initial model.
func New(deps Deps) Model {
	d := loadDict(deps.Config.Language)
	return Model{
		deps: deps,
		d:    d,
		view: viewMenu,
		menu: []menuItem{
			{label: d.t("menu.today"), desc: d.t("menu.today.desc")},
			{label: d.t("menu.plans"), desc: d.t("menu.plans.desc")},
			{label: d.t("menu.review"), desc: d.t("menu.review.desc")},
			{label: d.t("menu.quit"), desc: ""},
		},
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.practice != nil {
			m.practice.resize(msg.Width, msg.Height)
		}
		return m, nil

	case pickResultMsg:
		m.busy = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		pm := newPracticeModel(m.deps, msg.problem)
		pm.planCtx = msg.planCtx
		pm.resize(m.width, m.height)
		m.practice = &pm
		m.view = viewPractice
		m.err = ""
		return m, nil

	case submitResultMsg:
		if m.practice != nil {
			m.practice.applySubmit(msg)
		}
		if msg.result.Accepted && m.practice != nil && m.practice.planCtx != nil {
			pc := *m.practice.planCtx
			return m, markDoneCmd(m.deps, pc)
		}
		return m, nil

	case planMarkedMsg:
		// Keep the cached done-set fresh so the plan view reflects it on return.
		if msg.planID == m.curPlanID && m.planDone != nil {
			m.planDone[msg.fid] = true
		}
		return m, nil

	case testResultMsg:
		if m.practice != nil {
			m.practice.applyTest(msg)
		}
		return m, nil

	case editorDoneMsg:
		if m.practice != nil {
			m.practice.reloadCode(m.deps)
		}
		return m, nil

	case coachStartedMsg:
		m.coachStream = msg.stream
		if m.practice != nil {
			m.practice.coachTier = msg.tier
		}
		return m, listenCoach(msg.stream)

	case coachChunkMsg:
		if m.practice != nil {
			m.practice.applyCoachChunk(msg.text)
		}
		if m.coachStream != nil {
			return m, listenCoach(m.coachStream)
		}
		return m, nil

	case coachDoneMsg:
		m.coachStream = nil
		if m.practice != nil {
			m.practice.applyCoachDone()
			if txt := m.practice.coachText; txt != "" {
				_, _ = m.deps.Store.InsertConversation(context.Background(), domain.Conversation{
					Slug:      m.practice.problem.Slug,
					Tier:      m.practice.coachTier,
					Role:      domain.RoleAssistant,
					Content:   txt,
					CreatedAt: time.Now(),
				})
			}
		}
		return m, nil

	case coachErrMsg:
		m.coachStream = nil
		if m.practice != nil {
			m.practice.applyCoachErr(msg.err)
		}
		return m, nil
	}

	// Key handling.
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	str := key.String()

	// Global: ctrl+c always quits.
	if str == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.view {
	case viewMenu:
		return m.updateMenu(str)
	case viewPlanList:
		return m.updatePlanList(str)
	case viewPlanItems:
		return m.updatePlanItems(str)
	case viewPractice:
		if m.practice == nil {
			return m, nil
		}
		return m.practiceUpdate(key, str)
	}
	return m, nil
}

func (m Model) updateMenu(str string) (tea.Model, tea.Cmd) {
	switch str {
	case "j", "down":
		m.cursor = (m.cursor + 1) % len(m.menu)
	case "k", "up":
		m.cursor = (m.cursor - 1 + len(m.menu)) % len(m.menu)
	case "enter":
		switch m.cursor {
		case 0: // Today's problem
			m.busy = true
			m.err = ""
			return m, pickCmd(m.deps, "today", nil)
		case 1: // Study plans
			if m.deps.Plans != nil {
				m.planList = m.deps.Plans.Plans()
				m.planCursor = 0
				m.view = viewPlanList
				m.err = ""
			}
		case 2: // Due for review
			m.notice = m.d.t("menu.reviewNotice")
		case 3: // Quit
			return m, tea.Quit
		}
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) practiceUpdate(key tea.KeyMsg, str string) (tea.Model, tea.Cmd) {
	p := m.practice

	// Expand mode: dedicated key handling (scroll the detail pane, o/esc to exit).
	if p.expanded {
		switch str {
		case "o", "esc":
			p.expanded = false
			return m, nil
		case "up", "k":
			p.expandVP.LineUp(1)
		case "down", "j":
			p.expandVP.LineDown(1)
		case "pgup":
			p.expandVP.HalfViewUp()
		case "pgdown":
			p.expandVP.HalfViewDown()
		case "q":
			return m, tea.Quit
		}
		return m, nil
	}

	// Scrolling and pane focus.
	switch str {
	case "tab":
		if p.focus == focusCode {
			p.focus = focusCoach
		} else {
			p.focus = focusCode
		}
		return m, nil
	case "up", "k":
		m.scrollFocused(p, -1, false)
		return m, nil
	case "down", "j":
		m.scrollFocused(p, 1, false)
		return m, nil
	case "pgup":
		m.scrollFocused(p, -1, true)
		return m, nil
	case "pgdown":
		m.scrollFocused(p, 1, true)
		return m, nil
	}

	switch str {
	case "e":
		return m, openEditor(m.deps, p.problem.CodePath)
	case "t":
		p.status = m.d.t("practice.testing")
		return m, testCmd(m.deps, p.qid())
	case "s":
		p.status = m.d.t("practice.submitting")
		return m, submitCmd(m.deps, p.problem.Slug, p.qid(), p.gaveUp)
	case "1":
		return m.startCoach(domain.TierHint, false)
	case "2":
		return m.startCoach(domain.TierNudge, false)
	case "3":
		return m.startCoach(domain.TierReview, false)
	case "4":
		p.answerConfirm = true
		p.coachErr = ""
		return m, nil
	case "o":
		p.openExpand()
		return m, nil
	case "y", "Y":
		if p.answerConfirm {
			p.answerConfirm = false
			return m.startCoach(domain.TierAnswer, true)
		}
	case "n", "N":
		if p.answerConfirm {
			p.answerConfirm = false
			return m, nil
		}
	case "b", "esc":
		m.practice = nil
		m.view = viewMenu
		return m, nil
	case "q":
		return m, tea.Quit
	}
	_ = key
	return m, nil
}

// scrollFocused scrolls the focused pane (code or coach) by one line or half page.
func (m Model) scrollFocused(p *practiceModel, dir int, page bool) {
	vp := &p.viewport
	if p.focus == focusCoach {
		vp = &p.coachVP
	}
	switch {
	case page && dir < 0:
		vp.HalfViewUp()
	case page:
		vp.HalfViewDown()
	case dir < 0:
		vp.LineUp(1)
	default:
		vp.LineDown(1)
	}
}

// startCoach kicks off a coaching turn at the given tier. gaveUp marks the
// attempt (used by Answer tier so submit records it).
func (m Model) startCoach(tier domain.Tier, gaveUp bool) (tea.Model, tea.Cmd) {
	p := m.practice
	if m.deps.Coach == nil {
		p.coachErr = "LLM 未配置：请在 config.yaml 配置 provider，并设置 API key 环境变量"
		return m, nil
	}
	// Always re-read code from disk so coaching reflects the latest edits, even
	// if the learner changed the file outside leetmate's `e` command.
	p.reloadCode(m.deps)
	p.coaching = true
	p.coachTier = tier
	p.coachText = ""
	p.coachErr = ""
	p.coachVP.SetContent("")
	if gaveUp {
		p.gaveUp = true
	}
	hist, _ := m.deps.Store.RecentConversations(context.Background(), p.problem.Slug, m.deps.Config.LLM.MaxHistory)
	req := coach.Request{
		Tier:     tier,
		Problem:  p.problem,
		Code:     p.code,
		Lang:     m.deps.Leetgo.Lang(),
		History:  hist,
	}
	return m, coachStartCmd(m.deps, req)
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.view {
	case viewPractice:
		if m.practice != nil {
			return m.practice.view()
		}
	case viewPlanList:
		return m.renderPlanList()
	case viewPlanItems:
		return m.renderPlanItems()
	}
	return m.menuView()
}

func (m Model) menuView() string {
	var b strings.Builder
	b.WriteString(brandStyle.Render("⚔  LeetMate") + subtleStyle.Render("  "+m.d.t("brand.subtitle")) + "\n\n")

	if m.busy {
		b.WriteString(subtleStyle.Render(m.d.t("menu.busy")) + "\n")
	}
	for i, item := range m.menu {
		marker := "  "
		line := item.label
		if i == m.cursor {
			marker = "▸ "
			line = selectedStyle.Render(item.label)
		} else {
			line = normalStyle.Render(item.label)
		}
		desc := ""
		if item.desc != "" {
			desc = subtleStyle.Render("  " + item.desc)
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", marker, line, desc))
	}
	b.WriteString(hintStyle.Render("\n" + m.d.t("menu.hint")))

	if m.notice != "" {
		b.WriteString("\n" + subtleStyle.Render(m.notice))
	}
	if m.err != "" {
		b.WriteString("\n" + statusErrStyle.Render("⚠ "+m.err))
	}
	return b.String()
}
