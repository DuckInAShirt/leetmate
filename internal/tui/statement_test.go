package tui

import (
	"strings"
	"testing"
)

// Test data approximating leetgo's question.md for problem 3286 (CN).
const sample3286 = `# [3286. 穿越网格图的安全路径][link] (Medium)

  [link]: https://leetcode.cn/problems/find-a-safe-walk-through-a-grid/

给你一个下标从 **0** 开始、大小为 ` + "`m x n`" + ` 的二维网格 ` + "`grid`" + ` 。

请你判断是否存在一条路径。
`

func TestCleanStatement(t *testing.T) {
	out := cleanStatement(sample3286, 10)
	// Title heading dropped.
	if strings.Contains(out, "# 3286") {
		t.Errorf("title heading not dropped:\n%s", out)
	}
	// Markdown link definition / backticks / bold stripped.
	for _, bad := range []string{"[link]:", "`", "**", "[link]"} {
		if strings.Contains(out, bad) {
			t.Errorf("markdown noise %q remains:\n%s", bad, out)
		}
	}
	// Multiple lines preserved (not flattened to one line).
	if strings.Count(out, "\n") < 1 {
		t.Errorf("expected multiple lines, got single line:\n%s", out)
	}
	// The actual content survives.
	if !strings.Contains(out, "给你一个下标从") {
		t.Errorf("content lost:\n%s", out)
	}
}

func TestWrapWidthRespectsDoubleWidth(t *testing.T) {
	// 6 CJK runes = 12 display cells; wrapping at 10 cells must break.
	out := wrapWidth("中文中文中文", 10)
	if !strings.Contains(out, "\n") {
		t.Errorf("expected wrapping for double-width runes, got %q", out)
	}
}
