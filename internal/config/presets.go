package config

// Preset is a ready-made LLM configuration. Users set `llm.preset` in
// config.yaml and only need to put the matching API key in .env — provider,
// base_url, model and the key's env-var name are filled in automatically.
//
// Pick by platform: e.g. preset: siliconflow. Override just the model if you
// want a different one on the same platform.
type Preset struct {
	Name      string // value for llm.preset
	Provider  string // "gemini" or "openai"
	BaseURL   string // OpenAI-compatible base URL (empty for gemini)
	Model     string // default model id
	APIKeyEnv string // env var name holding the key
	Desc      string
	SignupURL string // where to get a key
	Free      bool   // whether a free tier is available
}

// Presets lists the built-in LLM presets. The first one (gemini) is the
// default; siliconflow is the recommended pick for users in mainland China.
var Presets = []Preset{
	{
		Name: "gemini", Provider: "gemini", Model: "gemini-2.0-flash",
		APIKeyEnv: "GEMINI_API_KEY", Free: true,
		Desc:      "Google Gemini，全球可用，有免费 tier",
		SignupURL: "https://aistudio.google.com/apikey",
	},
	{
		Name: "siliconflow", Provider: "openai", BaseURL: "https://api.siliconflow.cn/v1",
		Model: "THUDM/GLM-4-9B-0414", APIKeyEnv: "SILICONFLOW_API_KEY", Free: true,
		Desc:      "硅基流动（国内访问稳；免费模型多为 7-9B 级、需实名认证；可在 llm.model 覆盖）",
		SignupURL: "https://cloud.siliconflow.cn",
	},
	{
		Name: "groq", Provider: "openai", BaseURL: "https://api.groq.com/openai/v1",
		Model: "llama-3.3-70b-versatile", APIKeyEnv: "GROQ_API_KEY", Free: true,
		Desc:      "Groq（免费且极快，海外网络）",
		SignupURL: "https://console.groq.com/keys",
	},
	{
		Name: "deepseek", Provider: "openai", BaseURL: "https://api.deepseek.com/v1",
		Model: "deepseek-chat", APIKeyEnv: "DEEPSEEK_API_KEY", Free: false,
		Desc:      "DeepSeek 官方（极便宜，指令遵循强）",
		SignupURL: "https://platform.deepseek.com",
	},
}

// presetByName returns the preset with the given name, or nil.
func presetByName(name string) *Preset {
	for i := range Presets {
		if Presets[i].Name == name {
			return &Presets[i]
		}
	}
	return nil
}
