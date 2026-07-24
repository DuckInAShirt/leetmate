// Package leetgo wraps the external `leetgo` CLI. LeetMate never imports
// leetgo's internal packages (they are not a stable API) — instead it shells
// out to the binary, parsing structured output where possible.
//
// All commands run inside the configured leetgo workspace (the directory that
// holds leetgo.yaml), so the user must have run `leetgo init` there.
package leetgo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/domain"
)

// defaultTimeout caps individual leetgo invocations. Submit/test may hit the
// network; pick is mostly local.
const defaultTimeout = 60 * time.Second

// Client is the leetgo adapter. The zero value is not usable; use New.
type Client struct {
	cfg       config.LeetgoConfig
	workspace string
	lang      string // code language, read from leetgo.yaml (e.g. "go")
}

// New builds a Client. If Workspace is empty it falls back to the current
// working directory. Missing leetgo.yaml is not fatal here (so the caller can
// surface a friendly error), but lang defaults to "go".
func New(cfg config.LeetgoConfig) (*Client, error) {
	ws := cfg.Workspace
	if ws == "" {
		var err error
		ws, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	c := &Client{cfg: cfg, workspace: ws, lang: "go"}
	if lang, err := readLang(ws); err == nil && lang != "" {
		c.lang = lang
	}
	return c, nil
}

// Workspace returns the resolved leetgo workspace directory.
func (c *Client) Workspace() string { return c.workspace }

// Lang returns the code language from leetgo.yaml.
func (c *Client) Lang() string { return c.lang }

// CheckAvailable verifies the leetgo binary exists and that the workspace looks
// initialized. It returns a human-readable message when something is missing.
func (c *Client) CheckAvailable() error {
	if _, err := exec.LookPath(c.cfg.Binary); err != nil {
		return fmt.Errorf("leetgo binary %q not found in PATH — install it via `brew install leetgo` or `go install github.com/j178/leetgo@latest`", c.cfg.Binary)
	}
	if _, err := os.Stat(filepath.Join(c.workspace, "leetgo.yaml")); err != nil {
		return fmt.Errorf("no leetgo.yaml in workspace %q — run `leetgo init` there first", c.workspace)
	}
	return nil
}

// run executes a leetgo command inside the workspace and returns trimmed stdout.
// A "yes\n" stream is attached to stdin so leetgo's interactive overwrite /
// submit confirmations (which read EOF and abort when stdout is captured) are
// auto-accepted. leetgo's stderr (rich progress output) is captured rather than
// streamed, because writing it to the terminal while bubbletea owns the alt
// screen corrupts the layout; it is surfaced only on error.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.cfg.Binary, args...)
	cmd.Dir = c.workspace
	cmd.Stdin = strings.NewReader(strings.Repeat("y\n", 200))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		tail := strings.TrimSpace(stderr.String())
		if tail != "" {
			return out, fmt.Errorf("leetgo %s: %w\n%s", strings.Join(args, " "), err, tail)
		}
		return out, fmt.Errorf("leetgo %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// Info fetches problem metadata via `leetgo info --format json`.
func (c *Client) Info(ctx context.Context, qid string) (domain.ProblemMeta, error) {
	out, err := c.run(ctx, "info", qid, "--format", "json")
	if err != nil {
		return domain.ProblemMeta{}, err
	}
	return parseMeta(out)
}

// Pick generates the problem skeleton via `leetgo pick`, then loads the
// statement and locates the code file on disk.
func (c *Client) Pick(ctx context.Context, qid string) (domain.Problem, error) {
	if _, err := c.run(ctx, "pick", qid); err != nil {
		return domain.Problem{}, err
	}
	meta, err := c.Info(ctx, qid)
	if err != nil {
		return domain.Problem{}, err
	}
	problem := domain.Problem{ProblemMeta: meta}
	if dir, err := c.resolveProblemDir(meta.Slug); err == nil {
		problem.CodePath = c.codeFile(dir)
		problem.Content = readStatement(dir, meta.Slug)
	}
	return problem, nil
}

// ReadCode returns the current contents of the problem's code file.
func (c *Client) ReadCode(slug string) (string, error) {
	dir, err := c.resolveProblemDir(slug)
	if err != nil {
		return "", err
	}
	path := c.codeFile(dir)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// runFull is like run but also returns stderr, for commands whose useful
// detail (e.g. failing test cases) is printed there even on exit 0/1.
func (c *Client) runFull(ctx context.Context, args ...string) (stdout, stderr []byte, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.cfg.Binary, args...)
	cmd.Dir = c.workspace
	cmd.Stdin = strings.NewReader(strings.Repeat("y\n", 200))
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func combineOutput(stdout, stderr []byte) string {
	out := strings.TrimRight(string(stdout), "\n")
	err := strings.TrimRight(string(stderr), "\n")
	switch {
	case out == "":
		return err
	case err == "":
		return out
	default:
		return out + "\n" + err
	}
}

// Test runs `leetgo test <qid>` and parses the outcome. The full output is kept
// in TestResult.Raw so the UI can expand remote judge details such as Input /
// Output / Expected when a case fails.
func (c *Client) Test(ctx context.Context, qid string) (domain.TestResult, error) {
	stdout, stderr, err := c.runFull(ctx, "test", qid)
	combined := combineOutput(stdout, stderr)
	res := parseTestOutput(combined)
	res.Raw = combined
	if err != nil {
		return res, fmt.Errorf("leetgo test %s: %w\n%s", qid, err, strings.TrimSpace(combined))
	}
	return res, nil
}

// Submit runs `leetgo submit <qid>` and parses the verdict.
func (c *Client) Submit(ctx context.Context, qid string) (domain.SubmitResult, error) {
	out, err := c.run(ctx, "submit", qid)
	if err != nil {
		// leetgo may exit non-zero on Wrong Answer but still print a result;
		// fall through to parse whatever we got.
		if len(out) == 0 {
			return domain.SubmitResult{}, err
		}
	}
	return parseSubmitOutput(string(out)), nil
}

// codeFile locates the learner's main code file inside the problem directory.
// leetgo's default Go template emits `solution.go` (not `<dirname>.go`), so we
// scan the directory: prefer `solution<ext>`, otherwise the first non-test
// source file matching the language extension.
//
// When the configured language has no file in this directory (e.g. the learner
// switched code.lang but this problem was only ever generated for another
// language), fall back to any recognized source file so the practice view never
// comes up empty.
func (c *Client) codeFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	if f := findCodeFile(entries, dir, langExt(c.lang), c.lang); f != "" {
		return f
	}
	// Fall back across other languages — the configured one had no match here.
	for _, lc := range knownLangExts {
		if lc.ext == langExt(c.lang) {
			continue
		}
		if f := findCodeFile(entries, dir, lc.ext, lc.lang); f != "" {
			return f
		}
	}
	return ""
}

// findCodeFile returns the path to the learner's code file among entries for the
// given extension: `solution<ext>` wins, otherwise the first non-test source.
func findCodeFile(entries []os.DirEntry, dir, ext, lang string) string {
	var fallback string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ext) {
			continue
		}
		if isTestFile(name, lang) {
			continue
		}
		if name == "solution"+ext {
			return filepath.Join(dir, name)
		}
		if fallback == "" {
			fallback = filepath.Join(dir, name)
		}
	}
	return fallback
}

