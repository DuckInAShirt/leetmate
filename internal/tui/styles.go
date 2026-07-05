package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// Centralized styles. These are intentionally simple for the MVP; the M4
// polish pass turns them into the showy palette that sells the demo gif.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DD3FC")).
			Padding(0, 1)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA")).
			MarginBottom(1)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FDE68A")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBD5E1"))

	difficultyStyle = func(diff string) lipgloss.Style {
		switch diff {
		case "Easy":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#86EFAC"))
		case "Medium":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#FCD34D"))
		case "Hard":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#FCA5A5"))
		default:
			return lipgloss.NewStyle()
		}
	}

	statusOKStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#86EFAC"))
	statusErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FCA5A5"))
	hintStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).MarginTop(1)

	homeCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2)

	homePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(1, 0, 0)

	homeMenuStyle = lipgloss.NewStyle().PaddingLeft(1)

	homePromptStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#60A5FA")).
			Foreground(lipgloss.Color("#CBD5E1")).
			Padding(0, 1)

	homePromptMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C084FC")).
				Bold(true)

	homeTipStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A5B4FC"))
	homeStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
)

func codeEditorFocusedStyle() textarea.Style {
	return textarea.Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       lipgloss.NewStyle().Background(lipgloss.Color("#111827")),
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.Color("#FDE68A")),
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.Color("#0F172A")),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")),
		Placeholder:      subtleStyle,
		Prompt:           subtleStyle,
		Text:             normalStyle,
	}
}

func codeEditorBlurredStyle() textarea.Style {
	return textarea.Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       normalStyle,
		CursorLineNumber: subtleStyle,
		EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.Color("#0F172A")),
		LineNumber:       subtleStyle,
		Placeholder:      subtleStyle,
		Prompt:           subtleStyle,
		Text:             normalStyle,
	}
}
