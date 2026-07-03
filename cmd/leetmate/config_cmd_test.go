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
	for _, want := range []string{"config:", "leetgo: /tmp/leetgo", "preset=gemini", "GEMINI_API_KEY=set"} {
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
