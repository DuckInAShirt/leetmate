package tui

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	codeKeywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C084FC")).Bold(true)
	codeStringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#86EFAC"))
	codeCommentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true)
)

func highlightCode(code, lang string) string {
	keywords := syntaxKeywords(lang)
	if len(keywords) == 0 {
		return code
	}
	set := make(map[string]bool, len(keywords))
	for _, kw := range keywords {
		set[kw] = true
	}
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lines[i] = highlightLine(line, set)
	}
	return strings.Join(lines, "\n")
}

func highlightLine(line string, keywords map[string]bool) string {
	var b strings.Builder
	for i := 0; i < len(line); {
		if strings.HasPrefix(line[i:], "//") || strings.HasPrefix(line[i:], "#") {
			b.WriteString(codeCommentStyle.Render(line[i:]))
			break
		}
		if line[i] == '"' || line[i] == '\'' || line[i] == '`' {
			end := scanString(line, i)
			b.WriteString(codeStringStyle.Render(line[i:end]))
			i = end
			continue
		}
		r, size := rune(line[i]), 1
		if r >= 0x80 {
			r, size = utf8Rune(line[i:])
		}
		if isIdentStart(r) {
			end := i + size
			for end < len(line) {
				r2, size2 := rune(line[end]), 1
				if r2 >= 0x80 {
					r2, size2 = utf8Rune(line[end:])
				}
				if !isIdentPart(r2) {
					break
				}
				end += size2
			}
			word := line[i:end]
			if keywords[word] {
				b.WriteString(codeKeywordStyle.Render(word))
			} else {
				b.WriteString(word)
			}
			i = end
			continue
		}
		b.WriteString(line[i : i+size])
		i += size
	}
	return b.String()
}

func scanString(line string, start int) int {
	quote := line[start]
	for i := start + 1; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == quote {
			return i + 1
		}
	}
	return len(line)
}

func utf8Rune(s string) (rune, int) {
	for _, r := range s {
		return r, len(string(r))
	}
	return rune(s[0]), 1
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func syntaxKeywords(lang string) []string {
	switch syntaxLang(lang) {
	case "go":
		return []string{"break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var"}
	case "python":
		return []string{"False", "None", "True", "and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"}
	case "cpp", "c":
		return []string{"auto", "bool", "break", "case", "char", "class", "const", "continue", "default", "delete", "do", "double", "else", "enum", "extern", "float", "for", "if", "inline", "int", "long", "namespace", "new", "private", "protected", "public", "return", "short", "signed", "sizeof", "static", "struct", "switch", "template", "this", "typedef", "typename", "union", "unsigned", "using", "void", "while"}
	case "java", "kotlin", "csharp", "javascript", "typescript", "rust", "ruby", "swift":
		return []string{"as", "async", "await", "break", "case", "catch", "class", "const", "continue", "def", "default", "do", "else", "enum", "false", "fn", "for", "func", "function", "if", "import", "in", "interface", "let", "match", "new", "nil", "null", "package", "private", "protected", "public", "return", "static", "struct", "switch", "this", "throw", "true", "try", "type", "val", "var", "void", "while"}
	default:
		return nil
	}
}

func codeLanguage(deps Deps, path string) string {
	if deps.Leetgo != nil {
		if lang := syntaxLang(deps.Leetgo.Lang()); lang != "" {
			return lang
		}
	}
	return syntaxLang(strings.TrimPrefix(filepath.Ext(path), "."))
}

func syntaxLang(lang string) string {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return "go"
	case "python", "python3", "py":
		return "python"
	case "cpp", "c++", "cc", "cxx", "hpp":
		return "cpp"
	case "c":
		return "c"
	case "java":
		return "java"
	case "rust", "rs":
		return "rust"
	case "javascript", "js":
		return "javascript"
	case "typescript", "ts":
		return "typescript"
	case "kotlin", "kt":
		return "kotlin"
	case "ruby", "rb":
		return "ruby"
	case "swift":
		return "swift"
	case "csharp", "c#", "cs":
		return "csharp"
	default:
		return lang
	}
}
