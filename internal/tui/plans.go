package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderPlanList shows all study plans with their progress.
func (m Model) renderPlanList() string {
	var b strings.Builder
	b.WriteString(brandStyle.Render("📋 "+m.d.t("menu.plans")) + "\n\n")
	if len(m.planList) == 0 {
		b.WriteString(subtleStyle.Render("（暂无题单）"))
		return b.String()
	}
	cursor := clampCursor(m.planCursor, len(m.planList))
	ctx := context.Background()
	for i, p := range m.planList {
		done, total := m.deps.Plans.Progress(ctx, p.ID)
		marker := "  "
		label := normalStyle.Render(p.Title)
		if i == cursor {
			marker = "▸ "
			label = selectedStyle.Render(p.Title)
		}
		prog := subtleStyle.Render(fmt.Sprintf("  %d/%d", done, total))
		b.WriteString(marker + label + prog + "\n")
		if i == cursor && p.Description != "" {
			b.WriteString("    " + subtleStyle.Render(p.Description) + "\n")
		}
	}
	b.WriteString(hintStyle.Render("\n" + m.d.t("plan.hint.list")))
	return b.String()
}

// renderPlanItems shows the problems in the current plan with done markers.
// Only a window around the cursor is rendered (the full 100-item list would
// overflow the screen and hide the cursor).
func (m Model) renderPlanItems() string {
	var b strings.Builder
	p := m.deps.Plans.Plan(m.curPlanID)
	if p == nil {
		return ""
	}
	doneCount := 0
	for _, fid := range m.planItems {
		if m.planDone[fid] {
			doneCount++
		}
	}
	b.WriteString(titleStyle.Render(p.Title) +
		subtleStyle.Render(fmt.Sprintf("  %d/%d", doneCount, len(m.planItems))) + "\n\n")

	// Compute the visible window so the cursor always stays on screen.
	h := m.height
	if h < 10 {
		h = 40
	}
	visible := h - 8 // header(2) + blank + "↑/↓ more" lines + hint(2) + margin
	if visible < 5 {
		visible = 5
	}
	total := len(m.planItems)
	cursor := clampCursor(m.planCursor, total)
	start := cursor - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > total {
		end = total
	}
	if end-start < visible && start > 0 {
		start = end - visible
		if start < 0 {
			start = 0
		}
	}

	if start > 0 {
		b.WriteString(subtleStyle.Render(fmt.Sprintf("  ↑ 还有 %d 题\n", start)))
	}
	for i := start; i < end; i++ {
		fid := m.planItems[i]
		mark := "○"
		if m.planDone[fid] {
			mark = "✓"
		}
		line := fmt.Sprintf("#%s  %s", fid, mark)
		marker := "  "
		if i == cursor {
			marker = "▸ "
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}
		b.WriteString(marker + line + "\n")
	}
	if end < total {
		b.WriteString(subtleStyle.Render(fmt.Sprintf("  ↓ 还有 %d 题\n", total-end)))
	}
	b.WriteString(hintStyle.Render("\n" + m.d.t("plan.hint.items")))
	return b.String()
}

func (m Model) updatePlanList(str string) (tea.Model, tea.Cmd) {
	m.planCursor = clampCursor(m.planCursor, len(m.planList))
	switch str {
	case "j", "down":
		if m.planCursor < len(m.planList)-1 {
			m.planCursor++
		}
	case "k", "up":
		if m.planCursor > 0 {
			m.planCursor--
		}
	case "enter":
		if len(m.planList) == 0 {
			return m, nil
		}
		p := m.planList[m.planCursor]
		m.curPlanID = p.ID
		m.planItems = p.Items
		m.planDone = m.deps.Plans.DoneSet(context.Background(), p.ID)
		// Jump the cursor to the first not-yet-done problem.
		m.planCursor = 0
		for i, fid := range m.planItems {
			if !m.planDone[fid] {
				m.planCursor = i
				break
			}
		}
		m.view = viewPlanItems
	case "b", "esc":
		m.view = viewMenu
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updatePlanItems(str string) (tea.Model, tea.Cmd) {
	m.planCursor = clampCursor(m.planCursor, len(m.planItems))
	switch str {
	case "j", "down":
		if m.planCursor < len(m.planItems)-1 {
			m.planCursor++
		}
	case "k", "up":
		if m.planCursor > 0 {
			m.planCursor--
		}
	case "enter":
		if len(m.planItems) == 0 {
			return m, nil
		}
		fid := m.planItems[m.planCursor]
		pc := &planCtx{planID: m.curPlanID, fid: fid}
		m.busy = true
		m.err = ""
		return m, pickCmd(m.deps, fid, pc)
	case "b", "esc":
		// Refresh progress when returning (a submission may have marked one done).
		m.planDone = m.deps.Plans.DoneSet(context.Background(), m.curPlanID)
		m.planCursor = m.currentPlanListIndex()
		m.view = viewPlanList
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func clampCursor(cursor, total int) int {
	if total <= 0 || cursor < 0 {
		return 0
	}
	if cursor >= total {
		return total - 1
	}
	return cursor
}

func (m Model) currentPlanListIndex() int {
	for i, p := range m.planList {
		if p.ID == m.curPlanID {
			return i
		}
	}
	return clampCursor(m.planCursor, len(m.planList))
}
