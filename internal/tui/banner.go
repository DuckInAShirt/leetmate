package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// leetmateLogo is a flat pixel rendering of "LEETMATE". It intentionally
// avoids 3-D shadows so the word stays visually centered in terminal screenshots.
const leetmateLogo = `█        ██████  ██████  ████████  ███   ███   █████   ████████  ██████
█        █       █          ██     ████ ████  ██   ██     ██     █
█        █████   █████      ██     ██ ███ ██  ███████     ██     █████
█        █       █          ██     ██  █  ██  ██   ██     ██     █
███████  ██████  ██████     ██     ██     ██  ██   ██     ██     ██████`

// logoGradient tints each row violet → blue → teal, a nod to the gemini-cli
// aesthetic. Hex TrueColor; lipgloss downgrades gracefully on basic terminals.
var logoGradient = []string{
	"#c084fc", "#a78bfa", "#818cf8", "#60a5fa",
	"#38bdf8", "#22d3ee", "#2dd4bf", "#34d399",
}

// renderLogo returns the gradient-tinted ASCII logo when the terminal is wide
// enough to fit it; otherwise "" so the caller can fall back to a plain title.
// Trailing spaces per line are trimmed to keep bubbletea layout from shifting.
func renderLogo(width int) string {
	lines := strings.Split(strings.Trim(leetmateLogo, "\n"), "\n")
	maxLen := 0
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " ")
		if w := lipgloss.Width(lines[i]); w > maxLen {
			maxLen = w
		}
	}
	if width > 0 && width < maxLen+2 {
		return ""
	}
	var b strings.Builder
	for i, ln := range lines {
		color := logoGradient[i%len(logoGradient)]
		line := padRightCells(ln, maxLen)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line))
		b.WriteString("\n")
	}
	return b.String()
}

func padRightCells(s string, width int) string {
	padding := width - lipgloss.Width(s)
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
}
