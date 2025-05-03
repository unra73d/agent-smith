package server

import (
	"agentsmith/src/agent"
	"agentsmith/src/logger"

	"github.com/gin-gonic/gin"
)

/*
Connect to last session or to session with specific id
*/
var agentConnectIDURI = "/connect/:id"
var agentConnectURI = "/connect"

type AgentConnectReq struct {
	ID string `uri:"id" binding:"omitempty"`
}

func agentConnectHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req AgentConnectReq
	err := c.BindUri(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	res, err := agent.ConnectSession(req.ID)
	if err != nil {
		c.JSON(500, map[string]string{"error": "Session with requested id does not exist"})
	} else {
		c.JSON(200, map[string]any{"session": res, "error": ""})
	}
}

/*
Get list of available models and active model id
*/
var agentListModelsURI = "/models/list"

func agentListModelsHandler(c *gin.Context) {
	models, id := agent.GetModels()
	c.JSON(200, map[string]any{"models": models, "activeModelID": id})
}

/*
API for sending message to AI directly and get response.
No tools will be called in response.
*/
var agentDirectChatURI = "/directchat"

type AgentDirectChatReq struct {
	SessionID string `json:"sessionID" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

func agentDirectChatHandler(c *gin.Context) {
	defer logger.BreakOnError()

	var req AgentDirectChatReq
	err := c.Bind(&req)
	log.CheckE(err, func() { c.Status(400) }, "Failed to unpack API parameters")

	response, err := agent.DirectChat(req.SessionID, req.Message)
	if err != nil {
		c.JSON(500, map[string]string{"error": "Unknown error"})
	} else {
		c.JSON(200, map[string]string{"response": response, "error": ""})
	}
}

func InitAgentRoutes(router *gin.Engine) {
	group := router.Group("/agent")
	{
		group.GET(agentConnectIDURI, agentConnectHandler)
		group.GET(agentConnectURI, agentConnectHandler)
		group.GET(agentListModelsURI, agentListModelsHandler)
		group.POST(agentDirectChatURI, agentDirectChatHandler)
	}
}
