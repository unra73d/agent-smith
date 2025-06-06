// Package agent implements logic of the AI agent that orchestrates requests between AI model
// and various tools
package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/mcptools"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var log = logger.Logger("agent", 1, 1, 1)

type agent struct {
	flashSession *Session
	builtinTools []*mcptools.Tool
	apiProviders []*ai.APIProvider
	roles        []*Role
	mcps         []*mcptools.MCPServer
	sessions     []*Session
}

var Agent = agent{
	builtinTools: make([]*mcptools.Tool, 0),
	apiProviders: make([]*ai.APIProvider, 0),
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
		Agent.mcps = mcptools.LoadMCPServers(onMCPUpdate)
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
		models = append(models, apiProvider.Models...)
	}
	return models
}

func GetProviders() []*ai.APIProvider {
	return Agent.apiProviders
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

func onMCPUpdate(mcp *mcptools.MCPServer) {
	sseCh <- &SSEMessage{
		Type: SSEMessageMCPListUpdate,
		Data: Agent.mcps,
	}
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

func TestMCPServer(name string, transport string, url string, command string, active bool) (res bool) {
	res = false
	defer logger.BreakOnError()

	mcp := &mcptools.MCPServer{
		Name:      name,
		Transport: mcptools.MCPTransport(transport),
		URL:       url,
		Command:   command,
		Active:    active, // Set the Active field
	}

	return mcp.Test()
}

func CreateMCPServer(name string, transport string, url string, command string, active bool) (err error) {
	defer logger.BreakOnError()

	mcp := &mcptools.MCPServer{
		ID:        uuid.NewString(),
		Name:      name,
		Transport: mcptools.MCPTransport(transport),
		URL:       url,
		Command:   command,
		Active:    active,
	}
	go func() {
		mcp.LoadTools()
		mcp.Loaded = true
		sseCh <- &SSEMessage{
			Type: SSEMessageMCPListUpdate,
			Data: Agent.mcps,
		}
		log.W(err, "failed to load MCP server")
	}()

	mcp.Save()
	Agent.mcps = append(Agent.mcps, mcp)
	sseCh <- &SSEMessage{
		Type: SSEMessageMCPListUpdate,
		Data: Agent.mcps,
	}

	return
}

func UpdateMCPServer(id string, name string, transport string, url string, command string, active bool) (err error) {
	defer logger.BreakOnError()

	err = errors.New("trying to update non existing MCP")
	for _, mcp := range Agent.mcps {
		if mcp.ID == id {
			mcp.Name = name
			mcp.Active = active
			if mcp.Transport != mcptools.MCPTransport(transport) || mcp.URL != url || mcp.Command != command {
				mcp.Transport = mcptools.MCPTransport(transport)
				mcp.URL = url
				mcp.Command = command
				mcp.Tools = []*mcptools.Tool{}

				go func() {
					err := mcp.LoadTools()
					mcp.Loaded = true
					sseCh <- &SSEMessage{
						Type: SSEMessageMCPListUpdate,
						Data: Agent.mcps,
					}
					log.CheckW(err, "failed to load MCP server")
				}()
			}
			err = mcp.Save()

			sseCh <- &SSEMessage{
				Type: SSEMessageMCPListUpdate,
				Data: Agent.mcps,
			}
			err = nil
			break
		}
	}
	return
}

func DeleteMCPServer(ID string) (err error) {
	err = errors.New("trying to delete non existing MCP")
	for i, mcp := range Agent.mcps {
		if mcp.ID == ID {
			mcp.Delete()
			Agent.mcps = append(Agent.mcps[:i], Agent.mcps[i+1:]...)
			sseCh <- &SSEMessage{
				Type: SSEMessageMCPListUpdate,
				Data: Agent.mcps,
			}
			err = nil
			break
		}
	}
	return
}

func TesProvider(Name string, APIURL string, APIKey string, RateLimit int) (res bool) {
	provider := &ai.APIProvider{
		Name:    Name,
		APIURL:  APIURL,
		APIKey:  APIKey,
		APIType: ai.APITypeOpenAICompatible,
	}
	return provider.Test()
}

func UpdateProvider(ID string, Name string, APIURL string, APIKey string, RateLimit int) (err error) {
	for _, provider := range Agent.apiProviders {
		if provider.ID == ID {
			provider.Name = Name
			provider.RateLimit = RateLimit
			if APIURL != provider.APIURL || APIKey != provider.APIKey {
				provider.APIURL = APIURL
				provider.APIKey = APIKey
				go func() {
					err := provider.LoadModels()
					sseCh <- &SSEMessage{
						Type: SSEMessageProviderListUpdate,
						Data: Agent.apiProviders,
					}
					log.CheckW(err, "failed to load models")
				}()
			}

			err = provider.Save()

			sseCh <- &SSEMessage{SSEMessageProviderListUpdate, Agent.apiProviders}
			break
		}
	}
	return err
}

func CreateProvider(Name string, APIURL string, APIKey string, RateLimit int) error {
	provider, err := ai.NewProvider(uuid.NewString(), ai.APITypeOpenAICompatible, Name, APIURL, APIKey, RateLimit)
	provider.Save()
	Agent.apiProviders = append(Agent.apiProviders, provider)
	sseCh <- &SSEMessage{SSEMessageProviderListUpdate, Agent.apiProviders}
	return err
}

func DeleteProvider(id string) (err error) {
	logger.BreakOnError()
	for i, provider := range Agent.apiProviders {
		if provider.ID == id {
			err = provider.Delete()
			log.CheckE(err, nil, "failed to delete provider")
			Agent.apiProviders = append(Agent.apiProviders[:i], Agent.apiProviders[i+1:]...)
			sseCh <- &SSEMessage{SSEMessageProviderListUpdate, Agent.apiProviders}
			return nil
		}
	}
	log.E("trying to delete non existing provider", id)
	return errors.New("provider not found")
}

func CreateRole(config RoleConfig) (*Role, error) {
	role := &Role{
		ID:     uuid.NewString(),
		Config: config,
	}
	err := role.Save()
	if err == nil {
		Agent.roles = append(Agent.roles, role)
		sseCh <- &SSEMessage{SSEMessageRoleListUpdate, Agent.roles}
	}
	return role, err
}

func UpdateRole(id string, config RoleConfig) (*Role, error) {
	for _, role := range Agent.roles {
		if role.ID == id {
			role.Config = config
			err := role.Save()
			sseCh <- &SSEMessage{SSEMessageRoleListUpdate, Agent.roles}
			return role, err
		}
	}
	return nil, errors.New("role not found")
}

func DeleteRole(id string) error {
	for i, role := range Agent.roles {
		if role.ID == id {
			role.Delete()
			Agent.roles = append(Agent.roles[:i], Agent.roles[i+1:]...)
			sseCh <- &SSEMessage{SSEMessageRoleListUpdate, Agent.roles}
			return nil
		}
	}
	return errors.New("role not found")
}

func findModel(modelID string) *ai.Model {
	for _, provider := range Agent.apiProviders {
		for _, model := range provider.Models {
			if model.ID == modelID {
				return model
			}
		}
	}
	return nil
}
