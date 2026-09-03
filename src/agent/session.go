package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/mcptools"
	"agentsmith/src/util"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string        `json:"id"`
	Date      time.Time     `json:"date"`
	Messages  []*ai.Message `json:"messages"`
	Summary   string        `json:"summary"`
	temporary bool          `json:"-"`
}

func LoadSessions() []*Session {
	log.D("Loading sessions from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()
	sessions := make([]*Session, 0, 32)

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open session db")
	defer db.Close()

	query := "SELECT session_id, date, summary, data FROM sessions ORDER BY date DESC;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select sessions from DB")
	defer rows.Close()

	for rows.Next() {
		var session Session
		var dataJSON string
		var dateStr string
		var summary sql.NullString

		// Scan the row data into variables
		err = rows.Scan(&session.ID, &dateStr, &summary, &dataJSON)
		if err != nil {
			log.W("Failed to scan session row:", err)
			continue
		}

		session.Date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.W("Failed to parse session date: ", dateStr, err)
			session.Date = time.Time{}
		}

		if summary.Valid {
			session.Summary = summary.String
		} else {
			session.Summary = ""
		}

		// Unmarshal the JSON data from the 'data' column into Messages
		if dataJSON != "" {
			err = json.Unmarshal([]byte(dataJSON), &session.Messages)
			if err != nil {
				log.W("Failed to unmarshal messages for session:", session.ID, err)
				session.Messages = make([]*ai.Message, 0)
			}
		} else {
			session.Messages = make([]*ai.Message, 0)
		}

		// Append the successfully loaded session to the slice
		sessions = append(sessions, &session)
	}

	log.D("Loaded sessions from DB:", len(sessions))
	return sessions
}

func newSession() *Session {
	session := &Session{uuid.NewString(), time.Now(), make([]*ai.Message, 0, 32), "New chat", false}
	return session
}

func NewTempSession() *Session {
	session := &Session{uuid.NewString(), time.Now(), make([]*ai.Message, 0, 32), "New chat", true}
	return session
}

func (s *Session) Save() (err error) {
	// log.D("Saving session to ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	var db *sql.DB
	db, err = sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	messagesJSON, err := json.Marshal(s.Messages)
	log.CheckE(err, nil, "Failed to marshal messages for session ", s.ID)

	// Use INSERT OR REPLACE (UPSERT) to handle both new and existing sessions
	query := `
	INSERT INTO sessions (session_id, date, summary, data)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(session_id) DO UPDATE SET
		date=excluded.date,
		summary=excluded.summary,
		data=excluded.data;
	`
	// Format date to a standard string format for SQLite
	dateStr := s.Date.Format(time.RFC3339)

	_, err = db.Exec(query, s.ID, dateStr, s.Summary, string(messagesJSON))
	log.CheckW(err, "Failed to update session DB")

	log.D("Saved session", s.ID)
	return
}

func (s *Session) Delete() {
	log.D("Deleting session from ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	query := "DELETE FROM sessions WHERE session_id=?"
	db.Exec(query, s.ID)
}
func (s *Session) AddMessage(origin ai.MessageOrigin, text string, toolRequests []*mcptools.ToolCallRequest) error {
	message := &ai.Message{
		ID:           uuid.NewString(),
		Origin:       origin,
		Text:         text,
		ToolRequests: toolRequests,
	}

	s.Messages = append(s.Messages, message)
	s.Date = time.Now()

	sseCh <- &SSEMessage{Type: SSEMessageNewMessage, Data: map[string]any{"message": message, "sessionId": s.ID}}
	sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: s}

	var err error
	if !s.temporary {
		err = s.Save()
	}

	return err
}

func (s *Session) UpdateLastMessage(newText string) {
	if len(s.Messages) > 0 {
		message := s.Messages[len(s.Messages)-1]
		message.Text = message.Text + newText
		sseCh <- &SSEMessage{
			Type: SSEMessageLastMessageUpdate,
			Data: map[string]any{
				"sessionId": s.ID,
				"message":   message,
			}}
	}

}

// RecordResponseStatistics stores the final assistant response measurements
// and delivers them through the existing last-message update path.
func (s *Session) RecordResponseStatistics(outputTokens int, elapsed time.Duration) error {
	if len(s.Messages) == 0 || outputTokens <= 0 || elapsed <= 0 {
		return errors.New("invalid response statistics")
	}

	message := s.Messages[len(s.Messages)-1]
	if message.Origin != ai.MessageOriginAI {
		return errors.New("response statistics require an assistant message")
	}

	elapsedMilliseconds := elapsed.Milliseconds()
	if elapsedMilliseconds == 0 {
		elapsedMilliseconds = 1
	}

	message.OutputTokens = outputTokens
	message.ElapsedMilliseconds = elapsedMilliseconds
	if !s.temporary {
		if err := s.Save(); err != nil {
			return err
		}
	}
	sseCh <- &SSEMessage{Type: SSEMessageLastMessageUpdate, Data: map[string]any{"sessionId": s.ID, "message": message}}
	return nil
}

