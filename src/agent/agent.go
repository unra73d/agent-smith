// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/tools"
	"errors"
	"time"

	"github.com/google/uuid"
)

type agent struct {
	sessions     map[string]*Session
	apiProviders map[string]ai.IAPIProvider
	flashSession *Session
	roles        map[string]*Role
	mcps         map[string]*tools.MCPServer
}

var log = logger.Logger("agent", 1, 1, 1)
var Agent = agent{
	sessions:     make(map[string]*Session),
	apiProviders: make(map[string]ai.IAPIProvider),
	roles:        make(map[string]*Role),
	mcps:         make(map[string]*tools.MCPServer),
}

func LoadAgent() {
	// load api providers
	providerList := ai.LoadProviders()
	for _, provider := range providerList {
		Agent.apiProviders[uuid.NewString()] = provider
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

func GetModels() []*ai.Model {
	models := make([]*ai.Model, 0, 32)
	for _, apiProvider := range Agent.apiProviders {
		models = append(models, apiProvider.Models()...)
	}
	return models
}

func GetSessions() map[string]*Session {
	return Agent.sessions
}

func GetRoles() map[string]*Role {
	return Agent.roles
}

func GetMCPServers() map[string]*tools.MCPServer {
	return Agent.mcps
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
	model := findModel(modelID)
	if model != nil {
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

	model := findModel(modelID)
	if model != nil {
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

func findModel(modelID string) *ai.Model {
	for _, provider := range Agent.apiProviders {
		for _, model := range provider.Models() {
			if model.ID == modelID {
				return model
			}
		}
	}
	return nil
}
