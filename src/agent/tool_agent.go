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
		} else {
			openToolTag := strings.Index(content, "<tool_call>")
			closeToolTag := strings.LastIndex(content, "</tool_call>")
			if openToolTag != -1 && closeToolTag != -1 && openToolTag < closeToolTag {
				content = content[openToolTag : closeToolTag+1]
				openBracketIndex := strings.Index(content, "{")
				closeBracketIndex := strings.LastIndex(content, "}")
				if openBracketIndex != -1 && closeBracketIndex != -1 && openBracketIndex < closeBracketIndex {
					content = content[openBracketIndex : closeBracketIndex+1]
				}
			}
		}

		var callRequest mcptools.ToolCallRequest
		err := json.Unmarshal([]byte(content), &callRequest)

		if err == nil {
			tools := append(GetTools(), GetBuiltinTools()...)
			for _, tool := range tools {
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
If you need to use a tool, your response must immediately start with the tool JSON wrapped in <tool_call><tool_call> tags.
When to use tools:
For real-time data (e.g., weather, news, search).
For specific calculations or code execution.
To access specialized knowledge bases, storages, databases.
Tool Call Format:
Provide a JSON object like this:
<tool_call>{
	"name": "time",
	"params": {
		"location": "New York",
		"24hr": true
	}
}</tool_call>
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

		modelResponseCh := make(chan string)
		modelDoneCh := make(chan error)
		toolCh := make(chan []*mcptools.ToolCallRequest)

		session.AddMessage(ai.MessageOriginUser, query, nil)
		session.AddMessage(ai.MessageOriginAI, "", nil)

		var toolCalls []*mcptools.ToolCallRequest
		selectedTools := make([]*mcptools.Tool, 0, len(Agent.mcps)*4)
		for _, mcp := range Agent.mcps {
			if mcp.Active {
				selectedTools = append(selectedTools, mcp.Tools...)
			}
		}

		selectedTools = append(selectedTools, GetBuiltinTools()...)

		if len(selectedTools) > 0 {
			sysPrompt = sysPrompt + toolUsePrompt
		}

		chatCompletion := func() {
			model.Provider.WaitForAllowance()
			err := model.Provider.ChatCompletionStream(
				ctx,
				session.Messages[:len(session.Messages)-1],
				sysPrompt,
				model,
				selectedTools,
				modelResponseCh,
				toolCh,
			)
			modelDoneCh <- err
		}

		go chatCompletion()

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
						mcp = GetMCPForTool(callRequest.Name)
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
					if mcp != nil || callRequest.Name == "lua_code_runner" {
						session.Messages[len(session.Messages)-1].ToolRequests = []*mcptools.ToolCallRequest{callRequest}
						session.UpdateLastMessage("")

						var toolResult string
						if callRequest.Name == "lua_code_runner" {
							toolResult = mcptools.RunLua(callRequest)
						} else {
							toolResult, err = mcp.CallTool(callRequest)
						}
						log.D("Tool execution result: ", toolResult)

						if err != nil {
							log.E("Error during tool call")
							session.AddMessage(ai.MessageOriginTool, "Tool returned an error", []*mcptools.ToolCallRequest{callRequest})
							streamDoneCh <- false
							return
						}
						session.AddMessage(ai.MessageOriginTool, toolResult, []*mcptools.ToolCallRequest{callRequest})
						session.AddMessage(ai.MessageOriginAI, "", nil)

						toolCalls = nil
						go chatCompletion()
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

func GetMCPForTool(name string) (mcp *mcptools.MCPServer) {
	for _, tool := range GetTools() {
		if tool.Name == name {
			mcp = tool.Server
			break
		}
	}
	return
}
