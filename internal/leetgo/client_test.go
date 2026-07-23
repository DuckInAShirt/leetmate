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

// TestResolveProblemDirPrefersConfiguredLang guards against the multi-language
// regression: when a workspace has both go/<id>.<slug> and python/<id>.<slug>,
// the dir under the configured code language must win. Otherwise codeFile
// cannot find the source file and the practice view comes up empty.
func TestResolveProblemDirPrefersConfiguredLang(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"),
		[]byte("code:\n  lang: python\n"), 0o644); err != nil {
		t.Fatalf("write leetgo.yaml: %v", err)
	}
	const slug = "minimum-window-substring"
	// Seed both go/ and python/ dirs (go sorts before python, so a naive
	// first-match Walk would pick the wrong one).
	for _, lang := range []string{"go", "python"} {
		dir := filepath.Join(workspace, lang, "0076."+slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "question.md"), []byte("# 76\n"), 0o644); err != nil {
			t.Fatalf("write question.md: %v", err)
		}
	}
	c, err := New(config.LeetgoConfig{Workspace: workspace, Binary: "fake-leetgo"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got, err := c.resolveProblemDir(slug)
	if err != nil {
		t.Fatalf("resolveProblemDir: %v", err)
	}
	want := filepath.Join(workspace, "python", "0076."+slug)
	if got != want {
		t.Fatalf("resolveProblemDir = %q, want %q (configured lang is python, not go)", got, want)
	}
}
