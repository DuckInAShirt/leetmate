package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

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
type oaiRequest struct {
	Model       string      `json:"model"`
	Messages    []oaiMessage `json:"messages"`
	Stream      bool        `json:"stream"`
	Temperature float32     `json:"temperature,omitempty"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
}
type oaiChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (p *openaiProvider) Chat(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error) {
	msgs := make([]oaiMessage, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, oaiMessage{Role: string(m.Role), Content: m.Content})
	}
	body, err := json.Marshal(oaiRequest{
		Model:       p.cfg.Model,
		Messages:    msgs,
		Stream:      true,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
	})
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

func parseOpenAIChunk(payload []byte) (string, error) {
	var oc oaiChunk
	if err := json.Unmarshal(payload, &oc); err != nil {
		return "", err
	}
	if len(oc.Choices) == 0 {
		return "", nil
	}
	return oc.Choices[0].Delta.Content, nil
}
