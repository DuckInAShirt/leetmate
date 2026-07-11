package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFirstRunNonInteractiveNeverPrompts(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workspace := filepath.Join(root, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte("code:\n  lang: go\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEETMATE_CONFIG_DIR", configDir)

	var out bytes.Buffer
	handled, err := runFirstRun(strings.NewReader("ignored\n"), &out, workspace, false)
	if !handled || !errors.Is(err, errDoctorFailed) {
		t.Fatalf("runFirstRun() = handled %v, err %v", handled, err)
	}
	if !strings.Contains(out.String(), "leetmate init --workspace") {
		t.Fatalf("missing actionable command:\n%s", out.String())
	}
	if _, statErr := os.Stat(configDir); !os.IsNotExist(statErr) {
		t.Fatalf("non-interactive first run created config dir: %v", statErr)
	}
}

func TestRunFirstRunInteractiveDiscoversWorkspace(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workspace := filepath.Join(root, "workspace")
	leetgo := filepath.Join(root, "leetgo")
	if err := os.MkdirAll(filepath.Join(workspace, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	leetgoYAML := "code:\n  lang: go\nleetcode:\n  credentials:\n    from: none\n"
	if err := os.WriteFile(filepath.Join(workspace, "leetgo.yaml"), []byte(leetgoYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(leetgo, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEETMATE_CONFIG_DIR", configDir)
	t.Setenv("PATH", root)

	input := strings.NewReader("en\n\n")
	var out bytes.Buffer
	handled, err := runFirstRun(input, &out, filepath.Join(workspace, "nested"), true)
	if err != nil {
		t.Fatalf("runFirstRun: %v\n%s", err, out.String())
	}
	if !handled {
		t.Fatal("runFirstRun should handle missing config")
	}
	configBytes, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	configText := string(configBytes)
	for _, want := range []string{"language: en", "workspace: " + workspace, "preset: gemini"} {
		if !strings.Contains(configText, want) {
			t.Errorf("config missing %q:\n%s", want, configText)
		}
	}
	if !strings.Contains(out.String(), "Found leetgo workspace") || !strings.Contains(out.String(), "Environment check") {
		t.Fatalf("missing onboarding progress:\n%s", out.String())
	}
}

func TestRunFirstRunLeavesExistingConfigUntouched(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "config.yaml")
	original := "not: [valid"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEETMATE_CONFIG_DIR", configDir)

	var out bytes.Buffer
	handled, err := runFirstRun(strings.NewReader(""), &out, t.TempDir(), true)
	if err != nil || handled {
		t.Fatalf("runFirstRun() = handled %v, err %v", handled, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("existing config changed: %q", got)
	}
}
