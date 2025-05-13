package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/mcptools"
	"agentsmith/src/util"
	"context"
	"encoding/json"
	"errors"
	"strings"
)

type AgentAction string

const (
	AgentActionAnswer   = "answer"
	AgentActionToolCall = "tool_call"
	AgentActionError    = "error"
)

func inferNextAction(message string) (AgentAction, *mcptools.MCPServer, *mcptools.ToolCallRequest) {
	content := util.CutThinking(message)
	if len(content) > 0 {
		content = strings.TrimSpace(content)

		// some models wrap text into markdown, try to strip it
		if strings.HasPrefix(content, "```") {
			openBracketIndex := strings.Index(content, "{")
			closeBracketIndex := strings.LastIndex(content, "}")

			if openBracketIndex != -1 && closeBracketIndex != -1 && openBracketIndex < closeBracketIndex {
				content = content[openBracketIndex : closeBracketIndex+1]
			}
		}

		var callRequest mcptools.ToolCallRequest
		err := json.Unmarshal([]byte(content), &callRequest)

		if err == nil {
			for _, tool := range GetTools() {
				if strings.Contains(callRequest.Name, tool.Name) {
					callRequest.Name = tool.Name
					return AgentActionToolCall, tool.Server, &callRequest
				}
			}
		}
		return AgentActionAnswer, nil, nil
	}
	return AgentActionError, nil, nil
}

const toolUsePrompt = `
## Tool usage: \n
You have tools/functions (which are two names for same term) to find information or perform actions.
You can use one tool per message and will receive the result of that tool use in the response.
If you need to use a tool, your response must immediately start with the JSON for the tool.
When to use tools:
For real-time data (e.g., weather, news, search).
For specific calculations or code execution.
To access specialized knowledge bases, storages, databases.
Tool Call Format:
Provide a JSON object like this:
{
	"name": "time",
	"params": {
		"location": "New York",
		"24hr": true
	}
}
Only use provided tool names and their defined parameters.
If you can answer directly from your knowledge, do so without using a tool.
`

func ToolChatStreaming(ctx context.Context, sessionID string, modelID string, roleID string, query string, streamDoneCh chan bool) {
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
		modelDoneCh := make(chan error)
		toolCh := make(chan []*mcptools.ToolCallRequest)

		session.AddMessage(ai.MessageOriginUser, query, nil)
		session.AddMessage(ai.MessageOriginAI, "", nil)

		var toolCalls []*mcptools.ToolCallRequest
		go func() {
			log.D("Starting initial chat completion")
			err := model.Provider.ChatCompletionStream(
				ctx,
				session.Messages[:len(session.Messages)-1],
				sysPrompt,
				model,
				GetTools(),
				modelResponseCh,
				toolCh,
			)
			modelDoneCh <- err
		}()

		for {
			select {
			case msg := <-modelResponseCh:
				session.UpdateLastMessage(msg)
			case toolCalls = <-toolCh:
			case err := <-modelDoneCh:
				log.D("Model response done")
				session.Save()

				var action AgentAction
				var mcp *mcptools.MCPServer
				var callRequest *mcptools.ToolCallRequest

				if err != nil {
					action = AgentActionError
				} else {
					if len(toolCalls) > 0 {
						action = AgentActionToolCall
						callRequest = toolCalls[0]
						for _, tool := range GetTools() {
							if tool.Name == toolCalls[0].Name {
								mcp = tool.Server
								break
							}
						}
					} else {
						action, mcp, callRequest = inferNextAction(session.Messages[len(session.Messages)-1].Text)
					}
				}

				switch action {
				case AgentActionError:
					log.E("Error during parse of model intention")
					streamDoneCh <- false
					return
				case AgentActionAnswer:
					log.D("Model will answer ")
					streamDoneCh <- true
					return
				case AgentActionToolCall:
					log.D("Model will call tool")
					if mcp != nil {
						toolResult, _ := mcp.CallTool(callRequest)
						log.D("Tool execution result: ", toolResult)

						session.Messages[len(session.Messages)-1].ToolRequests = []*mcptools.ToolCallRequest{callRequest}
						session.AddMessage(ai.MessageOriginTool, toolResult, []*mcptools.ToolCallRequest{callRequest})
						session.AddMessage(ai.MessageOriginAI, "", nil)

						toolCalls = nil
						go func() {
							err := model.Provider.ChatCompletionStream(
								ctx,
								session.Messages[:len(session.Messages)-1],
								sysPrompt,
								model,
								GetTools(),
								modelResponseCh,
								toolCh,
							)
							modelDoneCh <- err
						}()
					} else {
						log.E("didnt find mcp to call")
						streamDoneCh <- false
						return
					}
				}
			case <-ctx.Done():
				streamDoneCh <- false
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
		err = session.AddMessage(ai.MessageOriginUser, query, nil)

		message, err := model.Provider.ChatCompletion(
			session.Messages,
			sysPrompt,
			model,
			[]*mcptools.Tool{},
		)
		log.CheckE(err, nil, "Failed to get completion for message")

		err = session.AddMessage(ai.MessageOriginAI, message, nil)
		log.CheckE(err, nil, "Failed to store new message in agent")

		response = message
	} else {
		err = errors.New("model not selected")
	}
	return
}