// resolveProblemDir finds the generated directory for a slug under the
// workspace (handles the <NNNN>.<slug> naming convention across language
// subdirectories). When the workspace contains several language subdirs for
// the same problem (e.g. both go/0076.<slug> and python/0076.<slug>), prefer
// the one under the configured code language — otherwise codeFile cannot find
// the source file and the practice view comes up empty.
func (c *Client) resolveProblemDir(slug string) (string, error) {
	var match, langMatch string
	err := filepath.Walk(c.workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		// Directory basename is "<id>.<slug>"; match by suffix ".<slug>".
		if !strings.HasSuffix(info.Name(), "."+slug) {
			return nil
		}
		if match == "" {
			match = path
		}
		// Prefer the directory nested under the configured language subdir
		// (e.g. "python/0076.<slug>" when lang is "python").
		if langMatch == "" {
			if rel, rerr := filepath.Rel(c.workspace, path); rerr == nil {
				if top, _, _ := strings.Cut(filepath.ToSlash(rel), "/"); top == c.lang {
					langMatch = path
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	chosen := langMatch
	if chosen == "" {
		chosen = match
	}
	if chosen == "" {
		return "", fmt.Errorf("no generated directory found for slug %q under %s", slug, c.workspace)
	}
	return chosen, nil
}
