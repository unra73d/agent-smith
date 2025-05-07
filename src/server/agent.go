package server

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"
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
	sessionMap := agent.GetSessions()
	sessionList := make([]*agent.Session, 0, len(sessionMap))
	for _, val := range sessionMap {
		sessionList = append(sessionList, val)
	}
	c.JSON(200, map[string]any{"sessions": sessionList})
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
Get list of available models and active model id
*/
var listModelsURI = "/models/list"

func listModelsHandler(c *gin.Context) {
	modelMap := agent.GetModels()
	modelList := make([]map[string]any, 0, len(modelMap))
	for key, val := range modelMap {
		modelList = append(modelList, map[string]any{
			"name": val.Name,
			"id":   key,
		})
	}
	c.JSON(200, map[string]any{"models": modelList})
}

/*
API for sending message to AI directly and get response.
No tools will be called in response.
*/
var directChatURI = "/directchat"

type directChatReq struct {
	SessionID string `json:"sessionID" binding:"required"`
	ModelID   string `json:"modelID" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

func agentDirectChatHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req directChatReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	response, err := agent.DirectChat(req.SessionID, req.ModelID, strings.TrimSpace(req.Message))
	if err != nil {
		c.JSON(500, map[string]string{"error": "Unknown error"})
	} else {
		c.JSON(200, map[string]string{"response": response, "error": ""})
	}
}

/*
API for sending message to AI directly and get response as SSE events.
No tools will be called in response.
*/
var directChatStreamURI = "/directchat/stream"

type directChatStreamReq struct {
	SessionID string `json:"sessionID" binding:"required"`
	ModelID   string `json:"modelID" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

func agentDirectChatStreamHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req directChatStreamReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	streamCh := make(chan string)
	streamDoneCh := make(chan bool)

	go agent.DirectChatStreaming(req.SessionID, req.ModelID, strings.TrimSpace(req.Message), streamCh, streamDoneCh)

	// blocking call
	c.Stream(func(w io.Writer) bool {
		for {
			select {
			case msg := <-streamCh:
				w.Write([]byte(msg))
				c.Writer.Flush()
			case <-streamDoneCh:
				log.D("Stream finalized")
				c.Status(200)
				return false
			case <-time.After(100 * time.Second):
				log.W("Stream message timed out")
				c.Status(500)
				return false
			}
		}
	})
}

func InitAgentRoutes(router *gin.Engine) {
	group := router.Group("/agent")
	{
		group.GET(listSessionsURI, listSessionsHandler)
		group.GET(createSessionURI, createSessionHandler)
		group.GET(deleteSessionURI, deleteSessionHandler)

		group.GET(listModelsURI, listModelsHandler)

		group.POST(directChatURI, agentDirectChatHandler)
		group.POST(directChatStreamURI, agentDirectChatStreamHandler)
	}
}
