package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoctorReportsReadyWithWarnings(t *testing.T) {
	configDir, workspace, leetgo := doctorFixture(t, "en", "none")
	t.Setenv("LEETMATE_CONFIG_DIR", configDir)
	t.Setenv("LEETMATE_LEETGO_BINARY", leetgo)
	t.Setenv("LEETMATE_LEETGO_WORKSPACE", workspace)

	var out bytes.Buffer
	if err := runDoctor(nil, &out); err != nil {
		t.Fatalf("runDoctor: %v\n%s", err, out.String())
	}
	for _, want := range []string{"LeetMate environment check", "[PASS] config", "[WARN] auth", "Coach is unavailable"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("doctor output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunDoctorReportsInvalidConfigAsStructuredFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("not: [valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEETMATE_CONFIG_DIR", dir)

	var out bytes.Buffer
	err := runDoctor(nil, &out)
	if !errors.Is(err, errDoctorFailed) {
		t.Fatalf("runDoctor error = %v", err)
	}
	if !strings.Contains(out.String(), "[FAIL] 配置") || !strings.Contains(out.String(), "配置不可读") {
		t.Fatalf("invalid config not diagnosed:\n%s", out.String())
	}
}

func TestRunDoctorJSONIsStructuredAndReturnsFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LEETMATE_CONFIG_DIR", filepath.Join(dir, "missing"))
	t.Setenv("LEETMATE_LEETGO_BINARY", filepath.Join(dir, "missing-leetgo"))

	var out bytes.Buffer
	err := runDoctor([]string{"--json"}, &out)
	if !errors.Is(err, errDoctorFailed) {
		t.Fatalf("runDoctor error = %v, want errDoctorFailed", err)
	}
	var report struct {
		Checks []struct {
			ID    string `json:"id"`
			Level string `json:"level"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(report.Checks) == 0 || report.Checks[0].ID != "config" || report.Checks[0].Level != "fail" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func doctorFixture(t *testing.T, lang, auth string) (configDir, workspace, leetgo string) {
	t.Helper()
	root := t.TempDir()
	configDir = filepath.Join(root, "config")
	workspace = filepath.Join(root, "workspace")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	configYAML := "language: " + lang + "\nleetgo:\n  workspace: " + workspace + "\n  binary: " + filepath.Join(root, "leetgo") + "\nllm:\n  preset: gemini\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte("code:\n  lang: go\nleetcode:\n  credentials:\n    from: "+auth+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	leetgo = filepath.Join(root, "leetgo")
	if err := os.WriteFile(leetgo, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return configDir, workspace, leetgo
}
