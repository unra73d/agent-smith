// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"errors"
	"time"

	"github.com/google/uuid"
)

type agent struct {
	sessions     map[string]*Session
	models       map[string]*ai.Model
	flashSession *Session
	roles        map[string]*Role
}

var log = logger.Logger("agent", 1, 1, 1)
var Agent = agent{
	sessions: make(map[string]*Session),
	models:   make(map[string]*ai.Model),
	roles:    make(map[string]*Role),
}

func LoadAgent() {
	// load models
	modelList := ai.LoadModels()
	for _, model := range modelList {
		Agent.models[uuid.NewString()] = model
	}

	// load historical sessions
	sessionList := LoadSessions()
	for _, session := range sessionList {
		Agent.sessions[session.ID] = session
	}

	// create global 'flash' session
	Agent.flashSession = newSession()

	// load roles
	roleList := LoadRoles()
	for _, role := range roleList {
		Agent.roles[role.ID] = role
	}
}

func GetModels() map[string]*ai.Model {
	return Agent.models
}

func GetSessions() map[string]*Session {
	return Agent.sessions
}

func GetRoles() map[string]*Role {
	return Agent.roles
}

func CreateSession() *Session {
	session := newSession()
	session.Save()
	Agent.sessions[session.ID] = session
	return session
}

func DeleteSession(id string) error {
	if session, ok := Agent.sessions[id]; ok {
		session.Delete()
		delete(Agent.sessions, id)
		return nil
	} else {
		log.E("trying to delete non existing session", id)
		return errors.New("session not found")
	}
}

func DirectChatStreaming(sessionID string, modelID string, roleID string, query string, streamCh chan string, streamDoneCh chan bool) {
	if model, ok := Agent.models[modelID]; ok {
		session, permanentSession := Agent.sessions[sessionID]
		if !permanentSession {
			session = Agent.flashSession
		}

		sysPrompt := ""
		if role, ok := Agent.roles[roleID]; ok {
			sysPrompt = "## General instruction: \n" + role.Config.GeneralInstruction +
				"## Role and personality: \n" + role.Config.Role +
				"## Text style and tone: \n" + role.Config.Style
		}

		modelResponseCh := make(chan string)
		modelDoneCh := make(chan bool)
		go func() {
			for {
				select {
				case msg := <-modelResponseCh:
					session.UpdateLastMessage(msg)
					streamCh <- msg
				case <-modelDoneCh:
					if permanentSession {
						session.Save()
					}
					return
				case <-time.After(60 * time.Second):
					return
				}
			}
		}()

		session.AddMessage(ai.MessageOriginUser, query)
		err := session.AddMessage(ai.MessageOriginAI, "")
		log.CheckW(err, "Failed to add new message in agent")

		model.Provider.ChatCompletionStream(
			session.Messages[:len(session.Messages)-1],
			sysPrompt,
			model,
			false,
			modelResponseCh,
		)
		modelDoneCh <- true
		streamDoneCh <- true
	} else {
		log.E("Model not found")
		streamDoneCh <- false
	}
}

func DirectChat(sessionID string, modelID string, roleID string, query string) (response string, err error) {
	logger.BreakOnError()
	response = ""

	if model, ok := Agent.models[modelID]; ok {
		session, permanentSession := Agent.sessions[sessionID]
		if !permanentSession {
			session = Agent.flashSession
		}

		sysPrompt := ""
		if role, ok := Agent.roles[roleID]; ok {
			sysPrompt = "## General instruction: \n" + role.Config.GeneralInstruction +
				"## Role and personality: \n" + role.Config.Role +
				"## Text style and tone: \n" + role.Config.Style
		}

		message, err := model.Provider.ChatCompletion(
			session.Messages,
			sysPrompt,
			model,
			false,
		)
		log.CheckE(err, nil, "Failed to get completion for message")

		err = session.AddMessage(ai.MessageOriginAI, message)
		log.CheckE(err, nil, "Failed to store new message in agent")

		response = message
	} else {
		err = errors.New("model not selected")
	}
	return
}
