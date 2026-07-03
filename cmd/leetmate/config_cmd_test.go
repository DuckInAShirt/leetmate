package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitWritesConfigAndEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)

	var out bytes.Buffer
	if err := runInit([]string{"--preset", "gemini", "--workspace", "/tmp/leetgo", "--lang", "en"}, &out); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	cfg, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	for _, want := range []string{"language: en", "workspace: /tmp/leetgo", "preset: gemini"} {
		if !strings.Contains(string(cfg), want) {
			t.Errorf("config.yaml missing %q:\n%s", want, cfg)
		}
	}

	env, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	if !strings.Contains(string(env), "GEMINI_API_KEY=") {
		t.Errorf(".env missing GEMINI_API_KEY:\n%s", env)
	}
	if !strings.Contains(out.String(), "next:") {
		t.Errorf("output missing next steps:\n%s", out.String())
	}
}

func TestRunInitRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	if err := runInit(nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("first runInit: %v", err)
	}
	if err := runInit(nil, &bytes.Buffer{}); err == nil {
		t.Fatal("second runInit should refuse overwriting existing files")
	}
	if err := runInit([]string{"--force"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit --force: %v", err)
	}
}

func TestRunConfigShowsResolvedStatus(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	t.Setenv("GEMINI_API_KEY", "test-key")
	if err := runInit([]string{"--preset", "gemini", "--workspace", "/tmp/leetgo"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var out bytes.Buffer
	if err := runConfig(nil, &out); err != nil {
		t.Fatalf("runConfig: %v", err)
	}
	for _, want := range []string{"config:", "editor:", "leetgo: /tmp/leetgo", "preset=gemini", "GEMINI_API_KEY=set"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("config output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunConfigListsPresets(t *testing.T) {
	var out bytes.Buffer
	if err := runConfig([]string{"--presets"}, &out); err != nil {
		t.Fatalf("runConfig --presets: %v", err)
	}
	for _, want := range []string{"gemini", "siliconflow", "GEMINI_API_KEY", "SILICONFLOW_API_KEY"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("presets output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunConfigSetWritesConfigValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	if err := runInit([]string{"--preset", "gemini", "--lang", "zh"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var out bytes.Buffer
	if err := runConfig([]string{"set", "language", "en"}, &out); err != nil {
		t.Fatalf("config set language: %v", err)
	}
	if !strings.Contains(out.String(), "set language=en") {
		t.Errorf("set output missing confirmation:\n%s", out.String())
	}
	cfg, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	if !strings.Contains(string(cfg), "language: en") {
		t.Fatalf("language not updated:\n%s", cfg)
	}
}

func TestRunConfigSetPresetPrintsKeyHint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	if err := runInit([]string{"--preset", "gemini"}, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var out bytes.Buffer
	if err := runConfig([]string{"set", "llm.preset", "siliconflow"}, &out); err != nil {
		t.Fatalf("config set llm.preset: %v", err)
	}
	for _, want := range []string{"set llm.preset=siliconflow", "SILICONFLOW_API_KEY"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("set output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunConfigSetCodeLangWritesLeetgoYAML(t *testing.T) {
	dir := t.TempDir()
	workspace := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte("leetcode:\n  site: cn\ncode:\n  lang: go\n"), 0o600); err != nil {
		t.Fatalf("write leetgo.yaml: %v", err)
	}
	if err := runInit([]string{"--workspace", workspace}, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	var out bytes.Buffer
	if err := runConfig([]string{"set", "code.lang", "python3"}, &out); err != nil {
		t.Fatalf("config set code.lang: %v", err)
	}
	if !strings.Contains(out.String(), "set code.lang=python3") {
		t.Errorf("set output missing confirmation:\n%s", out.String())
	}
	leetgoYAML, err := os.ReadFile(filepath.Join(workspace, "leetgo.yaml"))
	if err != nil {
		t.Fatalf("read leetgo.yaml: %v", err)
	}
	if !strings.Contains(string(leetgoYAML), "lang: python3") {
		t.Fatalf("code.lang not updated:\n%s", leetgoYAML)
	}

	out.Reset()
	if err := runConfig(nil, &out); err != nil {
		t.Fatalf("runConfig: %v", err)
	}
	if !strings.Contains(out.String(), "code:   lang=python3") {
		t.Errorf("config output missing code lang:\n%s", out.String())
	}
}

func TestRunConfigSetRejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", dir)
	if err := runInit(nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	for _, args := range [][]string{
		{"set", "language", "fr"},
		{"set", "llm.preset", "unknown"},
		{"set", "llm.max_history", "0"},
		{"set", "nope", "value"},
	} {
		if err := runConfig(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("runConfig(%v) should fail", args)
		}
	}
}
