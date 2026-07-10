package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

func TestFindWorkspaceSearchesParents(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\n")
	nested := filepath.Join(workspace, "go", "two-sum")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindWorkspace(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != workspace {
		t.Fatalf("FindWorkspace() = %q, want %q", got, workspace)
	}
}

func TestRunResolvesRelativeBinaryFromWorkspace(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(filepath.Join(workspace, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials: none\n")
	configPath := filepath.Join(dir, "config.yaml")
	writeFile(t, configPath, "language: en\n")
	binary := filepath.Join(workspace, "bin", "leetgo")
	writeFile(t, binary, "#!/bin/sh\n")
	if err := os.Chmod(binary, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Dir = dir
	cfg.Leetgo.Workspace = workspace
	cfg.Leetgo.Binary = "./bin/leetgo"

	report := Run(cfg, configPath, dir)
	if report.Binary != binary {
		t.Fatalf("Binary = %q, want %q", report.Binary, binary)
	}
	assertCheck(t, report, "leetgo", Pass, "found")
}

func TestRunTreatsOptionalServicesAsWarnings(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials:\n    from: none\n")
	configPath := filepath.Join(dir, "config.yaml")
	writeFile(t, configPath, "language: en\n")
	leetgo := filepath.Join(dir, "leetgo")
	writeFile(t, leetgo, "#!/bin/sh\n")
	if err := os.Chmod(leetgo, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Dir = dir
	cfg.Leetgo.Binary = leetgo
	cfg.Leetgo.Workspace = workspace
	report := Run(cfg, configPath, workspace)
	if report.HasFailures() {
		t.Fatalf("Run() unexpectedly failed: %#v", report.Checks)
	}
	assertCheck(t, report, "auth", Warn, "runtime_unverified")
	assertCheck(t, report, "llm", Warn, "missing")
}

func TestInspectCredentialsSupportsScalarAndSequence(t *testing.T) {
	for name, from := range map[string]string{
		"scalar":   "cookies",
		"sequence": "[cookies]",
	} {
		t.Run(name, func(t *testing.T) {
			workspace := t.TempDir()
			writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "leetcode:\n  credentials:\n    from: "+from+"\n")
			writeFile(t, filepath.Join(workspace, ".env"), "LEETCODE_SESSION=secret-session\nLEETCODE_CSRFTOKEN=secret-csrf\n")

			check := inspectCredentials(workspace)
			wantLevel, wantReason := Pass, "cookies_ready"
			if name == "sequence" {
				wantLevel, wantReason = Warn, "runtime_unverified"
			}
			if check.Level != wantLevel || check.Reason != wantReason {
				t.Fatalf("inspectCredentials() = %#v", check)
			}
			if strings.Contains(check.Value+check.Extra, "secret-") {
				t.Fatal("credential value leaked into check result")
			}
		})
	}
}

func TestInspectCredentialsSupportsLegacyCredentialsScalar(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "leetcode:\n  credentials: cookies\n")
	writeFile(t, filepath.Join(workspace, ".env"), "LEETCODE_SESSION=secret-session\nLEETCODE_CSRFTOKEN=secret-csrf\n")

	check := inspectCredentials(workspace)
	if check.Level != Pass || check.Reason != "cookies_ready" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsSupportsProcessEnvironment(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "secret-session")
	t.Setenv("LEETCODE_CSRFTOKEN", "secret-csrf")
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "leetcode:\n  credentials:\n    from: cookies\n")

	check := inspectCredentials(workspace)
	if check.Level != Pass || check.Reason != "cookies_ready" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsHonorsEmptyProcessEnvironment(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRFTOKEN", "")
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials:\n    from: cookies\n")
	writeFile(t, filepath.Join(workspace, ".env"), "LEETCODE_SESSION=valid-session\nLEETCODE_CSRFTOKEN=valid-csrf\n")

	check := inspectCredentials(workspace)
	if check.Level != Warn || check.Reason != "cookies_missing" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsAllowsFallbackProvider(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRFTOKEN", "")
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials:\n    from: [cookies, browser]\n")

	check := inspectCredentials(workspace)
	if check.Level != Warn || check.Reason != "runtime_unverified" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsSupportsExportDotenv(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials: cookies\n")
	writeFile(t, filepath.Join(workspace, ".env"), "export LEETCODE_SESSION=valid-session\nexport LEETCODE_CSRFTOKEN=valid-csrf\n")

	check := inspectCredentials(workspace)
	if check.Level != Pass || check.Reason != "cookies_ready" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsRejectsPlaceholderValues(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials:\n    from: cookies\n")
	writeFile(t, filepath.Join(workspace, ".env"), "LEETCODE_SESSION=<LEETCODE_SESSION cookie>\nLEETCODE_CSRFTOKEN=<csrftoken cookie>\n")

	check := inspectCredentials(workspace)
	if check.Level != Warn || check.Reason != "cookies_missing" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestInspectCredentialsWarnsWhenCookieValuesAreMissing(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRFTOKEN", "")
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "leetcode:\n  credentials:\n    from: cookies\n")

	check := inspectCredentials(workspace)
	if check.Level != Warn || check.Reason != "cookies_missing" {
		t.Fatalf("inspectCredentials() = %#v", check)
	}
}

func TestRunFailsForUnknownLLMProvider(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "code:\n  lang: go\nleetcode:\n  credentials: none\n")
	configPath := filepath.Join(dir, "config.yaml")
	writeFile(t, configPath, "language: en\n")
	leetgo := filepath.Join(dir, "leetgo")
	writeFile(t, leetgo, "#!/bin/sh\n")
	if err := os.Chmod(leetgo, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Dir = dir
	cfg.Leetgo.Binary = leetgo
	cfg.Leetgo.Workspace = workspace
	cfg.LLM.Provider = "typo"

	report := Run(cfg, configPath, workspace)
	assertCheck(t, report, "llm", Fail, "invalid_provider")
}

func TestRunFailsWhenWorkspaceHasNoCodeLanguage(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(workspace, "leetgo.yaml"), "leetcode:\n  credentials: none\n")
	configPath := filepath.Join(dir, "config.yaml")
	writeFile(t, configPath, "language: en\n")
	leetgo := filepath.Join(dir, "leetgo")
	writeFile(t, leetgo, "#!/bin/sh\n")
	if err := os.Chmod(leetgo, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Dir = dir
	cfg.Leetgo.Binary = leetgo
	cfg.Leetgo.Workspace = workspace

	report := Run(cfg, configPath, workspace)
	assertCheck(t, report, "workspace", Fail, "missing_language")
}

func TestRunFailsWhenRequiredSetupIsMissing(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "missing-config-dir")
	cfg := config.Default()
	cfg.Dir = configDir
	cfg.Leetgo.Binary = filepath.Join(dir, "missing-leetgo")

	report := Run(cfg, filepath.Join(configDir, "config.yaml"), dir)
	if !report.HasFailures() {
		t.Fatalf("Run() should fail: %#v", report.Checks)
	}
	assertCheck(t, report, "config", Fail, "missing")
	assertCheck(t, report, "leetgo", Fail, "missing")
	assertCheck(t, report, "workspace", Fail, "missing")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("doctor created config dir: %v", err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertCheck(t *testing.T, report Report, id string, level Level, reason string) {
	t.Helper()
	for _, check := range report.Checks {
		if check.ID == id {
			if check.Level != level || check.Reason != reason {
				t.Fatalf("check %q = %#v, want level=%q reason=%q", id, check, level, reason)
			}
			return
		}
	}
	t.Fatalf("check %q not found in %#v", id, report.Checks)
}
