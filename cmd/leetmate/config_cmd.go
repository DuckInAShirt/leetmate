package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/doctor"
	"github.com/DuckInAShirt/leetmate/internal/tui"
	"gopkg.in/yaml.v3"
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
	if *workspace == "" {
		if cwd, cwdErr := os.Getwd(); cwdErr == nil {
			if found, findErr := doctor.FindWorkspace(cwd); findErr == nil {
				*workspace = found
			}
		}
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
	if err := writeInitFiles(configPath, []byte(configTemplate(*lang, *workspace, *preset)), envPath, []byte(envTemplate(*preset)), *force); err != nil {
		return err
	}

	fmt.Fprintln(out, tui.Textf(*lang, "init.wrote", configPath))
	fmt.Fprintln(out, tui.Textf(*lang, "init.wrote", envPath))
	fmt.Fprintln(out, tui.Text(*lang, "init.next"))
	if *workspace == "" {
		fmt.Fprintln(out, tui.Text(*lang, "init.set_workspace"))
		fmt.Fprintln(out, tui.Text(*lang, "init.run_doctor.second"))
	} else {
		fmt.Fprintln(out, tui.Text(*lang, "init.run_doctor.first"))
		fmt.Fprintln(out, tui.Text(*lang, "init.run_app"))
	}
	fmt.Fprintln(out, tui.Textf(*lang, "init.optional_key", filepath.Join(dir, ".env")))
	return nil
}

func runConfig(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(out)
	showPresets := fs.Bool("presets", false, "list built-in LLM presets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *showPresets {
		if fs.NArg() != 0 {
			return fmt.Errorf("unexpected args with --presets: %s", strings.Join(fs.Args(), " "))
		}
		printPresets(out)
		return nil
	}
	if fs.NArg() == 0 {
		return printConfig(out)
	}

	switch fs.Arg(0) {
	case "set":
		return runConfigSet(fs.Args()[1:], out)
	default:
		return fmt.Errorf("unknown config command %q\nusage: leetmate config [--presets]\n       leetmate config set <key> <value>\nkeys: %s", fs.Arg(0), configSetKeys())
	}
}

func printConfig(out io.Writer) error {
	cfg, path, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "config: %s\n", path)
	fmt.Fprintf(out, "dir:    %s\n", cfg.Dir)
	fmt.Fprintf(out, "lang:   %s\n", cfg.Language)
	fmt.Fprintf(out, "editor: %s\n", editorStatus(cfg))
	fmt.Fprintf(out, "leetgo: %s (%s)\n", emptyAsUnset(cfg.Leetgo.Workspace), cfg.Leetgo.Binary)
	if cfg.Leetgo.Workspace != "" {
		if lang, err := readCodeLang(cfg.Leetgo.Workspace); err == nil && lang != "" {
			fmt.Fprintf(out, "code:   lang=%s\n", lang)
		}
	}
	fmt.Fprintf(out, "llm:    preset=%s provider=%s model=%s\n", emptyAsUnset(cfg.LLM.Preset), cfg.LLM.Provider, cfg.LLM.Model)
	fmt.Fprintf(out, "key:    %s=%s\n", cfg.LLM.APIKeyEnv, keyStatus(cfg.APIKey()))
	fmt.Fprintf(out, "db:     %s\n", cfg.DB.Path)
	return nil
}

func runConfigSet(args []string, out io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: leetmate config set <key> <value>\nkeys: %s", configSetKeys())
	}
	key := strings.TrimSpace(args[0])
	value := strings.TrimSpace(strings.Join(args[1:], " "))
	if key == "code.lang" {
		return setCodeLang(value, out)
	}
	return setConfigFileValue(key, value, out)
}

type configSetSpec struct {
	canonical string
	path      []string
	tag       string
	validate  func(string) error
}

func configSetSpecFor(key string) (configSetSpec, bool) {
	switch key {
	case "language", "lang":
		return configSetSpec{canonical: "language", path: []string{"language"}, tag: "!!str", validate: validateUILang}, true
	case "editor":
		return configSetSpec{canonical: "editor", path: []string{"editor"}, tag: "!!str"}, true
	case "leetgo.workspace", "workspace":
		return configSetSpec{canonical: "leetgo.workspace", path: []string{"leetgo", "workspace"}, tag: "!!str"}, true
	case "leetgo.binary", "binary":
		return configSetSpec{canonical: "leetgo.binary", path: []string{"leetgo", "binary"}, tag: "!!str"}, true
	case "llm.preset", "preset":
		return configSetSpec{canonical: "llm.preset", path: []string{"llm", "preset"}, tag: "!!str", validate: validatePresetName}, true
	case "llm.model", "model":
		return configSetSpec{canonical: "llm.model", path: []string{"llm", "model"}, tag: "!!str"}, true
	case "llm.max_history", "max_history":
		return configSetSpec{canonical: "llm.max_history", path: []string{"llm", "max_history"}, tag: "!!int", validate: validatePositiveInt}, true
	case "db.path", "db":
		return configSetSpec{canonical: "db.path", path: []string{"db", "path"}, tag: "!!str"}, true
	default:
		return configSetSpec{}, false
	}
}

