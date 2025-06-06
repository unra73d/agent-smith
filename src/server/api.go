package server

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skratchdot/open-golang/open"
)

/*
Get list of chat sessions
*/
var listSessionsURI = "/sessions/list"

func listSessionsHandler(c *gin.Context) {
	c.JSON(200, map[string]any{"sessions": agent.GetSessions()})
}

/*
Create new session
*/
var createSessionURI = "/sessions/new"

func createSessionHandler(c *gin.Context) {
	session := agent.CreateSession()
	c.JSON(200, map[string]any{"session": session})
}

/*
Delete session by id
*/
var deleteSessionURI = "/sessions/delete/:id"

type DeleteSessionReq struct {
	ID string `uri:"id" binding:"required"`
}

func deleteSessionHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req DeleteSessionReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.DeleteSession(req.ID)
	if err != nil {
		c.JSON(500, map[string]any{"error": err})
	} else {
		c.JSON(200, map[string]any{"error": nil})
	}

}

/*
Deletes all messages from session starting with given id
*/
var truncateSessionURI = "/sessions/:sessionId/truncate/:messageId"

type TruncateSessionReq struct {
	SessionID string `uri:"sessionId" binding:"required"`
	MessageID string `uri:"messageId" binding:"required"`
}

func truncateSessionHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req TruncateSessionReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.TruncateSession(req.SessionID, req.MessageID)
	if err != nil {
		c.JSON(500, map[string]any{"error": err})
	} else {
		c.JSON(200, map[string]any{"error": nil})
	}

}

/*
Delete message by id
*/
var deleteMessageURI = "/sessions/:sessionId/messages/delete/:messageId"

type DeleteMessageReq struct {
	SessionID string `uri:"sessionId" binding:"required"`
	MessageID string `uri:"messageId" binding:"required"`
}

func deleteMessageHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req DeleteMessageReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.DeleteMessage(req.SessionID, req.MessageID)
	if err != nil {
		c.JSON(500, map[string]any{"error": err})
	} else {
		c.JSON(200, map[string]any{"error": nil})
	}

}

/*
Get list of available models
*/
var listModelsURI = "/models/list"

func listModelsHandler(c *gin.Context) {
	c.JSON(200, map[string]any{"models": agent.GetModels()})
}

/*
Get list of available AI providers
*/
var listProvidersURI = "/providers/list"

func listProvidersHandler(c *gin.Context) {
	c.JSON(200, map[string]any{"providers": agent.GetProviders()})
}

/*
Test AI provider connectivity
*/
var testProviderURI = "/provider/test"

type testProviderReq struct {
	Name      string `json:"name" binding:"required"`
	APIURL    string `json:"url" binding:"required"`
	APIKey    string `json:"apiKey,omitempty"`
	RateLimit int    `json:"rateLimit,omitempty"`
}

func testProviderHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req testProviderReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	c.JSON(200, map[string]any{"response": agent.TesProvider(req.Name, req.APIURL, req.APIKey, req.RateLimit)})
}

/*
Update AI Provider
*/
var updateProviderURI = "/provider/update"