func (s *Session) ClearMessages() {
	s.Messages = make([]*ai.Message, 0, 32)
}

const titleGenerationSysPrompt = `You are naming a chat conversation based on its content so far. ` +
	`Respond with a short, human-readable title of 2-4 words summarizing what the conversation is about. ` +
	`Respond with only the title text - no punctuation, no quotes, no explanation.`

// filterMessagesForTitle keeps only user/assistant messages with non-empty
// (post-trim) text, which is the content used to generate a session title.
func filterMessagesForTitle(messages []*ai.Message) []*ai.Message {
	filtered := make([]*ai.Message, 0, len(messages))
	for _, message := range messages {
		if message.Origin != ai.MessageOriginUser && message.Origin != ai.MessageOriginAI {
			continue
		}
		if strings.TrimSpace(message.Text) == "" {
			continue
		}
		filtered = append(filtered, message)
	}
	return filtered
}

// buildTitlePrompt flattens the filtered conversation into a single user
// turn. Passing the raw multi-turn history straight to the model - which
// necessarily ends on an assistant turn, since that's the last thing said in
// the real conversation - leaves some models (observed with LM Studio
// serving a non-reasoning instruct model) confused about whose turn it is:
// they emit an empty completion and stop immediately instead of continuing.
// Ending on a user turn asking for the title reliably elicits an actual
// response.
func buildTitlePrompt(messages []*ai.Message) string {
	var b strings.Builder
	for _, message := range messages {
		role := "User"
		if message.Origin == ai.MessageOriginAI {
			role = "Assistant"
		}
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(strings.TrimSpace(message.Text))
		b.WriteString("\n")
	}
	return b.String()
}

// maxSanitizedTitleLen is well above any real 2-4 word title. A sanitized
// result longer than this is treated as a leaked reasoning trace rather than
// a title.
const maxSanitizedTitleLen = 60

// sanitizeGeneratedTitle strips any reasoning/thinking content the model may
// have emitted before the actual title, then trims whitespace and any
// surrounding quotes left over from the model's response formatting.
//
// Some providers (observed with LM Studio serving a reasoning model, once
// the conversation history already contains a prior assistant turn) don't
// wrap reasoning in <think> tags at all - reasoning and the final title are
// both plain text with no delimiter, so CutThinking has nothing to strip.
// These traces consistently end with the actual short title as the last
// paragraph, separated by a blank line, so once CutThinking has done what it
// can, fall back to the last non-empty line whenever what's left is still
// clearly too long to be a 2-4 word title.
func sanitizeGeneratedTitle(raw string) string {
	title := util.CutThinking(raw)
	title = strings.TrimSpace(title)

	if len(title) > maxSanitizedTitleLen || strings.Contains(title, "\n") {
		lines := strings.Split(title, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			if candidate := strings.TrimSpace(lines[i]); candidate != "" {
				title = candidate
				break
			}
		}
	}

	title = strings.Trim(title, `"'`)
	return strings.TrimSpace(title)
}

// MaybeGenerateTitle (re)generates the session's Summary from the
// conversation content so far, if the session is eligible: it is not
// temporary, and it has had between 1 and 3 user messages so far. Title
// (re)generation runs asynchronously so it never delays the chat response;
// on success it persists the new Summary and broadcasts it via the existing
// SSEMessageSessionUpdate mechanism.
func (s *Session) MaybeGenerateTitle(model *ai.Model) {
	if s.temporary {
		return
	}

	userMessageCount := 0
	for _, message := range s.Messages {
		if message.Origin == ai.MessageOriginUser {
			userMessageCount++
		}
	}
	if userMessageCount == 0 || userMessageCount > 3 {
		return
	}

	if model == nil {
		return
	}

	filtered := filterMessagesForTitle(s.Messages)
	if len(filtered) == 0 {
		return
	}
	promptMessages := []*ai.Message{
		{Origin: ai.MessageOriginUser, Text: buildTitlePrompt(filtered)},
	}

	go func() {
		defer logger.BreakOnError()

		model.Provider.WaitForAllowance()
		response, err := model.Provider.ChatCompletion(promptMessages, titleGenerationSysPrompt, model, nil)
		if err != nil {
			log.W("Failed to generate session title:", err)
			return
		}

		title := sanitizeGeneratedTitle(response)
		if title == "" {
			return
		}

		s.Summary = title

		if !s.temporary {
			if err := s.Save(); err != nil {
				log.W("Failed to save session after title generation:", err)
			}
		}

		sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: s}
	}()
}
