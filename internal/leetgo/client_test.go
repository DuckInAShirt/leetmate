package leetgo

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

func TestClientTestPreservesOutputOnExitError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake binary is Unix-only")
	}
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte("code:\n  lang: go\n"), 0o644); err != nil {
		t.Fatalf("write leetgo.yaml: %v", err)
	}
	bin := filepath.Join(workspace, "fake-leetgo")
	script := `#!/bin/sh
if [ "$1" = "test" ]; then
  printf '✘ Wrong Answer\n\nPassed cases:  ✘✔\nInput:         [0,1,0,2,1,0,1,3,2,1,2,1]\nOutput:        -8\n'
  printf 'Expected:      6\n' >&2
  exit 1
fi
exit 0
`
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	c, err := New(config.LeetgoConfig{Workspace: workspace, Binary: bin})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := c.Test(context.Background(), "42")
	if err == nil {
		t.Fatal("expected leetgo exit error")
	}
	if res.Passed {
		t.Fatal("expected failed test result")
	}
	for _, want := range []string{"Wrong Answer", "Input:", "Output:", "Expected:"} {
		if !strings.Contains(res.Raw, want) {
			t.Fatalf("raw output missing %q: %q", want, res.Raw)
		}
	}
	if !strings.Contains(err.Error(), "Wrong Answer") || !strings.Contains(err.Error(), "Expected:") {
		t.Fatalf("error should include combined output, got %q", err.Error())
	}
}

func TestResolveProblemDirPrefersCurrentLanguage(t *testing.T) {
	workspace := t.TempDir()
	// Simulate a workspace that holds both go/ and python/ scaffolds for the same
	// problem — common after the user switches code.lang while history remains.
	for _, lang := range []string{"go", "python"} {
		dir := filepath.Join(workspace, lang, "0239.sliding-window-maximum")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", lang, err)
		}
		ext := ".go"
		if lang == "python" {
			ext = ".py"
		}
		if err := os.WriteFile(filepath.Join(dir, "solution"+ext), []byte("# scaffold\n"), 0o644); err != nil {
			t.Fatalf("write solution: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte("code:\n  lang: python\n"), 0o644); err != nil {
		t.Fatalf("write leetgo.yaml: %v", err)
	}
	c, err := New(config.LeetgoConfig{Workspace: workspace})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	dir, err := c.resolveProblemDir("sliding-window-maximum")
	if err != nil {
		t.Fatalf("resolveProblemDir: %v", err)
	}
	// Must pick the python/ subdir (the configured language), not the
	// lexically-first go/ — otherwise codeFile finds no .py and the scaffold
	// renders empty (the 239 bug).
	if !strings.HasSuffix(filepath.ToSlash(dir), "python/0239.sliding-window-maximum") {
		t.Fatalf("resolveProblemDir = %q, want the python/ subdir for lang=python", dir)
	}
	if c.codeFile(dir) == "" {
		t.Fatalf("codeFile returned empty for resolved dir %s", dir)
	}
}
