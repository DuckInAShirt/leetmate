package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// leetmateLogo is the 3-D figlet rendering of "leetmate", baked once with
// `figlet -f 3-d`. Kept as a constant so the binary stays self-contained
// (no font files at runtime).
const leetmateLogo = `  **                   **                           **
 /**                  /**                          /**
 /**  *****   *****  ****** **********   ******   ******  *****
 /** **///** **///**///**/ //**//**//** //////** ///**/  **//**
 /**/*******/*******  /**   /** /** /**  *******   /**  /*******
 /**/**//// /**////   /**   /** /** /** **////**   /**  /**////
 ***//******//******  //**  *** /** /**//********  //** //******
///  //////  //////    //  ///  //  //  ////////    //   //////`

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
	lines := strings.Split(leetmateLogo, "\n")
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
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(ln))
		b.WriteString("\n")
	}
	return b.String()
}
