//go:build integration

package leetgo

import (
	"context"
	"os"
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

// TestIntegrationReal exercises the real leetgo binary against a workspace.
// Run with:
//
//	LEETMATE_TEST_WORKSPACE=/path/to/leetgo/workspace \
//	go test -tags=integration -v -run TestIntegrationReal ./internal/leetgo/
//
// It only touches read-only paths (info / locate code file / read code) so it
// won't create submissions or modify your LeetCode account.
func TestIntegrationReal(t *testing.T) {
	ws := os.Getenv("LEETMATE_TEST_WORKSPACE")
	if ws == "" {
		t.Skip("set LEETMATE_TEST_WORKSPACE to run")
	}
	ctx := context.Background()

	c, err := New(config.LeetgoConfig{Workspace: ws, Binary: "leetgo"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.CheckAvailable(); err != nil {
		t.Fatalf("CheckAvailable: %v", err)
	}

	// leetgo info --format json (read-only; works anonymously).
	meta, err := c.Info(ctx, "3286")
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	t.Logf("meta: slug=%q frontend_id=%q title=%q difficulty=%s tags=%v paid=%v",
		meta.Slug, meta.FrontendID, meta.Title, meta.Difficulty, meta.Tags, meta.IsPaidOnly)
	if meta.Slug == "" || meta.FrontendID == "" || meta.Title == "" {
		t.Errorf("meta has empty core fields: %+v", meta)
	}

	// Locate the generated directory + code file (assumes the problem was
	// generated before; pick today / pick 3286 if not).
	dir, err := c.resolveProblemDir(meta.Slug)
	if err != nil {
		t.Skipf("resolveProblemDir: %v (run `leetgo pick %s` first)", err, meta.FrontendID)
	}
	codePath := c.codeFile(dir)
	t.Logf("dir=%s", dir)
	t.Logf("codeFile=%s", codePath)
	if codePath == "" {
		t.Fatalf("codeFile returned empty for %s", dir)
	}

	code, err := c.ReadCode(meta.Slug)
	if err != nil {
		t.Fatalf("ReadCode: %v", err)
	}
	t.Logf("code length=%d first120=%q", len(code), trunc(code, 120))
	if len(code) == 0 {
		t.Error("ReadCode returned empty content")
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