func setConfigFileValue(key, value string, out io.Writer) error {
	spec, ok := configSetSpecFor(key)
	if !ok {
		return fmt.Errorf("unknown config key %q\nkeys: %s", key, configSetKeys())
	}
	if spec.validate != nil {
		if err := spec.validate(value); err != nil {
			return err
		}
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.yaml")
	doc, mode, err := readYAMLDocument(path, []byte(configTemplate("zh", "", "gemini")))
	if err != nil {
		return err
	}
	if err := setYAMLScalar(doc, spec.path, value, spec.tag); err != nil {
		return err
	}
	if err := writeYAMLDocument(path, doc, mode); err != nil {
		return err
	}

	fmt.Fprintf(out, "✓ set %s=%s in %s\n", spec.canonical, value, path)
	if spec.canonical == "llm.preset" {
		if p := presetByNameForCLI(value); p != nil {
			fmt.Fprintf(out, "next: put %s=... in %s\n", p.APIKeyEnv, filepath.Join(dir, ".env"))
		}
	}
	return nil
}

func setCodeLang(value string, out io.Writer) error {
	if value == "" {
		return fmt.Errorf("code.lang cannot be empty")
	}
	cfg, _, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Leetgo.Workspace == "" {
		return fmt.Errorf("leetgo.workspace is unset; run `leetmate config set leetgo.workspace /path/to/workspace` first")
	}
	path := filepath.Join(cfg.Leetgo.Workspace, "leetgo.yaml")
	doc, mode, err := readYAMLDocument(path, nil)
	if err != nil {
		return err
	}
	if err := setYAMLScalar(doc, []string{"code", "lang"}, value, "!!str"); err != nil {
		return err
	}
	if err := writeYAMLDocument(path, doc, mode); err != nil {
		return err
	}
	fmt.Fprintf(out, "✓ set code.lang=%s in %s\n", value, path)
	return nil
}

func validateUILang(value string) error {
	if value != "zh" && value != "en" {
		return fmt.Errorf("invalid language %q (choose: zh or en)", value)
	}
	return nil
}

func validatePresetName(value string) error {
	if !validPreset(value) {
		return fmt.Errorf("unknown preset %q (choose: %s)", value, presetNames())
	}
	return nil
}

func validatePositiveInt(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fmt.Errorf("expected a positive integer, got %q", value)
	}
	return nil
}

func configSetKeys() string {
	return strings.Join([]string{
		"language",
		"editor",
		"leetgo.workspace",
		"leetgo.binary",
		"code.lang",
		"llm.preset",
		"llm.model",
		"llm.max_history",
		"db.path",
	}, ", ")
}

func writeInitFiles(configPath string, configData []byte, envPath string, envData []byte, force bool) error {
	oldEnv, envReadErr := os.ReadFile(envPath)
	envExisted := envReadErr == nil
	if err := os.WriteFile(envPath, envData, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		if force && envExisted {
			_ = os.WriteFile(envPath, oldEnv, 0o600)
		} else if !envExisted {
			_ = os.Remove(envPath)
		}
		return err
	}
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
	workspaceYAML, err := yaml.Marshal(workspace)
	if err != nil {
		workspaceYAML = []byte(`""`)
	}
	workspaceValue := strings.TrimSpace(string(workspaceYAML))
	return strings.Join([]string{
		fmt.Sprintf("language: %s", lang),
		"",
		"leetgo:",
		"  # Directory containing leetgo.yaml. Run \"leetgo init\" first if you do not have one.",
		fmt.Sprintf("  workspace: %s", workspaceValue),
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

func editorStatus(cfg config.Config) string {
	if cfg.Editor != "" {
		return cfg.Editor
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e + " ($EDITOR)"
	}
	return "vi (default)"
}

func readCodeLang(workspace string) (string, error) {
	doc, _, err := readYAMLDocument(filepath.Join(workspace, "leetgo.yaml"), nil)
	if err != nil {
		return "", err
	}
	return getYAMLScalar(doc, []string{"code", "lang"}), nil
}

func readYAMLDocument(path string, fallback []byte) (*yaml.Node, os.FileMode, error) {
	mode := os.FileMode(0o600)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && fallback != nil {
			b = fallback
		} else {
			return nil, 0, err
		}
	} else if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	var doc yaml.Node
	if len(strings.TrimSpace(string(b))) == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{newYAMLMapping()}}
	} else if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil, 0, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		doc.Content = []*yaml.Node{newYAMLMapping()}
	}
	if doc.Content[0].Kind != yaml.MappingNode {
		return nil, 0, fmt.Errorf("%s must contain a YAML mapping", path)
	}
	return &doc, mode, nil
}

func writeYAMLDocument(path string, doc *yaml.Node, mode os.FileMode) error {
	b, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	if mode == 0 {
		mode = 0o600
	}
	return os.WriteFile(path, b, mode)
}

func setYAMLScalar(doc *yaml.Node, path []string, value, tag string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty YAML path")
	}
	m := doc.Content[0]
	for _, key := range path[:len(path)-1] {
		next := yamlMappingValue(m, key)
		if next == nil {
			next = newYAMLMapping()
			appendYAMLMappingPair(m, key, next)
		}
		if next.Kind != yaml.MappingNode {
			return fmt.Errorf("%s must be a YAML mapping", strings.Join(path[:len(path)-1], "."))
		}
		m = next
	}
	last := path[len(path)-1]
	val := yamlMappingValue(m, last)
	if val == nil {
		val = &yaml.Node{Kind: yaml.ScalarNode}
		appendYAMLMappingPair(m, last, val)
	}
	val.Kind = yaml.ScalarNode
	val.Tag = tag
	val.Value = value
	return nil
}

func getYAMLScalar(doc *yaml.Node, path []string) string {
	if len(path) == 0 || len(doc.Content) == 0 {
		return ""
	}
	n := doc.Content[0]
	for _, key := range path {
		if n.Kind != yaml.MappingNode {
			return ""
		}
		n = yamlMappingValue(n, key)
		if n == nil {
			return ""
		}
	}
	if n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}

func yamlMappingValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func appendYAMLMappingPair(m *yaml.Node, key string, value *yaml.Node) {
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func newYAMLMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}
