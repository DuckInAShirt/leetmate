package config

import "testing"

func TestPresetSiliconflowFillsFields(t *testing.T) {
	c := Default()
	c.LLM = LLMConfig{Preset: "siliconflow"}
	c.applyDefaults()
	if c.LLM.Provider != "openai" {
		t.Errorf("provider = %q, want openai", c.LLM.Provider)
	}
	if c.LLM.BaseURL != "https://api.siliconflow.cn/v1" {
		t.Errorf("base_url = %q", c.LLM.BaseURL)
	}
	if c.LLM.Model != "THUDM/GLM-4-9B-0414" {
		t.Errorf("model = %q", c.LLM.Model)
	}
	if c.LLM.APIKeyEnv != "SILICONFLOW_API_KEY" {
		t.Errorf("api_key_env = %q", c.LLM.APIKeyEnv)
	}
}

func TestPresetAllowsModelOverride(t *testing.T) {
	c := Default()
	c.LLM = LLMConfig{Preset: "siliconflow", Model: "Qwen/Qwen2.5-72B-Instruct"}
	c.applyDefaults()
	if c.LLM.Model != "Qwen/Qwen2.5-72B-Instruct" {
		t.Errorf("user model override ignored: %q", c.LLM.Model)
	}
	// Preset-derived fields still apply.
	if c.LLM.APIKeyEnv != "SILICONFLOW_API_KEY" {
		t.Errorf("api_key_env = %q", c.LLM.APIKeyEnv)
	}
}

func TestDefaultPresetIsGemini(t *testing.T) {
	c := Default()
	c.applyDefaults()
	if c.LLM.Provider != "gemini" || c.LLM.Model != "gemini-2.0-flash" {
		t.Errorf("default should resolve to gemini: %+v", c.LLM)
	}
}
