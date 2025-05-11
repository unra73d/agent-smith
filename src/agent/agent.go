// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/tools"
	"errors"
	"sync"
)

var log = logger.Logger("agent", 1, 1, 1)

type agent struct {
	flashSession *Session
	builtinTools []*tools.Tool
	apiProviders []ai.IAPIProvider
	roles        []*Role
	mcps         []*tools.MCPServer
	sessions     []*Session
}

var Agent = agent{
	builtinTools: make([]*tools.Tool, 0),
	apiProviders: make([]ai.IAPIProvider, 0),
	roles:        make([]*Role, 0),
	mcps:         make([]*tools.MCPServer, 0),
	sessions:     make([]*Session, 0),
}

func LoadAgent() {
	var signal sync.WaitGroup

	// load api providers
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.apiProviders = ai.LoadProviders()
	}()

	// load historical sessions
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.sessions = LoadSessions()
	}()

	// create global 'flash' session
	Agent.flashSession = newSession()

	// load roles
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.roles = LoadRoles()
	}()

	// load MCP servers
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.mcps = tools.LoadMCPServers()
	}()

	// load builtin tools
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.builtinTools = tools.GetBuiltinTools()
	}()

	signal.Wait()
}

func GetModels() []*ai.Model {
	models := make([]*ai.Model, 0, 32)
	for _, apiProvider := range Agent.apiProviders {
		models = append(models, apiProvider.Models()...)
	}
	return models
}

func GetSessions() []*Session {
	return Agent.sessions
}

func GetRoles() []*Role {
	return Agent.roles
}

func GetMCPServers() []*tools.MCPServer {
	return Agent.mcps
}

func GetBuiltinTools() []*tools.Tool {
	return Agent.builtinTools
}

func CreateSession() *Session {
	session := newSession()
	session.Save()
	Agent.sessions = append(Agent.sessions, session)
	return session
}

func DeleteSession(id string) error {
	for i, session := range Agent.sessions {
		if session.ID == id {
			session.Delete()
			Agent.sessions = append(Agent.sessions[:i], Agent.sessions[i+1:]...)
			return nil
		}
	}
	log.E("trying to delete non existing session", id)
	return errors.New("session not found")
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
