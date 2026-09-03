package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStreamingTestProvider(t *testing.T, events ...string) (*APIProvider, *Model) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		for _, event := range events {
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		}
	}))
	t.Cleanup(server.Close)

	provider := &APIProvider{APIURL: server.URL, APIType: APITypeOpenAICompatible}
	return provider, &Model{ID: "test-model", Provider: provider}
}

func TestChatCompletionStream_CapturesTerminalCompletionUsage(t *testing.T) {
	provider, model := newStreamingTestProvider(t,
		`{"choices":[{"delta":{"content":"hello "}}]}`,
		`{"choices":[{"delta":{"content":"world"}}]}`,
		`{"choices":[],"usage":{"completion_tokens":7}}`,
		`[DONE]`,
	)
	writeCh := make(chan string, 2)

	tokens, err := provider.ChatCompletionStream(context.Background(), nil, "", model, nil, writeCh, nil)
	if err != nil {
		t.Fatalf("ChatCompletionStream() error = %v", err)
	}
	if tokens != 7 {
		t.Fatalf("completion tokens = %d, want 7", tokens)
	}
	if got := <-writeCh + <-writeCh; got != "hello world" {
		t.Fatalf("forwarded text = %q, want %q", got, "hello world")
	}
}

func TestChatCompletionStream_ReportsMissingUsage(t *testing.T) {
	provider, model := newStreamingTestProvider(t,
		`{"choices":[{"delta":{"content":"answer"}}]}`,
		`[DONE]`,
	)
	writeCh := make(chan string, 1)

	tokens, err := provider.ChatCompletionStream(context.Background(), nil, "", model, nil, writeCh, nil)
	if err != nil {
		t.Fatalf("ChatCompletionStream() error = %v", err)
	}
	if tokens != 0 {
		t.Fatalf("completion tokens = %d, want unavailable (0)", tokens)
	}
	if got := <-writeCh; got != "answer" {
		t.Fatalf("forwarded text = %q, want %q", got, "answer")
	}
}