type updateProviderReq struct {
	ID        string `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	APIURL    string `json:"url" binding:"required"`
	APIKey    string `json:"apiKey,omitempty"`
	RateLimit int    `json:"rateLimit,omitempty"`
}

func updateProviderHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req updateProviderReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.UpdateProvider(req.ID, req.Name, req.APIURL, req.APIKey, req.RateLimit)
	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err})
	}
}

/*
Create AI Provider
*/
var createProviderURI = "/provider/create"

type createProviderReq struct {
	Name      string `json:"name" binding:"required"`
	APIURL    string `json:"url" binding:"required"`
	APIKey    string `json:"apiKey,omitempty"`
	RateLimit int    `json:"rateLimit,omitempty"`
}

func createProviderHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req createProviderReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.CreateProvider(req.Name, req.APIURL, req.APIKey, req.RateLimit)
	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err})
	}
}

/*
Delete provider by id
*/
var deleteProviderURI = "/provider/delete/:id"

type DeleteProviderReq struct {
	ID string `uri:"id" binding:"required"`
}

func deleteProviderHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req DeleteProviderReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.DeleteProvider(req.ID)
	if err != nil {
		c.JSON(500, map[string]any{"error": err})
	} else {
		c.JSON(200, map[string]any{"error": nil})
	}

}

/*
API for sending message to AI directly and get response via system SSE connection. This call will return when generation ends.
No tools will be called in response.
*/
var directChatStreamURI = "/directchat/stream"

type directChatStreamReq struct {
	SessionID string `json:"sessionID" binding:"required"`
	ModelID   string `json:"modelID" binding:"required"`
	RoleID    string `json:"roleID"`
	Message   string `json:"message" binding:"required"`
}

func directChatStreamHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req directChatStreamReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	streamDoneCh := make(chan bool)

	go agent.DirectChatStreaming(c.Request.Context(), req.SessionID, req.ModelID, req.RoleID, strings.TrimSpace(req.Message), streamDoneCh)

	// blocking call
	c.Stream(func(w io.Writer) bool {
		for {
			select {
			case <-streamDoneCh:
				log.D("Stream finalized")
				c.Status(200)
				return false
			case <-c.Request.Context().Done():
				return false
			case <-time.After(3600 * time.Second):
				log.W("Stream message timed out")
				c.Status(500)
				return false
			}
		}
	})
}

/*
API for sending message to agent in non-streaming mode. It can be used directly but originally intended for LLM
calling this as a tool.
Internal behavior is same as tool streaming chat, meaning it can call tools and even recursively call this API.
The difference to regular tool chat API is that output returned as complete message instead of streaming.
In this mode agent also considers a depth of recursion and may decide to break it if deemed too deep to prevent infinite loops.
No messages or sessions are saved during this call.
Agent configured only via system prompt
*/
var dynamicAgentChatURI = "/dynamicagentchat"

type dynamicAgentChatReq struct {
	ModelID   string `json:"modelID" binding:"required"`
	Message   string `json:"message" binding:"required"`
	SysPrompt string `json:"sysPrompt" binding:"required"`
}

func dynamicAgentChatHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req dynamicAgentChatReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	response, err := agent.DynamicAgentChat(req.ModelID, req.Message, req.SysPrompt)
	if err != nil {
		c.JSON(500, map[string]string{"error": "Unknown error"})
	} else {
		c.JSON(200, map[string]string{"response": response, "error": ""})
	}
}

/*
API for sending message to AI through an agent. Internally agent and LLM can exchange multiple messages
and evet include user in this loop. Loop can be broken by LLM, user prompt or if agent detects recursion
LLM can call tools or dynamic agents.
Response is SSE stream
*/
var toolChatStreamURI = "/toolchat/stream"

type toolChatStreamReq struct {
	SessionID string `json:"sessionID"`
	ModelID   string `json:"modelID" binding:"required"`
	RoleID    string `json:"roleID"`
	Message   string `json:"message" binding:"required"`
}

func toolChatStreamHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req toolChatStreamReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	streamDoneCh := make(chan bool)

	go agent.ToolChatStreaming(c.Request.Context(), req.SessionID, req.ModelID, req.RoleID, strings.TrimSpace(req.Message), streamDoneCh)

	// blocking call
	c.Stream(func(w io.Writer) bool {
		for {
			select {
			case <-streamDoneCh:
				log.D("Stream finalized")
				c.Status(200)
				return false
			case <-c.Request.Context().Done():
				return false
			case <-time.After(30 * time.Second):
				w.Write([]byte("."))
				c.Writer.Flush()
			case <-time.After(3600 * time.Second):
				log.W("Stream message timed out")
				c.Status(500)
				return false
			}
		}
	})
}

/*
Get list of available roles
*/
var listRolesURI = "/roles/list"

func listRolesHandler(c *gin.Context) {
	roles := agent.GetRoles()
	c.JSON(200, map[string]any{"roles": roles})
}

type roleReq struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name" binding:"required"`
	GeneralInstruction string `json:"generalInstruction"`
	Role               string `json:"role"`
	Style              string `json:"style"`
}

/*
Create new role
*/
var createRoleURI = "/roles/create"

func createRoleHandler(c *gin.Context) {
	defer logger.BreakOnError()
	var req roleReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")
	role, err := agent.CreateRole(agent.RoleConfig{
		Name:               req.Name,
		GeneralInstruction: req.GeneralInstruction,
		Role:               req.Role,
		Style:              req.Style,
	})
	if err == nil {
		c.JSON(200, map[string]any{"role": role})
	} else {
		c.JSON(500, map[string]any{"error": err.Error()})
	}
}

/*
Update role
*/
var updateRoleURI = "/roles/update"

func updateRoleHandler(c *gin.Context) {
	defer logger.BreakOnError()
	var req roleReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")
	role, err := agent.UpdateRole(req.ID, agent.RoleConfig{
		Name:               req.Name,
		GeneralInstruction: req.GeneralInstruction,
		Role:               req.Role,
		Style:              req.Style,
	})
	if err == nil {
		c.JSON(200, map[string]any{"role": role})
	} else {
		c.JSON(500, map[string]any{"error": err.Error()})
	}
}

/*
Delete role
*/
var deleteRoleURI = "/roles/delete/:id"

type deleteRoleReq struct {
	ID string `uri:"id" binding:"required"`
}

func deleteRoleHandler(c *gin.Context) {
	defer logger.BreakOnError()
	var req deleteRoleReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")
	err = agent.DeleteRole(req.ID)
	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err.Error()})
	}
}

/*
Get list of available MCP servers
*/
var listMCPServersURI = "/mcp/list"

func listMCPServersHandler(c *gin.Context) {
	c.JSON(200, map[string]any{"mcpServers": agent.GetMCPServers()})
}

