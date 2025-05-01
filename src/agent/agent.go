// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/logger"
	"errors"
)

type agent struct {
	sessions      []Session
	activeSession *Session
	models        []Model
	activeModel   *Model
}

var log = logger.Logger("agent", 1, 1, 1)
var Agent agent

func LoadAgent() {
	// load models
	Agent.models = LoadModels()
	// select active model
	if len(Agent.models) > 0 {
		Agent.activeModel = &Agent.models[0]
	}

	// load historical sessions
	Agent.sessions = LoadSessions()
	// select active session
	if len(Agent.sessions) > 0 {
		Agent.activeSession = &Agent.sessions[0]
	}
}

func ConnectSession(id string) (*Session, error) {
	var res *Session = nil
	var err error = nil
	if id == "" {
		if len(Agent.sessions) == 0 {
			// if there are no sessions in array, create it, add to sessions array and return id
			newSess := NewSession()
			Agent.sessions = append(Agent.sessions, *newSess)
			Agent.activeSession = newSess
			res = newSess
		} else {
			// get the last session
			lastSession := &Agent.sessions[len(Agent.sessions)-1]
			Agent.activeSession = lastSession
			res = lastSession
		}
	} else {
		// if id is not nil then search for that session and return its id
		for i := range Agent.sessions {
			if Agent.sessions[i].ID == id {
				Agent.activeSession = &Agent.sessions[i]
				res = &Agent.sessions[i]
			}
		}
		// if there is no session found with this id, return error
		err = errors.New("session not found")
	}

	return res, err
}

func DirectChat(sessionID string, message string) (string, error) {
	Agent.activeSession.AddMessage(MessageOriginUser, message)

	response := "received"
	Agent.activeSession.AddMessage(MessageOriginAI, response)

	return response, nil
}
