package llm

import (
	"errors"
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

func TestNewReturnsTypedMissingAPIKeyError(t *testing.T) {
	t.Setenv("TEST_LLM_KEY", "")
	_, err := New(config.LLMConfig{Provider: "gemini", APIKeyEnv: "TEST_LLM_KEY"})
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("New() error = %v, want ErrMissingAPIKey", err)
	}
}

func TestParseOpenAIChunkReasoningContent(t *testing.T) {
	chunk, err := parseOpenAIChunk([]byte(`{"choices":[{"delta":{"reasoning_content":"hidden thinking"}}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk: %v", err)
	}
	if chunk.Kind != ChunkReasoning {
		t.Fatalf("kind = %v, want ChunkReasoning", chunk.Kind)
	}
	if chunk.Text != "hidden thinking" {
		t.Fatalf("text = %q", chunk.Text)
	}
}

func TestParseOpenAIChunkPrefersContentOverReasoning(t *testing.T) {
	chunk, err := parseOpenAIChunk([]byte(`{"choices":[{"delta":{"content":"answer","reasoning_content":"hidden"}}]}`))
	if err != nil {
		t.Fatalf("parseOpenAIChunk: %v", err)
	}
	if chunk.Kind != ChunkText || chunk.Text != "answer" {
		t.Fatalf("chunk = %+v, want answer text", chunk)
	}
}

func TestDisableThinkingByDefaultForSiliconFlowReasoningModels(t *testing.T) {
	cases := []struct {
		name                string
		baseURL, model      string
		wantDisableThinking bool
	}{
		{"siliconflow qwen", "https://api.siliconflow.cn/v1", "Qwen/Qwen3.6-35B-A3B", true},
		{"siliconflow deepseek", "https://api.siliconflow.cn/v1", "deepseek-ai/DeepSeek-V4-flash", true},
		{"siliconflow non-reasoning model", "https://api.siliconflow.cn/v1", "zai-org/GLM-4.6", false},
		{"non-siliconflow qwen", "https://api.openai.com/v1", "Qwen/Qwen3.6-35B-A3B", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &openaiProvider{cfg: config.LLMConfig{BaseURL: tc.baseURL, Model: tc.model}}
			if got := p.disableThinkingByDefault(); got != tc.wantDisableThinking {
				t.Fatalf("disableThinkingByDefault(%s, %s) = %v, want %v", tc.baseURL, tc.model, got, tc.wantDisableThinking)
			}
		})
	}
}
