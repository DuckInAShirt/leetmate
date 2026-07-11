// Package doctor checks whether the local environment can start LeetMate.
package doctor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
	"gopkg.in/yaml.v3"
)

// Level is the severity of one environment check.
type Level string

const (
	Pass Level = "pass"
	Warn Level = "warn"
	Fail Level = "fail"
)

// Check identifies one local prerequisite and its structured result.
type Check struct {
	ID     string `json:"id"`
	Level  Level  `json:"level"`
	Reason string `json:"reason"`
	Value  string `json:"value,omitempty"`
	Extra  string `json:"extra,omitempty"`
}

// Report is the complete local environment diagnosis.
type Report struct {
	Checks    []Check `json:"checks"`
	Workspace string  `json:"workspace,omitempty"`
	Binary    string  `json:"binary,omitempty"`
}

// HasFailures reports whether LeetMate should stop before starting the TUI.
func (r Report) HasFailures() bool {
	for _, check := range r.Checks {
		if check.Level == Fail {
			return true
		}
	}
	return false
}

// Run performs local-only checks. It never contacts LeetCode or an LLM.
func Run(cfg config.Config, configPath, cwd string) Report {
	report := Report{}
	if _, err := os.Stat(configPath); err == nil {
		report.Checks = append(report.Checks, Check{ID: "config", Level: Pass, Reason: "found", Value: configPath})
	} else if errors.Is(err, os.ErrNotExist) {
		report.Checks = append(report.Checks, Check{ID: "config", Level: Fail, Reason: "missing", Value: configPath})
	} else {
		report.Checks = append(report.Checks, Check{ID: "config", Level: Fail, Reason: "unreadable", Value: configPath, Extra: err.Error()})
	}

	workspace := strings.TrimSpace(cfg.Leetgo.Workspace)
	if workspace == "" {
		workspace, _ = FindWorkspace(cwd)
	} else if absolute, err := filepath.Abs(workspace); err == nil {
		workspace = absolute
	}
	report.Workspace = workspace
	if workspace == "" {
		report.Checks = append(report.Checks, Check{ID: "workspace", Level: Fail, Reason: "missing"})
	} else {
		check := inspectWorkspace(workspace)
		report.Checks = append(report.Checks, check)
		if check.Level == Pass {
			report.Checks = append(report.Checks, inspectCredentials(workspace))
		}
	}

	binary := strings.TrimSpace(cfg.Leetgo.Binary)
	if binary == "" {
		binary = "leetgo"
	}
	if workspace != "" && !filepath.IsAbs(binary) && strings.ContainsAny(binary, `/\\`) {
		binary = filepath.Join(workspace, binary)
	}
	if path, err := exec.LookPath(binary); err == nil {
		report.Binary = path
		report.Checks = append(report.Checks, Check{ID: "leetgo", Level: Pass, Reason: "found", Value: path})
	} else {
		report.Binary = binary
		report.Checks = append(report.Checks, Check{ID: "leetgo", Level: Fail, Reason: "missing", Value: binary})
	}

	if cfg.LLM.Provider != "" && cfg.LLM.Provider != "gemini" && cfg.LLM.Provider != "openai" {
		report.Checks = append(report.Checks, Check{ID: "llm", Level: Fail, Reason: "invalid_provider", Value: cfg.LLM.Provider})
	} else if cfg.APIKey() == "" {
		report.Checks = append(report.Checks, Check{ID: "llm", Level: Warn, Reason: "missing", Value: cfg.LLM.APIKeyEnv})
	} else {
		report.Checks = append(report.Checks, Check{ID: "llm", Level: Pass, Reason: "found", Value: cfg.LLM.APIKeyEnv})
	}
	if err := probeWritable(cfg.Dir); err != nil {
		report.Checks = append(report.Checks, Check{ID: "config_dir", Level: Fail, Reason: "unwritable", Value: cfg.Dir, Extra: err.Error()})
	} else {
		report.Checks = append(report.Checks, Check{ID: "config_dir", Level: Pass, Reason: "writable", Value: cfg.Dir})
	}
	dbDir := cfg.Dir
	if cfg.DB.Path != "" {
		dbDir = filepath.Dir(cfg.DB.Path)
	}
	if err := probeWritable(dbDir); err != nil {
		report.Checks = append(report.Checks, Check{ID: "data", Level: Fail, Reason: "unwritable", Value: dbDir, Extra: err.Error()})
	} else {
		report.Checks = append(report.Checks, Check{ID: "data", Level: Pass, Reason: "writable", Value: dbDir})
	}
	return report
}

