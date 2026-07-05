package tui

import "testing"

func TestHighlightCodeFallbackForUnknownLanguage(t *testing.T) {
	code := "plain text"
	if got := highlightCode(code, "definitely-not-a-language"); got != code {
		t.Fatalf("unknown language should fall back to plain code, got %q", got)
	}
}

func TestSyntaxLangMapsLeetgoNames(t *testing.T) {
	cases := map[string]string{
		"python3": "python",
		"c++":     "cpp",
		"rs":      "rust",
		"ts":      "typescript",
		"c#":      "csharp",
	}
	for in, want := range cases {
		if got := syntaxLang(in); got != want {
			t.Fatalf("syntaxLang(%q) = %q, want %q", in, got, want)
		}
	}
}
