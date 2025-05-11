package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/logger"
	"agentsmith/src/tools"
	"errors"
)

func ToolChatStreaming(sessionID string, modelID string, roleID string, query string, streamCh chan string, streamDoneCh chan bool) {

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
