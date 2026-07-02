package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"leetmate/internal/config"
)

const geminiBase = "https://generativelanguage.googleapis.com/v1beta"

type geminiProvider struct {
	cfg config.LLMConfig
	key string
}

func newGemini(cfg config.LLMConfig, key string) (Provider, error) {
	return &geminiProvider{cfg: cfg, key: key}, nil
}

type geminiPart struct {
	Text string `json:"text"`
}
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}
type geminiGenCfg struct {
	Temperature     float32 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}
type geminiReq struct {
	SystemInstruction *geminiContent `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  *geminiGenCfg   `json:"generationConfig,omitempty"`
}

// Gemini streaming responses come as SSE; each data line is a JSON object whose
// candidates[].content.parts[].text holds the incremental text.
type geminiChunk struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *geminiProvider) Chat(ctx context.Context, messages []Message, opts Options) (<-chan Chunk, error) {
	sys, contents := toGeminiContents(messages)
	body, err := json.Marshal(geminiReq{
		SystemInstruction: sys,
		Contents:          contents,
		GenerationConfig:  &geminiGenCfg{Temperature: opts.Temperature, MaxOutputTokens: opts.MaxTokens},
	})
	if err != nil {
		return nil, err
	}
	model := g.cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", geminiBase, model, g.key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, friendlyError("Gemini", resp.StatusCode, b)
	}
	return streamSSE(resp.Body, parseGeminiChunk), nil
}

// parseGeminiChunk extracts incremental text from one SSE data payload.
func parseGeminiChunk(payload []byte) (string, error) {
	var gc geminiChunk
	if err := json.Unmarshal(payload, &gc); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, c := range gc.Candidates {
		for _, p := range c.Content.Parts {
			sb.WriteString(p.Text)
		}
	}
	return sb.String(), nil
}

func toGeminiContents(messages []Message) (sys *geminiContent, contents []geminiContent) {
	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			if sys == nil {
				sys = &geminiContent{Parts: []geminiPart{{}}}
			}
			if sys.Parts[0].Text != "" {
				sys.Parts[0].Text += "\n\n"
			}
			sys.Parts[0].Text += m.Content
		case RoleUser:
			contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: m.Content}}})
		case RoleAssistant:
			// Gemini calls the assistant role "model".
			contents = append(contents, geminiContent{Role: "model", Parts: []geminiPart{{Text: m.Content}}})
		}
	}
	return sys, contents
}
