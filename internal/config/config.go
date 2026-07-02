// Package config loads LeetMate's configuration. Resolution order, later wins:
//
//  1. Built-in defaults
//  2. ~/.config/leetmate/config.yaml
//  3. Environment variables (LEETMATE_*, plus the LLM key referenced by APIKeyEnv)
//
// A .env file next to config.yaml is loaded first to populate the environment,
// so users can keep GEMINI_API_KEY etc. out of the YAML.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config is the resolved, in-memory configuration handed to the app.
type Config struct {
	LLM      LLMConfig    `yaml:"llm"`
	Leetgo   LeetgoConfig `yaml:"leetgo"`
	DB       DBConfig     `yaml:"db"`
	Editor   string       `yaml:"editor"`
	Language string       `yaml:"language"` // "zh" (default) or "en"
	// Dir is the LeetMate config directory (~/.config/leetmate), set at load time.
	Dir string `yaml:"-"`
}

type LLMConfig struct {
	// Preset selects a built-in provider profile (gemini/siliconflow/groq/deepseek).
	// When set, provider/base_url/api_key_env are filled from the preset and the
	// user only needs to provide the matching key in .env. Model can still be
	// overridden below.
	Preset string `yaml:"preset"`
	// Provider is "gemini" or "openai" (any OpenAI-compatible endpoint).
	Provider string `yaml:"provider"`
	// BaseURL is the OpenAI-compatible base URL. Ignored for gemini.
	BaseURL string `yaml:"base_url"`
	// Model is the model id, e.g. "gemini-2.0-flash" or "Qwen/Qwen2.5-Coder-32B-Instruct".
	Model string `yaml:"model"`
	// APIKeyEnv names the environment variable holding the API key (e.g. GEMINI_API_KEY).
	APIKeyEnv string `yaml:"api_key_env"`
	// MaxHistory is the number of past coaching turns to include as context.
	MaxHistory int `yaml:"max_history"`
}

type LeetgoConfig struct {
	// Workspace is the path to the directory containing leetgo.yaml.
	Workspace string `yaml:"workspace"`
	// Binary is the leetgo executable name/path (default "leetgo").
	Binary string `yaml:"binary"`
}

type DBConfig struct {
	// Path to the SQLite file. Empty uses <config dir>/leetmate.db.
	Path string `yaml:"path"`
}

// Default returns a config populated with sensible defaults.
func Default() Config {
	return Config{
		Language: "zh",
		LLM: LLMConfig{
			Preset:    "gemini", // global default; mainland-China users switch to "siliconflow"
			MaxHistory: 12,
		},
		Leetgo: LeetgoConfig{
			Binary: "leetgo",
		},
		Editor: "",
	}
}

// ConfigDir returns the LeetMate config directory: $LEETMATE_CONFIG_DIR,
// $XDG_CONFIG_HOME/leetmate, or ~/.config/leetmate (per-platform).
func ConfigDir() (string, error) {
	if v := os.Getenv("LEETMATE_CONFIG_DIR"); v != "" {
		return v, nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "leetmate"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		// Prefer XDG-style; macOS users with XDG_CONFIG_HOME are respected above.
		return filepath.Join(home, ".config", "leetmate"), nil
	}
	return filepath.Join(home, ".config", "leetmate"), nil
}

// Load reads config from dir, applying defaults first. Missing config file is
// not an error — defaults are returned so the app can prompt the user to
// configure. Returns the resolved config and the path it tried to read.
func Load() (Config, string, error) {
	cfg := Default()

	dir, err := ConfigDir()
	if err != nil {
		return cfg, "", err
	}
	cfg.Dir = dir

	// Populate env from ~/.config/leetmate/.env (if present) so the LLM key and
	// leetgo credentials can live there instead of in shell rc files. Real
	// environment variables always take precedence.
	loadDotenv(filepath.Join(dir, ".env"))

	path := filepath.Join(dir, "config.yaml")
	b, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return cfg, path, fmt.Errorf("parse %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return cfg, path, err
	}

	cfg.applyDefaults()
	return cfg, path, nil
}

func (c *Config) applyDefaults() {
	if c.Language == "" {
		c.Language = "zh"
	}
	// Apply an LLM preset if set: fills provider/base_url/api_key_env, and model
	// unless the user overrode it. Preset takes precedence over bare fields.
	if p := presetByName(c.LLM.Preset); p != nil {
		c.LLM.Provider = p.Provider
		c.LLM.BaseURL = p.BaseURL
		c.LLM.APIKeyEnv = p.APIKeyEnv
		if c.LLM.Model == "" {
			c.LLM.Model = p.Model
		}
	}
	if c.LLM.Provider == "" {
		c.LLM.Provider = "gemini"
	}
	if c.LLM.Model == "" && c.LLM.Provider == "gemini" {
		c.LLM.Model = "gemini-2.0-flash"
	}
	if c.LLM.APIKeyEnv == "" {
		c.LLM.APIKeyEnv = "GEMINI_API_KEY"
	}
	if c.LLM.MaxHistory == 0 {
		c.LLM.MaxHistory = 12
	}
	if c.Leetgo.Binary == "" {
		c.Leetgo.Binary = "leetgo"
	}
	if c.DB.Path == "" {
		c.DB.Path = filepath.Join(c.Dir, "leetmate.db")
	}
}

// APIKey reads the resolved API key from the environment.
func (c *Config) APIKey() string {
	return os.Getenv(c.LLM.APIKeyEnv)
}

// EditorPath returns the configured editor, falling back to $EDITOR then "vi".
func (c *Config) EditorPath() string {
	if c.Editor != "" {
		return c.Editor
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}
