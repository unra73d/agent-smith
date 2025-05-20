package server

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
Response is SSE stream:
{
}
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
	roleList := make([]map[string]any, 0, len(roles))
	for _, val := range roles {
		roleList = append(roleList, map[string]any{
			"name":               val.Config.Name,
			"generalInstruction": val.Config.GeneralInstruction,
			"role":               val.Config.Role,
			"style":              val.Config.Style,
			"id":                 val.ID,
		})
	}
	c.JSON(200, map[string]any{"roles": roleList})
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
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
	URL     string `json:"url,omitempty"`
	Command string `json:"command,omitempty"`
	Args    string `json:"args,omitempty"`
}

func testMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req testMCPServerReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	c.JSON(200, map[string]any{"response": agent.TestMCPServer(req.Name, req.Type, req.URL, req.Command, req.Args)})
}

/*
Create new MCP server
*/
var createMCPServerURI = "/mcp/create"

type createMCPServerReq struct {
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
	URL     string `json:"url,omitempty"`
	Command string `json:"command,omitempty"`
	Args    string `json:"args,omitempty"`
}

func createMCPServerHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req createMCPServerReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	c.JSON(200, map[string]any{"error": agent.CreateMCPServer(req.Name, req.Type, req.URL, req.Command, req.Args)})
}

/*
SSE connection for receiving server updates. It implements following events:
- session_update:{date, summary}
- new_message:{origin, text}
- last_message_update:{sessionId, text}
- session_list_update:[{session}]
- model_list_update:[{model}]
- mcp_list_update:[{mcp}]
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

		group.POST(directChatStreamURI, directChatStreamHandler)
		group.POST(dynamicAgentChatURI, dynamicAgentChatHandler)
		group.POST(toolChatStreamURI, toolChatStreamHandler)

		group.GET(listRolesURI, listRolesHandler)

		group.GET(listMCPServersURI, listMCPServersHandler)
		group.POST(testMCPServerURI, testMCPServerHandler)
		group.POST(createMCPServerURI, createMCPServerHandler)

		group.GET(sseURI, sseHandler)
	}
}
