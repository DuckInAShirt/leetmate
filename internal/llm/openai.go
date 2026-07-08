package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
)

// openaiProvider talks to any OpenAI-compatible /chat/completions endpoint
// (Groq, DeepSeek, SiliconFlow, local ollama, OpenAI itself).
type openaiProvider struct {
	cfg config.LLMConfig
	key string
}

func newOpenAI(cfg config.LLMConfig, key string) (Provider, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	return &openaiProvider{cfg: cfg, key: key}, nil
}

type oaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type oaiChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Reasoning        string `json:"reasoning"`
		} `json:"delta"`
	} `json:"choices"`
}

func (p *openaiProvider) Chat(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error) {
	msgs := make([]oaiMessage, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, oaiMessage{Role: string(m.Role), Content: m.Content})
	}
	reqBody := map[string]any{
		"model":    p.cfg.Model,
		"messages": msgs,
		"stream":   true,
	}
	if opts.Temperature != 0 {
		reqBody["temperature"] = opts.Temperature
	}
	if opts.MaxTokens != 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}
	if p.disableThinkingByDefault() {
		reqBody["enable_thinking"] = false
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	url := p.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, friendlyError("LLM", resp.StatusCode, b)
	}
	return streamSSE(resp.Body, parseOpenAIChunk), nil
}

func (p *openaiProvider) disableThinkingByDefault() bool {
	baseURL := strings.ToLower(p.cfg.BaseURL)
	model := strings.ToLower(p.cfg.Model)
	return strings.Contains(baseURL, "siliconflow") && strings.Contains(model, "qwen")
}

func parseOpenAIChunk(payload []byte) (Chunk, error) {
	var oc oaiChunk
	if err := json.Unmarshal(payload, &oc); err != nil {
		return Chunk{}, err
	}
	if len(oc.Choices) == 0 {
		return Chunk{}, nil
	}
	delta := oc.Choices[0].Delta
	if delta.Content != "" {
		return Chunk{Text: delta.Content}, nil
	}
	if delta.ReasoningContent != "" {
		return Chunk{Text: delta.ReasoningContent, Kind: ChunkReasoning}, nil
	}
	if delta.Reasoning != "" {
		return Chunk{Text: delta.Reasoning, Kind: ChunkReasoning}, nil
	}
	return Chunk{}, nil
}
