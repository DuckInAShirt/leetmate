// Command leetmate launches the LeetMate TUI.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"

	"github.com/DuckInAShirt/leetmate/internal/coach"
	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/doctor"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			exitIfErr(runInit(os.Args[2:], os.Stdout))
			return
		case "config":
			exitIfErr(runConfig(os.Args[2:], os.Stdout))
			return
		case "doctor":
			exitIfErr(runDoctor(os.Args[2:], os.Stdout))
			return
		}
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("leetmate %s (%s) built %s\n", version, commit, date)
		return
	}
	if err := run(os.Stdin, os.Stdout, os.Stderr, term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())); err != nil {
		if !errors.Is(err, errDoctorFailed) {
			fmt.Fprintln(os.Stderr, "leetmate:", err)
		}
		os.Exit(1)
	}
}

func exitIfErr(err error) {
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return
	}
	if !errors.Is(err, errDoctorFailed) {
		fmt.Fprintln(os.Stderr, "leetmate:", err)
	}
	os.Exit(1)
}

func run(in io.Reader, out, errOut io.Writer, interactive bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := runFirstRun(in, out, cwd, interactive); err != nil {
		return err
	}

	cfg, configPath, err := config.Load()
	if err != nil {
		report := doctor.Report{Checks: []doctor.Check{{ID: "config", Level: doctor.Fail, Reason: "unreadable", Value: configPath, Extra: err.Error()}}}
		printDoctorReport(errOut, report, cfg.Language)
		return errDoctorFailed
	}
	report := doctor.Run(cfg, configPath, cwd)
	if report.Workspace != "" {
		cfg.Leetgo.Workspace = report.Workspace
	}
	if report.Binary != "" {
		cfg.Leetgo.Binary = report.Binary
	}
	if report.HasFailures() {
		printDoctorReport(errOut, report, cfg.Language)
		return errDoctorFailed
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DB.Path), 0o755); err != nil {
		return err
	}

	lc, err := leetgo.New(cfg.Leetgo)
	if err != nil {
		return err
	}

	st, err := store.Open(cfg.DB.Path)
	if err != nil {
		return err
	}
	defer st.Close()

	// LLM is optional: without an API key, leetgo-only mode still works
	// (pick/test/submit), coaching is just disabled.
	var cch *coach.Coach
	if provider, llmErr := llm.New(cfg.LLM); llmErr == nil {
		cch = coach.New(provider)
	} else if errors.Is(llmErr, llm.ErrMissingAPIKey) {
		if cfg.Language == "en" {
			fmt.Fprintln(errOut, "leetmate: LLM is disabled —", llmErr)
			fmt.Fprintln(errOut, "          Pick, test, and submit still work; Coach is unavailable.")
		} else {
			fmt.Fprintln(errOut, "leetmate: LLM 未启用 —", llmErr)
			fmt.Fprintln(errOut, "          仍可选题/测试/提交，辅导功能不可用。")
		}
	} else {
		return llmErr
	}

	deps := tui.Deps{Leetgo: lc, Store: st, Config: &cfg, Coach: cch}
	// Study plans (builtin + user-defined under <config dir>/studyplans/).
	plans, err := studyplan.All(filepath.Join(cfg.Dir, "studyplans"))
	if err != nil {
		return fmt.Errorf("load study plans: %w", err)
	}
	deps.Plans = studyplan.NewService(st, plans)

	p := tea.NewProgram(tui.New(deps), tea.WithAltScreen(), tea.WithInput(in), tea.WithOutput(out))
	_, err = p.Run()
	return err
}
