package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/mcptools"
	"time"
)

func DirectChatStreaming(sessionID string, modelID string, roleID string, query string, streamCh chan string, streamDoneCh chan bool) {
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
		modelDoneCh := make(chan bool)
		go func() {
			for {
				select {
				case msg := <-modelResponseCh:
					session.UpdateLastMessage(msg)
					streamCh <- msg
				case <-modelDoneCh:
					session.Save()
					return
				case <-time.After(60 * time.Second):
					return
				}
			}
		}()

		session.AddMessage(ai.MessageOriginUser, query)
		err := session.AddMessage(ai.MessageOriginAI, "")
		log.CheckW(err, "Failed to add new message in agent")

		model.Provider.ChatCompletionStream(
			session.Messages[:len(session.Messages)-1],
			sysPrompt,
			model,
			[]*mcptools.Tool{},
			modelResponseCh,
			nil,
		)
		modelDoneCh <- true
		streamDoneCh <- true
	} else {
		log.E("Model not found")
		streamDoneCh <- false
	}
}
