package agent

import (
	"agentsmith/src/ai"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupStreamingAgent(t *testing.T, handler http.HandlerFunc) (*Session, string) {
	t.Helper()
	setupTestDB(t)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	previousAgent := Agent
	previousSSECh := sseCh
	sseCh = make(chan *SSEMessage, 32)
	provider := &ai.APIProvider{APIURL: server.URL, APIType: ai.APITypeOpenAICompatible}
	model := &ai.Model{ID: "streaming-model", Provider: provider}
	session := &Session{ID: "streaming-session", Date: time.Now(), Messages: make([]*ai.Message, 0)}
	Agent.apiProviders = []*ai.APIProvider{{Models: []*ai.Model{model}}}
	Agent.sessions = []*Session{session}
	Agent.roles = nil
	Agent.mcps = nil
	Agent.builtinTools = nil
	t.Cleanup(func() {
		Agent = previousAgent
		sseCh = previousSSECh
	})
	return session, model.ID
}

func sseResponse(w http.ResponseWriter, events ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	flusher := w.(http.Flusher)
	for _, event := range events {
		fmt.Fprintf(w, "data: %s\n\n", event)
		flusher.Flush()
	}
}

func TestDirectChatRecordsStatisticsOnlyAfterSuccessfulCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		session, modelID := setupStreamingAgent(t, func(w http.ResponseWriter, r *http.Request) {
			sseResponse(w, `{"choices":[{"delta":{"content":"final answer"}}]}`, `{"choices":[],"usage":{"completion_tokens":9}}`, `[DONE]`)
		})
		done := make(chan bool, 1)
		DirectChatStreaming(context.Background(), session.ID, modelID, "", "question", done)
		if ok := <-done; !ok {
			t.Fatal("direct chat reported failure")
		}
		answer := session.Messages[len(session.Messages)-1]
		if answer.OutputTokens != 9 || answer.ElapsedMilliseconds <= 0 {
			t.Fatalf("successful answer statistics = (%d, %d), want provider tokens and positive duration", answer.OutputTokens, answer.ElapsedMilliseconds)
		}
	})

	t.Run("failed completion", func(t *testing.T) {
		session, modelID := setupStreamingAgent(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "provider failure", http.StatusInternalServerError)
		})
		done := make(chan bool, 1)
		DirectChatStreaming(context.Background(), session.ID, modelID, "", "question", done)
		if ok := <-done; ok {
			t.Fatal("direct chat reported success for incomplete stream")
		}
		answer := session.Messages[len(session.Messages)-1]
		if answer.OutputTokens != 0 || answer.ElapsedMilliseconds != 0 {
			t.Fatalf("failed answer statistics = (%d, %d), want unavailable", answer.OutputTokens, answer.ElapsedMilliseconds)
		}
	})

	t.Run("cancelled completion", func(t *testing.T) {
		session, modelID := setupStreamingAgent(t, func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := make(chan bool, 1)
		DirectChatStreaming(ctx, session.ID, modelID, "", "question", done)
		if ok := <-done; ok {
			t.Fatal("direct chat reported success for cancelled stream")
		}
		answer := session.Messages[len(session.Messages)-1]
		if answer.OutputTokens != 0 || answer.ElapsedMilliseconds != 0 {
			t.Fatalf("cancelled answer statistics = (%d, %d), want unavailable", answer.OutputTokens, answer.ElapsedMilliseconds)
		}
	})
}
