// Command leetmate launches the LeetMate TUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/leetgo"
	"github.com/DuckInAShirt/leetmate/internal/llm"
	"github.com/DuckInAShirt/leetmate/internal/store"
	"github.com/DuckInAShirt/leetmate/internal/studyplan"
	"github.com/DuckInAShirt/leetmate/internal/tui"
)

// Build-time variables, injected via -ldflags by GoReleaser. Defaults keep
// `go build` working locally without a release pipeline.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("leetmate %s (%s) built %s\n", version, commit, date)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "leetmate:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, _, err := config.Load()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(cfg.Dir, 0o755)

	lc, err := leetgo.New(cfg.Leetgo)
	if err != nil {
		return err
	}
	if err := lc.CheckAvailable(); err != nil {
		// leetgo not set up — surface a clear message rather than a stack trace.
		fmt.Fprintln(os.Stderr, "leetmate:", err)
		fmt.Fprintln(os.Stderr, "configure leetgo in", cfg.Leetgo.Workspace, "then re-run leetmate.")
		return nil
	}

	st, err := store.Open(cfg.DB.Path)
	if err != nil {
		return err
	}
	defer st.Close()

	// LLM is optional: without an API key, leetgo-only mode still works
	// (pick/test/submit), coaching is just disabled.
	var cch *coach.Coach
	if provider, err := llm.New(cfg.LLM); err == nil {
		cch = coach.New(provider)
	} else {
		fmt.Fprintln(os.Stderr, "leetmate: LLM 未启用 —", err)
		fmt.Fprintln(os.Stderr, "          仍可选题/测试/提交，辅导功能不可用。")
	}

	deps := tui.Deps{Leetgo: lc, Store: st, Config: &cfg, Coach: cch}
	// Study plans (builtin + user-defined under <config dir>/studyplans/).
	plans, err := studyplan.All(filepath.Join(cfg.Dir, "studyplans"))
	if err != nil {
		return fmt.Errorf("load study plans: %w", err)
	}
	deps.Plans = studyplan.NewService(st, plans)

	p := tea.NewProgram(tui.New(deps), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
