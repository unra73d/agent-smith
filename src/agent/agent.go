// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/mcptools"
	"errors"
	"sync"

	"github.com/google/shlex"
)

var log = logger.Logger("agent", 1, 1, 1)

type agent struct {
	flashSession *Session
	builtinTools []*mcptools.Tool
	apiProviders []ai.IAPIProvider
	roles        []*Role
	mcps         []*mcptools.MCPServer
	sessions     []*Session
}

var Agent = agent{
	builtinTools: make([]*mcptools.Tool, 0),
	apiProviders: make([]ai.IAPIProvider, 0),
	roles:        make([]*Role, 0),
	mcps:         make([]*mcptools.MCPServer, 0),
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
		Agent.mcps = mcptools.LoadMCPServers()
	}()

	// load builtin tools
	signal.Add(1)
	go func() {
		defer signal.Done()
		Agent.builtinTools = mcptools.GetBuiltinTools()
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

func GetMCPServers() []*mcptools.MCPServer {
	return Agent.mcps
}

func GetTools() []*mcptools.Tool {
	res := make([]*mcptools.Tool, 0, 32)
	for _, mcp := range Agent.mcps {
		res = append(res, mcp.Tools...)
	}
	// res = append(res, Agent.builtinTools...)

	return res
}

func GetBuiltinTools() []*mcptools.Tool {
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

func DeleteMessage(sessionID string, messageID string) error {
	// find the session
	// find the message
	// if its assistant message - delete backwards possible chain of tool requests and calls
	for _, session := range Agent.sessions {
		if session.ID == sessionID {
			for i, message := range session.Messages {
				if message.ID == messageID {
					if message.Origin == ai.MessageOriginUser {
						session.Messages = append(session.Messages[:i], session.Messages[i+1:]...)
						session.Save()
						sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: session}
						return nil
					} else if message.Origin == ai.MessageOriginAI {
						for k := i - 1; k >= 0; k-- {
							if session.Messages[k].Origin == ai.MessageOriginUser {
								session.Messages = append(session.Messages[:k+1], session.Messages[i+1:]...)
								session.Save()
								sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: session}
								break
							}
						}
						session.Messages = append(session.Messages[:i], session.Messages[i+1:]...)
						session.Save()
						sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: session}
					}
					return nil
				}
			}
			break
		}
	}
	log.E("trying to delete non existing message", sessionID, messageID)
	return errors.New("message not found")
}

func TruncateSession(sessionID string, messageID string) error {
	for _, session := range Agent.sessions {
		if session.ID == sessionID {
			for i, message := range session.Messages {
				if message.ID == messageID {
					session.Messages = session.Messages[:i]
					session.Save()
					sseCh <- &SSEMessage{Type: SSEMessageSessionUpdate, Data: session}
					return nil
				}
			}
			break
		}
	}
	log.E("trying to delete non existing message", sessionID, messageID)
	return errors.New("message not found")
}

func TestMCPServer(Name string, Type string, URL string, Command string, Args string) (res bool) {
	res = false
	defer logger.BreakOnError()

	argArray, err := shlex.Split(Args)
	log.CheckE(err, nil, "failed to parse CLI arguments for MCP")

	mcp := &mcptools.MCPServer{
		Name:      Name,
		Transport: mcptools.MCPTransport(Type),
		URL:       URL,
		Command:   Command,
		Args:      argArray,
	}

	return mcp.Test()
}

func CreateMCPServer(Name string, Type string, URL string, Command string, Args string) (err error) {
	defer logger.BreakOnError()

	argArray, err := shlex.Split(Args)
	log.CheckE(err, nil, "failed to parse CLI arguments for MCP")

	mcp := &mcptools.MCPServer{
		Name:      Name,
		Transport: mcptools.MCPTransport(Type),
		URL:       URL,
		Command:   Command,
		Args:      argArray,
	}
	err = mcp.LoadTools()
	log.CheckE(err, nil, "failed to load MCP server")

	mcp.Save()
	Agent.mcps = append(Agent.mcps, mcp)
	sseCh <- &SSEMessage{
		Type: SSEMessageMCPListUpdate,
		Data: Agent.mcps,
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
