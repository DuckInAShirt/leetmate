package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

func runInit(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(out)
	preset := fs.String("preset", "siliconflow", "LLM preset: gemini, siliconflow, groq, deepseek")
	workspace := fs.String("workspace", "", "leetgo workspace containing leetgo.yaml")
	lang := fs.String("lang", "zh", "UI language: zh or en")
	force := fs.Bool("force", false, "overwrite existing config.yaml and .env template")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected args: %s", strings.Join(fs.Args(), " "))
	}
	if !validPreset(*preset) {
		return fmt.Errorf("unknown preset %q (choose: %s)", *preset, presetNames())
	}
	if *lang != "zh" && *lang != "en" {
		return fmt.Errorf("invalid lang %q (choose: zh or en)", *lang)
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	configPath := filepath.Join(dir, "config.yaml")
	envPath := filepath.Join(dir, ".env")
	if err := ensureWritable(configPath, *force); err != nil {
		return err
	}
	if err := ensureWritable(envPath, *force); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, []byte(configTemplate(*lang, *workspace, *preset)), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(envPath, []byte(envTemplate(*preset)), 0o600); err != nil {
		return err
	}

	fmt.Fprintf(out, "✓ wrote %s\n", configPath)
	fmt.Fprintf(out, "✓ wrote %s\n", envPath)
	fmt.Fprintln(out, "next:")
	fmt.Fprintf(out, "  1. edit %s and fill your API key\n", filepath.Join(dir, ".env"))
	if *workspace == "" {
		fmt.Fprintf(out, "  2. set leetgo.workspace in %s\n", filepath.Join(dir, "config.yaml"))
	} else {
		fmt.Fprintln(out, "  2. run leetmate")
	}
	return nil
}

func runConfig(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(out)
	showPresets := fs.Bool("presets", false, "list built-in LLM presets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected args: %s", strings.Join(fs.Args(), " "))
	}
	if *showPresets {
		printPresets(out)
		return nil
	}

	cfg, path, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "config: %s\n", path)
	fmt.Fprintf(out, "dir:    %s\n", cfg.Dir)
	fmt.Fprintf(out, "lang:   %s\n", cfg.Language)
	fmt.Fprintf(out, "leetgo: %s (%s)\n", emptyAsUnset(cfg.Leetgo.Workspace), cfg.Leetgo.Binary)
	fmt.Fprintf(out, "llm:    preset=%s provider=%s model=%s\n", emptyAsUnset(cfg.LLM.Preset), cfg.LLM.Provider, cfg.LLM.Model)
	fmt.Fprintf(out, "key:    %s=%s\n", cfg.LLM.APIKeyEnv, keyStatus(cfg.APIKey()))
	fmt.Fprintf(out, "db:     %s\n", cfg.DB.Path)
	return nil
}

func ensureWritable(path string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func configTemplate(lang, workspace, preset string) string {
	return strings.Join([]string{
		fmt.Sprintf("language: %s", lang),
		"",
		"leetgo:",
		"  # Directory containing leetgo.yaml. Run \"leetgo init\" first if you do not have one.",
		fmt.Sprintf("  workspace: %s", workspace),
		"  binary: leetgo",
		"",
		"llm:",
		fmt.Sprintf("  # Choose: %s", presetNames()),
		fmt.Sprintf("  preset: %s", preset),
		"  max_history: 12",
		"",
		"# Optional: leave empty to use <config dir>/leetmate.db",
		"db:",
		"  path: \"\"",
		"",
		"# Optional: leave empty to use $EDITOR, then vi",
		"editor: \"\"",
		"",
	}, "\n")
}

func envTemplate(preset string) string {
	key := "GEMINI_API_KEY"
	if p := presetByNameForCLI(preset); p != nil {
		key = p.APIKeyEnv
	}
	return strings.Join([]string{
		"# Put secrets here. This file is loaded before config.yaml is resolved.",
		key + "=",
		"",
	}, "\n")
}

func printPresets(out io.Writer) {
	for _, p := range config.Presets {
		free := "paid"
		if p.Free {
			free = "free"
		}
		fmt.Fprintf(out, "%-12s %-7s %s\n  model: %s\n  key:   %s\n  url:   %s\n", p.Name, free, p.Desc, p.Model, p.APIKeyEnv, p.SignupURL)
	}
}

func presetNames() string {
	names := make([]string, 0, len(config.Presets))
	for _, p := range config.Presets {
		names = append(names, p.Name)
	}
	return strings.Join(names, " | ")
}

func validPreset(name string) bool {
	return presetByNameForCLI(name) != nil
}

func presetByNameForCLI(name string) *config.Preset {
	for i := range config.Presets {
		if config.Presets[i].Name == name {
			return &config.Presets[i]
		}
	}
	return nil
}

func emptyAsUnset(s string) string {
	if s == "" {
		return "<unset>"
	}
	return s
}

func keyStatus(key string) string {
	if key == "" {
		return "missing"
	}
	return "set"
}