/*
Test MCP server connectivity
*/
var testMCPServerURI = "/mcp/test"

type testMCPServerReq struct {
	Name      string `json:"name" binding:"required"`
	Transport string `json:"transport" binding:"required"`
	URL       string `json:"url,omitempty"`
	Command   string `json:"command,omitempty"`
	Active    bool   `json:"active"` // New field
}

func testMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req testMCPServerReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	c.JSON(200, map[string]any{"response": agent.TestMCPServer(req.Name, req.Transport, req.URL, req.Command, req.Active)})
}

/*
Create new MCP server
*/
var createMCPServerURI = "/mcp/create"

type createMCPServerReq struct {
	Name      string `json:"name" binding:"required"`
	Transport string `json:"transport" binding:"required"`
	URL       string `json:"url,omitempty"`
	Command   string `json:"command,omitempty"`
	Active    bool   `json:"active"` // New field
}

func createMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req createMCPServerReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.CreateMCPServer(req.Name, req.Transport, req.URL, req.Command, req.Active)
	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err})
	}
}

/*
Update MCP server
*/
var updateMCPServerURI = "/mcp/update"

type updateMCPServerReq struct {
	ID        string `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Transport string `json:"transport" binding:"required"`
	URL       string `json:"url,omitempty"`
	Command   string `json:"command,omitempty"`
	Active    bool   `json:"active"` // New field
}

func updateMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req updateMCPServerReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.UpdateMCPServer(req.ID, req.Name, req.Transport, req.URL, req.Command, req.Active)
	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err})
	}
}

/*
Delete MCP server by id
*/
var deleteMCPServerURI = "/mcp/delete/:id"

type DeleteMCPServerReq struct {
	ID string `uri:"id" binding:"required"`
}

func deleteMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req DeleteMCPServerReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = agent.DeleteMCPServer(req.ID)
	if err != nil {
		c.JSON(500, map[string]any{"error": err})
	} else {
		c.JSON(200, map[string]any{"error": nil})
	}

}

/*
Open URL in default browser
*/
var openLinkURI = "/desktop/url/open"

type openLinkReq struct {
	URL string `json:"url" binding:"required"`
}

func openLinkHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req openLinkReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	err = open.Run(req.URL)

	if err == nil {
		c.JSON(200, map[string]any{"error": nil})
	} else {
		c.JSON(500, map[string]any{"error": err})
	}
}

/*
SSE connection for receiving server updates. It implements following events:
- session_update:{date, summary}
- new_message:{origin, text}
- last_message_update:{sessionId, text}
- mcp_list_update:[{mcp}]
- provider_list_update:[{provider}]
- role_list_update:[{role}]
*/
var sseURI = "/sse"

func sseHandler(c *gin.Context) {
	defer logger.BreakOnError()

	sseCh := agent.NewSSEConnection()

	heartbeat := time.NewTicker(10 * time.Second)
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-sseCh:
			if !ok {
				return false
			}
			_, err := json.Marshal(msg.Data)
			log.CheckW(err, "failed to marshal sse message", msg)
			if err == nil {
				c.SSEvent(string(msg.Type), msg.Data)
			}
		case <-heartbeat.C:
			c.SSEvent("heartbeat", []byte("keep-alive"))
		}
		return true
	})
}

func InitAgentRoutes(router *gin.Engine) {
	group := router.Group("/agent")
	{
		group.GET(listSessionsURI, listSessionsHandler)
		group.GET(createSessionURI, createSessionHandler)
		group.GET(deleteSessionURI, deleteSessionHandler)
		group.GET(truncateSessionURI, truncateSessionHandler)
		group.GET(deleteMessageURI, deleteMessageHandler)

		group.GET(listModelsURI, listModelsHandler)
		group.GET(listProvidersURI, listProvidersHandler)
		group.POST(testProviderURI, testProviderHandler)
		group.POST(updateProviderURI, updateProviderHandler)
		group.POST(createProviderURI, createProviderHandler)
		group.GET(deleteProviderURI, deleteProviderHandler)

		group.POST(directChatStreamURI, directChatStreamHandler)
		group.POST(dynamicAgentChatURI, dynamicAgentChatHandler)
		group.POST(toolChatStreamURI, toolChatStreamHandler)

		group.GET(listRolesURI, listRolesHandler)
		group.POST(createRoleURI, createRoleHandler)
		group.POST(updateRoleURI, updateRoleHandler)
		group.GET(deleteRoleURI, deleteRoleHandler)

		group.GET(listMCPServersURI, listMCPServersHandler)
		group.POST(testMCPServerURI, testMCPServerHandler)
		group.POST(createMCPServerURI, createMCPServerHandler)
		group.POST(updateMCPServerURI, updateMCPServerHandler)
		group.GET(deleteMCPServerURI, deleteMCPServerHandler)

		group.POST(openLinkURI, openLinkHandler)

		group.GET(sseURI, sseHandler)
	}
}
