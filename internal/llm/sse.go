package llm

import (
	"bufio"
	"io"
	"strings"
)

// streamSSE reads a Server-Sent Events body and emits Chunk values on the
// returned channel. Each "data: <json>" line is handed to parse; "[DONE]" ends
// the stream. The body is closed and the channel closed when done. A scanner
// error surfaces as a final Err chunk.
func streamSSE(body io.ReadCloser, parse func([]byte) (Chunk, error)) <-chan Chunk {
	ch := make(chan Chunk, 16)
	go func() {
		defer close(ch)
		defer body.Close()
		sc := bufio.NewScanner(body)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				return
			}
			if payload == "" {
				continue
			}
			chunk, err := parse([]byte(payload))
			if err != nil {
				continue // skip malformed chunk rather than killing the stream
			}
			if chunk.Text != "" || chunk.Err != nil {
				ch <- chunk
			}
		}
		if err := sc.Err(); err != nil {
			ch <- Chunk{Err: err}
		}
	}()
	return ch
}
