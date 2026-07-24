package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"

	"github.com/DuckInAShirt/leetmate/internal/domain"
	"github.com/DuckInAShirt/leetmate/internal/leetgo"
)

func TestACMApplyRunDone(t *testing.T) {
	m := acmModel{d: loadDict("zh")}
	m.applyRun(runResultMsg{stdout: "42\n"})
	if !strings.HasPrefix(m.status, "✓") {
		t.Fatalf("status=%q, want ✓ prefix", m.status)
	}
	if !strings.Contains(m.output, "42") {
		t.Fatalf("output=%q, want stdout 42 surfaced", m.output)
	}
}

func TestACMApplyRunStderr(t *testing.T) {
	m := acmModel{d: loadDict("zh")}
	m.applyRun(runResultMsg{stderr: "NameError: name 'x' is not defined"})
	// stderr (tracebacks) must surface in the output panel for debugging.
	if !strings.Contains(m.output, "NameError") {
		t.Fatalf("output=%q, want stderr surfaced", m.output)
	}
}

func TestACMApplyRunTimeout(t *testing.T) {
	m := acmModel{d: loadDict("zh")}
	m.applyRun(runResultMsg{err: leetgo.ErrACMTimeout})
	if !strings.Contains(m.status, "超时") && !strings.Contains(m.status, "timed") {
		t.Fatalf("timeout status=%q", m.status)
	}
}

func TestACMApplyRunError(t *testing.T) {
	m := acmModel{d: loadDict("zh")}
	m.applyRun(runResultMsg{err: errors.New("exec: boom")})
	if !strings.HasPrefix(m.status, "⚠") {
		t.Fatalf("status=%q, want ⚠ prefix", m.status)
	}
}

func TestACMCycleExpand(t *testing.T) {
	m := acmModel{d: loadDict("zh"), expandKind: "statement"}
	m.expandVP = viewport.New(80, 20)
	m.cycleExpand()
	if m.expandKind != "code" {
		t.Fatalf("expandKind=%q, want code after one cycle", m.expandKind)
	}
	m.cycleExpand()
	if m.expandKind != "stdin" {
		t.Fatalf("expandKind=%q, want stdin after two cycles", m.expandKind)
	}
	m.cycleExpand()
	if m.expandKind != "output" {
		t.Fatalf("expandKind=%q, want output after three cycles", m.expandKind)
	}
}

func TestACMViewRender(t *testing.T) {
	m := acmModel{
		d:       loadDict("zh"),
		problem: domain.Problem{ProblemMeta: domain.ProblemMeta{FrontendID: "1", Title: "Two Sum"}},
	}
	m.codeEditor = newCodeEditor("print(1)\n")
	m.stdinEditor = newCodeEditor("")
	m.expandVP = viewport.New(80, 20)
	m.resize(100, 40)
	out := m.view()
	for _, want := range []string{"Two Sum", "print(1)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("compact view missing %q:\n%s", want, out)
		}
	}
}
