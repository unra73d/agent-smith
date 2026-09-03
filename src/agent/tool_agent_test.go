package agent

import (
	"agentsmith/src/ai"
	"context"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestToolChatRecordsStatisticsOnlyOnFinalAnswer(t *testing.T) {
	var requests atomic.Int32
	session, modelID := setupStreamingAgent(t, func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			sseResponse(w,
				`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call-1","function":{"name":"lua_code_runner","arguments":"{\"code\":\"return 2\"}"}}]}}]}`,
				`[DONE]`,
			)
			return
		}
		sseResponse(w, `{"choices":[{"delta":{"content":"The answer is 2."}}]}`, `{"choices":[],"usage":{"completion_tokens":6}}`, `[DONE]`)
	})
	Agent.builtinTools = GetBuiltinTools()

	done := make(chan bool, 1)
	ToolChatStreaming(context.Background(), session.ID, modelID, "", "calculate", done)
	if ok := <-done; !ok {
		t.Fatal("tool chat reported failure")
	}
	if requests.Load() != 2 {
		t.Fatalf("provider requests = %d, want 2", requests.Load())
	}

	var intermediate, final *ai.Message
	for _, message := range session.Messages {
		if message.Origin == ai.MessageOriginAI && len(message.ToolRequests) > 0 {
			intermediate = message
		}
		if message.Origin == ai.MessageOriginAI && message.Text == "The answer is 2." {
			final = message
		}
	}
	if intermediate == nil || final == nil {
		t.Fatalf("expected intermediate tool call and final answer, got %+v", session.Messages)
	}
	if intermediate.OutputTokens != 0 || intermediate.ElapsedMilliseconds != 0 {
		t.Fatalf("intermediate statistics = (%d, %d), want unavailable", intermediate.OutputTokens, intermediate.ElapsedMilliseconds)
	}
	if final.OutputTokens != 6 || final.ElapsedMilliseconds <= 0 {
		t.Fatalf("final statistics = (%d, %d), want (6, positive)", final.OutputTokens, final.ElapsedMilliseconds)
	}
}
