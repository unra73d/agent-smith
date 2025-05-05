package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID       string       `json:"id"`
	Date     time.Time    `json:"date"`
	Messages []ai.Message `json:"messages"`
	Summary  string       `json:"summary"`
}

func LoadSessions() []Session {
	log.D("Loading sessions from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()
	sessions := make([]Session, 0)

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open session db")
	defer db.Close()

	query := "SELECT session_id, date, data FROM sessions ORDER BY date DESC;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select sessions from DB")
	defer rows.Close()

	for rows.Next() {
		var session Session
		var dataJSON string
		var dateStr string

		// Scan the row data into variables
		err = rows.Scan(&session.ID, &dateStr, &dataJSON)
		if err != nil {
			log.W("Failed to scan session row:", err)
			continue
		}

		session.Date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.W("Failed to parse session date: ", dateStr, err)
			session.Date = time.Time{}
		}

		// Unmarshal the JSON data from the 'data' column into Messages
		if dataJSON != "" {
			err = json.Unmarshal([]byte(dataJSON), &session.Messages)
			if err != nil {
				log.W("Failed to unmarshal messages for session:", session.ID, err)
				session.Messages = make([]ai.Message, 0)
			}
		} else {
			session.Messages = make([]ai.Message, 0)
		}

		// Append the successfully loaded session to the slice
		sessions = append(sessions, session)
	}

	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		log.E("Error iterating session rows: %v", err)
	}

	if len(sessions) == 0 {
		sessions = append(sessions, *newSession())
	}
	log.D("Loaded sessions from DB:", len(sessions))
	return sessions
}

func newSession() *Session {
	session := &Session{uuid.NewString(), time.Now(), make([]ai.Message, 0, 32), "New chat"}
	session.Save()
	return session
}

func (s *Session) Save() error {
	log.D("Saving session to ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	messagesJSON, err := json.Marshal(s.Messages)
	log.CheckE(err, nil, "Failed to marshal messages for session ", s.ID)

	// Use INSERT OR REPLACE (UPSERT) to handle both new and existing sessions
	query := `
	INSERT INTO sessions (session_id, date, data)
	VALUES (?, ?, ?)
	ON CONFLICT(session_id) DO UPDATE SET
		date=excluded.date,
		data=excluded.data;
	`
	// Format date to a standard string format for SQLite
	dateStr := s.Date.Format(time.RFC3339)

	_, err = db.Exec(query, s.ID, dateStr, string(messagesJSON))
	log.CheckW(err, "Failed to update session DB")

	log.D("Saved session", s.ID)
	return err
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

func (s *Session) AddMessage(origin ai.MessageOrigin, text string) error {
	newMessage := ai.Message{
		ID:     uuid.NewString(),
		Origin: origin,
		Text:   text,
	}

	s.Messages = append(s.Messages, newMessage)
	err := s.Save()

	return err
}

func (s *Session) AddMessageFromMessage(message *ai.Message) error {
	s.Messages = append(s.Messages, *message)
	err := s.Save()

	return err
}

func (s *Session) UpdateLastMessage(newText string) {
	if len(s.Messages) > 0 {
		message := &s.Messages[len(s.Messages)-1]
		message.Text = message.Text + newText
	}

}
