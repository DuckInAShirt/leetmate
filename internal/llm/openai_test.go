package llm

import (
	"testing"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

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

func TestDisableThinkingByDefaultOnlyForSiliconFlowQwen(t *testing.T) {
	p := &openaiProvider{cfg: config.LLMConfig{BaseURL: "https://api.siliconflow.cn/v1", Model: "Qwen/Qwen3.6-35B-A3B"}}
	if !p.disableThinkingByDefault() {
		t.Fatal("expected SiliconFlow Qwen to disable thinking")
	}
	p.cfg.Model = "deepseek-ai/DeepSeek-V4-flash"
	if p.disableThinkingByDefault() {
		t.Fatal("did not expect non-Qwen model to disable thinking")
	}
	p.cfg.BaseURL = "https://api.openai.com/v1"
	p.cfg.Model = "Qwen/Qwen3.6-35B-A3B"
	if p.disableThinkingByDefault() {
		t.Fatal("did not expect non-SiliconFlow endpoint to disable thinking")
	}
}
