package agent

import (
	"agentsmith/src/ai"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB points AS_AGENT_DB_FILE at a fresh temporary sqlite DB with the
// sessions table created, mirroring the schema in src/server/debug.go.
func setupTestDB(t *testing.T) {
	t.Helper()

	dbFile := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("AS_AGENT_DB_FILE", dbFile)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		date DATETIME,
		summary TEXT,
		data TEXT
	);`)
	if err != nil {
		t.Fatalf("failed to create sessions table: %v", err)
	}
}

// newTestModel spins up an httptest server that answers /chat/completions
// with the given title text, and wraps it in an *ai.Model ready to be used
// as the "turn's model" argument to MaybeGenerateTitle.
func newTestModel(t *testing.T, title string) *ai.Model {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": title,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	provider := &ai.APIProvider{
		ID:      "test-provider",
		Name:    "test-provider",
		APIURL:  server.URL,
		APIType: ai.APITypeOpenAICompatible,
	}

	return &ai.Model{ID: "test-model", Name: "test-model", Provider: provider}
}

// buildExchangeMessages builds a plausible message history containing
// userCount user messages, each followed by an assistant reply.
func buildExchangeMessages(userCount int) []*ai.Message {
	messages := make([]*ai.Message, 0, userCount*2)
	for i := 0; i < userCount; i++ {
		messages = append(messages, &ai.Message{ID: "u", Origin: ai.MessageOriginUser, Text: "user message"})
		messages = append(messages, &ai.Message{ID: "a", Origin: ai.MessageOriginAI, Text: "assistant reply"})
	}
	return messages
}

func drainNoSSE(t *testing.T) {
	t.Helper()
	select {
	case msg := <-sseCh:
		t.Fatalf("expected no SSE message to be sent, got: %+v", msg)
	default:
	}
}

func expectSessionUpdateSSE(t *testing.T, sessionID string) *SSEMessage {
	t.Helper()
	select {
	case msg := <-sseCh:
		if msg.Type != SSEMessageSessionUpdate {
			t.Fatalf("expected SSEMessageSessionUpdate, got %v", msg.Type)
		}
		session, ok := msg.Data.(*Session)
		if !ok {
			t.Fatalf("expected SSE data to be *Session, got %T", msg.Data)
		}
		if session.ID != sessionID {
			t.Fatalf("expected SSE session id %q, got %q", sessionID, session.ID)
		}
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for session_update SSE message")
		return nil
	}
}

func TestMaybeGenerateTitle_ExchangeCountEligibility(t *testing.T) {
	setupTestDB(t)

	cases := []struct {
		name         string
		userCount    int
		wantTriggers bool
	}{
		{"zero user messages", 0, false},
		{"one user message", 1, true},
		{"two user messages", 2, true},
		{"three user messages", 3, true},
		{"four user messages", 4, false},
		{"five user messages", 5, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := newTestModel(t, "Generated Title")
			session := &Session{
				ID:       "session-" + tc.name,
				Date:     time.Now(),
				Messages: buildExchangeMessages(tc.userCount),
			}

			session.MaybeGenerateTitle(model)

			if tc.wantTriggers {
				expectSessionUpdateSSE(t, session.ID)
			} else {
				drainNoSSE(t)
			}
		})
	}
}

func TestFilterMessagesForTitle_ExcludesToolAndEmptyMessages(t *testing.T) {
	messages := []*ai.Message{
		{ID: "1", Origin: ai.MessageOriginUser, Text: "hello"},
		{ID: "2", Origin: ai.MessageOriginTool, Text: "tool result"},
		{ID: "3", Origin: ai.MessageOriginAI, Text: "hi there"},
		{ID: "4", Origin: ai.MessageOriginUser, Text: "   "},
		{ID: "5", Origin: ai.MessageOriginUser, Text: ""},
		{ID: "6", Origin: ai.MessageOriginSystem, Text: "system prompt"},
		{ID: "7", Origin: ai.MessageOriginAI, Text: "  final answer  "},
	}

	filtered := filterMessagesForTitle(messages)

	if len(filtered) != 3 {
		t.Fatalf("expected 3 filtered messages, got %d: %+v", len(filtered), filtered)
	}
	wantIDs := []string{"1", "3", "7"}
	for i, id := range wantIDs {
		if filtered[i].ID != id {
			t.Errorf("expected filtered[%d].ID = %q, got %q", i, id, filtered[i].ID)
		}
	}
}

func TestBuildTitlePrompt_EndsOnUserTurn(t *testing.T) {
	messages := []*ai.Message{
		{ID: "1", Origin: ai.MessageOriginUser, Text: "hi"},
		{ID: "2", Origin: ai.MessageOriginAI, Text: "Hi! How can I assist you today?"},
	}

	prompt := buildTitlePrompt(messages)

	want := "User: hi\nAssistant: Hi! How can I assist you today?\n"
	if prompt != want {
		t.Fatalf("buildTitlePrompt() = %q, want %q", prompt, want)
	}
}

func TestSanitizeGeneratedTitle_StripsThinkingContent(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "think tag",
			input: "<think>reasoning about the topic here</think>Trip Planning",
			want:  "Trip Planning",
		},
		{
			name:  "thinking tag",
			input: "<thinking>lots of reasoning</thinking>  Budget Review  ",
			want:  "Budget Review",
		},
		{
			name:  "no thinking tag",
			input: "Weather Forecast",
			want:  "Weather Forecast",
		},
		{
			name:  "wrapped in quotes",
			input: `"Recipe Ideas"`,
			want:  "Recipe Ideas",
		},
		{
			name:  "think tag plus quotes",
			input: "<think>hmm</think>'Trip Planning'",
			want:  "Trip Planning",
		},
		{
			name: "untagged reasoning trace falls back to last paragraph",
			input: "The user just said \"say 123\", which is a simple instruction to output " +
				"the number 123. I need a short 2-4 word title.\n\n" +
				"Possible titles: Number 123, Digit Output, Simple Request.\n\n" +
				"I'll go with \"Number 123\" as the title.\n\n\nNumber 123",
			want: "Number 123",
		},
		{
			name:  "short multi-line response reduces to its last non-empty line",
			input: "Trip\nPlanning",
			want:  "Planning",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeGeneratedTitle(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeGeneratedTitle(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestMaybeGenerateTitle_TemporarySessionsNeverTitled(t *testing.T) {
	setupTestDB(t)
	model := newTestModel(t, "Generated Title")

	session := &Session{
		ID:        "temp-session",
		Date:      time.Now(),
		Messages:  buildExchangeMessages(1),
		temporary: true,
	}

	session.MaybeGenerateTitle(model)

	drainNoSSE(t)
	if session.Summary != "" {
		t.Fatalf("expected temporary session summary to remain unset, got %q", session.Summary)
	}
}

func TestMaybeGenerateTitle_PersistsAndBroadcasts(t *testing.T) {
	setupTestDB(t)
	model := newTestModel(t, "<think>internal reasoning</think>Trip Planning  ")

	session := &Session{
		ID:       "persist-session",
		Date:     time.Now(),
		Messages: buildExchangeMessages(1),
	}

	session.MaybeGenerateTitle(model)

	expectSessionUpdateSSE(t, session.ID)

	if session.Summary != "Trip Planning" {
		t.Fatalf("expected session.Summary = %q, got %q", "Trip Planning", session.Summary)
	}

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	var summary string
	err = db.QueryRow("SELECT summary FROM sessions WHERE session_id = ?", session.ID).Scan(&summary)
	if err != nil {
		t.Fatalf("failed to query persisted summary: %v", err)
	}
	if summary != "Trip Planning" {
		t.Fatalf("expected persisted summary = %q, got %q", "Trip Planning", summary)
	}
}

func TestResponseStatisticsPersistAndBroadcastFinalMessage(t *testing.T) {
	setupTestDB(t)
	previousSSECh := sseCh
	sseCh = make(chan *SSEMessage, 1)
	t.Cleanup(func() { sseCh = previousSSECh })

	session := &Session{
		ID:       "statistics-session",
		Date:     time.Now(),
		Messages: []*ai.Message{{ID: "answer", Origin: ai.MessageOriginAI, Text: "final answer"}},
	}
	if err := session.RecordResponseStatistics(12, 1500*time.Millisecond); err != nil {
		t.Fatalf("RecordResponseStatistics() error = %v", err)
	}

	msg := <-sseCh
	if msg.Type != SSEMessageLastMessageUpdate {
		t.Fatalf("SSE type = %q, want %q", msg.Type, SSEMessageLastMessageUpdate)
	}
	payload := msg.Data.(map[string]any)
	updated := payload["message"].(*ai.Message)
	if updated.OutputTokens != 12 || updated.ElapsedMilliseconds != 1500 {
		t.Fatalf("SSE message statistics = (%d, %d), want (12, 1500)", updated.OutputTokens, updated.ElapsedMilliseconds)
	}

	loaded := LoadSessions()
	if len(loaded) != 1 || len(loaded[0].Messages) != 1 {
		t.Fatalf("loaded sessions = %+v, want one session with one message", loaded)
	}
	persisted := loaded[0].Messages[0]
	if persisted.OutputTokens != 12 || persisted.ElapsedMilliseconds != 1500 {
		t.Fatalf("persisted statistics = (%d, %d), want (12, 1500)", persisted.OutputTokens, persisted.ElapsedMilliseconds)
	}
}

func TestLegacyMessageLoadsWithoutResponseStatistics(t *testing.T) {
	setupTestDB(t)
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	legacyMessages := `[{"id":"legacy-answer","origin":"assistant","text":"old answer","toolRequests":null}]`
	if _, err := db.Exec("INSERT INTO sessions (session_id, date, summary, data) VALUES (?, ?, ?, ?)", "legacy-session", time.Now().Format(time.RFC3339), "", legacyMessages); err != nil {
		t.Fatalf("failed to insert legacy session: %v", err)
	}

	loaded := LoadSessions()
	if len(loaded) != 1 || len(loaded[0].Messages) != 1 {
		t.Fatalf("loaded sessions = %+v, want one session with one message", loaded)
	}
	message := loaded[0].Messages[0]
	if message.OutputTokens != 0 || message.ElapsedMilliseconds != 0 {
		t.Fatalf("legacy response statistics = (%d, %d), want unavailable", message.OutputTokens, message.ElapsedMilliseconds)
	}
}