// FindWorkspace searches start and its parents for leetgo.yaml.
func FindWorkspace(start string) (string, error) {
	if strings.TrimSpace(start) == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if info, statErr := os.Stat(filepath.Join(dir, "leetgo.yaml")); statErr == nil && !info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func inspectWorkspace(workspace string) Check {
	info, err := os.Stat(workspace)
	if err != nil {
		return Check{ID: "workspace", Level: Fail, Reason: "missing", Value: workspace, Extra: err.Error()}
	}
	if !info.IsDir() {
		return Check{ID: "workspace", Level: Fail, Reason: "not_directory", Value: workspace}
	}
	path := filepath.Join(workspace, "leetgo.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return Check{ID: "workspace", Level: Fail, Reason: "no_config", Value: workspace, Extra: err.Error()}
	}
	var cfg struct {
		Code struct {
			Lang string `yaml:"lang"`
		} `yaml:"code"`
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Check{ID: "workspace", Level: Fail, Reason: "invalid_config", Value: path, Extra: err.Error()}
	}
	if strings.TrimSpace(cfg.Code.Lang) == "" {
		return Check{ID: "workspace", Level: Fail, Reason: "missing_language", Value: path}
	}
	return Check{ID: "workspace", Level: Pass, Reason: "ready", Value: workspace}
}

type workspaceConfig struct {
	LeetCode struct {
		Credentials credentialsConfig `yaml:"credentials"`
	} `yaml:"leetcode"`
}

type credentialsConfig struct {
	From stringList `yaml:"from"`
}

func (c *credentialsConfig) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		var source string
		if err := node.Decode(&source); err != nil {
			return err
		}
		c.From.Values = []string{source}
		return nil
	}
	type plain credentialsConfig
	return node.Decode((*plain)(c))
}

type stringList struct {
	Values   []string
	Sequence bool
}

func (s *stringList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var value string
		if err := node.Decode(&value); err != nil {
			return err
		}
		if value != "" {
			s.Values = []string{value}
		}
		return nil
	case yaml.SequenceNode:
		s.Sequence = true
		return node.Decode(&s.Values)
	default:
		return fmt.Errorf("expected a string or list")
	}
}

func inspectCredentials(workspace string) Check {
	b, err := os.ReadFile(filepath.Join(workspace, "leetgo.yaml"))
	if err != nil {
		return Check{ID: "auth", Level: Warn, Reason: "unreadable", Extra: err.Error()}
	}
	var cfg workspaceConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Check{ID: "auth", Level: Warn, Reason: "unreadable", Extra: err.Error()}
	}
	sourceConfig := cfg.LeetCode.Credentials.From
	sources := sourceConfig.Values
	if len(sources) == 0 {
		return Check{ID: "auth", Level: Warn, Reason: "missing"}
	}
	cookieConfigured := false
	cookieMissing := false
	otherSource := false
	for _, source := range sources {
		switch source {
		case "cookies":
			session, sessionSet := os.LookupEnv("LEETCODE_SESSION")
			csrf, csrfSet := os.LookupEnv("LEETCODE_CSRFTOKEN")
			if !sessionSet || !csrfSet {
				env, envErr := readDotenv(filepath.Join(workspace, ".env"))
				if envErr == nil {
					if !sessionSet {
						session = env["LEETCODE_SESSION"]
					}
					if !csrfSet {
						csrf = env["LEETCODE_CSRFTOKEN"]
					}
				}
			}
			cookieConfigured = credentialConfigured(session) && credentialConfigured(csrf)
			cookieMissing = !cookieConfigured
		default:
			otherSource = true
		}
	}
	if cookieConfigured {
		if sourceConfig.Sequence || len(sources) > 1 {
			return Check{ID: "auth", Level: Warn, Reason: "runtime_unverified", Value: strings.Join(sources, ",")}
		}
		return Check{ID: "auth", Level: Pass, Reason: "cookies_ready", Value: "cookies"}
	}
	if cookieMissing && !otherSource {
		return Check{ID: "auth", Level: Warn, Reason: "cookies_missing", Value: "cookies"}
	}
	return Check{ID: "auth", Level: Warn, Reason: "runtime_unverified", Value: strings.Join(sources, ",")}
}

func credentialConfigured(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !(strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">"))
}

func readDotenv(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	for line := range strings.SplitSeq(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(key), "export "))
		if key == "LEETCODE_SESSION" || key == "LEETCODE_CSRFTOKEN" {
			values[key] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return values, nil
}

func probeWritable(dir string) error {
	if dir == "" {
		return fmt.Errorf("config directory is empty")
	}
	probeDir := dir
	for {
		info, err := os.Stat(probeDir)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", probeDir)
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(probeDir)
		if parent == probeDir {
			return err
		}
		probeDir = parent
	}
	f, err := os.CreateTemp(probeDir, ".leetmate-write-check-*")
	if err != nil {
		return err
	}
	name := f.Name()
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(name)
		return closeErr
	}
	return os.Remove(name)
}
