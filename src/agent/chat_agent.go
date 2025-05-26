package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/mcptools"
	"context"
)

func DirectChatStreaming(ctx context.Context, sessionID string, modelID string, roleID string, query string, streamDoneCh chan bool) {
	model := findModel(modelID)
	if model != nil {
		model.Provider.WaitForAllowance()
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
				case <-modelDoneCh:
					session.Save()
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		session.AddMessage(ai.MessageOriginUser, query, nil)
		err := session.AddMessage(ai.MessageOriginAI, "", nil)
		log.CheckW(err, "Failed to add new message in agent")

		model.Provider.ChatCompletionStream(
			ctx,
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
