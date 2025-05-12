package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/tools"
	"agentsmith/src/util"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type AgentAction string

const (
	AgentActionAnswer   = "answer"
	AgentActionToolCall = "tool_call"
	AgentActionError    = "error"
)

func inferNextAction(message string) (AgentAction, *tools.MCPServer, *tools.ToolCallRequest) {
	content := util.CutThinking(message)
	if strings.HasPrefix(content, "<tool/>") {
		content = content[len("<tool/>"):]
		var callRequest tools.ToolCallRequest
		err := json.Unmarshal([]byte(content), &callRequest)

		if err != nil {
			for _, tool := range GetTools() {
				if tool.Name == callRequest.Name {
					return AgentActionToolCall, tool.Server, &callRequest
				}
			}
		}
		return AgentActionError, nil, nil
	}
	return AgentActionAnswer, nil, nil
}

const toolUsePrompt = `
## Tool usage: \n
You have tools to find information or perform actions.
If you need to use a tool, your response must start with <tool/>, followed immediately by the JSON for the tool.
When to use tools:
For real-time data (e.g., weather, news, search).
For specific calculations or code execution.
To access specialized knowledge bases, storages, databases.
Tool Call Format:
After <tool/>, provide a JSON object like this:
{
	"tool_name": "time",
	"params": {
		"location": "New York",
		"24hr": true
	}
}
Only use provided tool names and their defined parameters.
If you can answer directly from your knowledge, do so without using a tool.
`

func ToolChatStreaming(sessionID string, modelID string, roleID string, query string, streamCh chan string, streamDoneCh chan bool) {
	log.D("Tool chat initiated")
	model := findModel(modelID)
	if model != nil {
		var session *Session

		for _, s := range Agent.sessions {
			if s.ID == sessionID {
				session = s
				break
			}
		}
		if session == nil {
			log.E("Session not found")
			streamDoneCh <- false
			return
		}

		sysPrompt := ""
		// Find the role by ID
		for _, role := range Agent.roles {
			if role.ID == roleID {
				sysPrompt = "## General instruction: \n" + role.Config.GeneralInstruction +
					"## Role and personality: \n" + role.Config.Role +
					"## Text style and tone: \n" + role.Config.Style
				break
			}
		}

		sysPrompt = sysPrompt + toolUsePrompt

		modelResponseCh := make(chan string)
		modelDoneCh := make(chan bool)

		session.AddMessage(ai.MessageOriginUser, query)
		session.AddMessage(ai.MessageOriginAI, "")

		go func() {
			log.D("Starting initial chat completion")
			model.Provider.ChatCompletionStream(
				session.Messages[:len(session.Messages)-1],
				sysPrompt,
				model,
				GetTools(),
				modelResponseCh,
			)
			modelDoneCh <- true
		}()

		for {
			select {
			case msg := <-modelResponseCh:
				session.UpdateLastMessage(msg)
				streamCh <- msg
			case <-modelDoneCh:
				log.D("Model response done")
				session.Save()

				action, mcp, callRequest := inferNextAction(session.Messages[len(session.Messages)-1].Text)
				log.D("Next action: ", action)
				switch action {
				case AgentActionError:
					log.E("Error during parse of model intention")
					return
				case AgentActionAnswer:
					streamDoneCh <- true
					return
				case AgentActionToolCall:
					toolResult, _ := mcp.CallTool(callRequest)
					log.D("Tool execution result: ", toolResult)
					streamCh <- toolResult
					session.AddMessage(ai.MessageOriginTool, toolResult)
					session.AddMessage(ai.MessageOriginAI, "")
					go func() {
						model.Provider.ChatCompletionStream(
							session.Messages[:len(session.Messages)-1],
							sysPrompt,
							model,
							GetTools(),
							modelResponseCh,
						)
						modelDoneCh <- true
					}()
				}

				return
			case <-time.After(600 * time.Second):
				return
			}
		}

	} else {
		log.E("Model not found")
		streamDoneCh <- false
	}
}

func DynamicAgentChat(modelID string, query string, sysPrompt string) (response string, err error) {
	logger.BreakOnError()
	response = ""

	model := findModel(modelID)
	if model != nil {
		session := NewTempSession()
		err = session.AddMessage(ai.MessageOriginUser, query)

		message, err := model.Provider.ChatCompletion(
			session.Messages,
			sysPrompt,
			model,
			[]*tools.Tool{},
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
