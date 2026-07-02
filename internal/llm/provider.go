// Package llm provides a provider-agnostic chat abstraction with streaming.
// Implementations live in sub-files (gemini.go, openai.go); the router in this
// file picks one based on config.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

// Role identifies the speaker of a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single chat message.
type Message struct {
	Role    Role
	Content string
}

// Options tunes generation. Zero values mean "provider default".
type Options struct {
	MaxTokens   int
	Temperature float32
}

// Chunk is one piece of a streamed response. Text is the incremental text;
// Err, when non-nil, signals a mid-stream failure (the channel is closed
// immediately after). A clean end is signaled by the channel closing with no
// Err chunk sent.
type Chunk struct {
	Text string
	Err  error
}

// Provider is the streaming chat interface every backend implements.
type Provider interface {
	// Chat starts a streaming completion. The returned error is for immediate
	// setup failures (auth, bad model). Text arrives on the channel; the channel
	// closes when the response is complete.
	Chat(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error)
}

// New selects and constructs a Provider from config. The API key is read from
// the environment variable named by cfg.APIKeyEnv.
func New(cfg config.LLMConfig) (Provider, error) {
	key := os.Getenv(cfg.APIKeyEnv)
	if key == "" {
		return nil, fmt.Errorf("llm: API key env var %s is not set — get a free key and export it (e.g. GEMINI_API_KEY=...)", cfg.APIKeyEnv)
	}
	switch cfg.Provider {
	case "gemini", "":
		return newGemini(cfg, key)
	case "openai":
		return newOpenAI(cfg, key)
	default:
		return nil, fmt.Errorf("llm: unknown provider %q (want gemini or openai)", cfg.Provider)
	}
}

// apiErrorPayload matches the {"error":{"message":...}} shape both Gemini and
// OpenAI-compatible endpoints return.
type apiErrorPayload struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// parseAPIError extracts a human-readable message from an error response body,
// falling back to a trimmed copy of the raw body.
func parseAPIError(body []byte) string {
	var p apiErrorPayload
	if err := json.Unmarshal(body, &p); err == nil && p.Error.Message != "" {
		return p.Error.Message
	}
	s := string(body)
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}

// friendlyError turns a non-200 response into a concise, actionable message.
func friendlyError(provider string, status int, body []byte) error {
	msg := parseAPIError(body)
	switch status {
	case 401:
		return fmt.Errorf("%s 认证失败（HTTP 401）：%s\n检查 API key 是否正确。",
			provider, msg)
	case 403:
		low := strings.ToLower(msg)
		// SiliconFlow returns "Model disabled" (code 30003) when the account
		// hasn't completed real-name verification — free models require it.
		if strings.Contains(low, "model disabled") || strings.Contains(low, "30003") || strings.Contains(low, "verify") {
			return fmt.Errorf("%s 该模型不可用（HTTP 403）：%s\n几乎都是因为还没「实名认证」——国内平台的免费模型都要求先实名。\n去对应平台完成实名认证后即可解锁；或在 config.yaml 的 llm.model 改一个你账号可用的模型。",
				provider, msg)
		}
		return fmt.Errorf("%s 权限不足（HTTP 403）：%s\n常见原因：API key 错误、未实名认证、或该模型未对当前账号启用。",
			provider, msg)
	case 429:
		return fmt.Errorf("%s 配额超限（HTTP 429）：%s\n提示：免费额度有每分钟/每日上限——可换一个 key，或在 config.yaml 改用 openai provider 接 Groq/DeepSeek/SiliconFlow（都有免费额度）。",
			provider, msg)
	case 404:
		return fmt.Errorf("%s 模型不存在（HTTP 404）：%s\n检查 config.yaml 里的 model 名是否正确。", provider, msg)
	default:
		return fmt.Errorf("%s HTTP %d：%s", provider, status, msg)
	}
}
