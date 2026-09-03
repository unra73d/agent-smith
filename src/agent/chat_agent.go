package agent

import (
	"agentsmith/src/ai"
	"agentsmith/src/mcptools"
	"context"
	"time"
)

type streamResult struct {
	completionTokens int
	err              error
	elapsed          time.Duration
}

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
		modelDoneCh := make(chan streamResult)
		statisticsRecordedCh := make(chan struct{})
		go func() {
			for {
				select {
				case msg := <-modelResponseCh:
					session.UpdateLastMessage(msg)
				case result := <-modelDoneCh:
					if result.err == nil {
						outputTokens := result.completionTokens
						if outputTokens <= 0 {
							outputTokens = ai.EstimateOutputTokens(session.Messages[len(session.Messages)-1].Text)
						}
						if err := session.RecordResponseStatistics(outputTokens, result.elapsed); err != nil {
							log.W("Failed to record response statistics:", err)
						}
					} else if !session.temporary {
						if err := session.Save(); err != nil {
							log.W("Failed to save completed session:", err)
						}
					}
					close(statisticsRecordedCh)
					return
				}
			}
		}()

		session.AddMessage(ai.MessageOriginUser, query, nil)
		err := session.AddMessage(ai.MessageOriginAI, "", nil)
		log.CheckW(err, "Failed to add new message in agent")

		started := time.Now()
		completionTokens, err := model.Provider.ChatCompletionStream(
			ctx,
			session.Messages[:len(session.Messages)-1],
			sysPrompt,
			model,
			[]*mcptools.Tool{},
			modelResponseCh,
			nil,
		)
		modelDoneCh <- streamResult{completionTokens: completionTokens, err: err, elapsed: time.Since(started)}
		<-statisticsRecordedCh
		if err != nil {
			streamDoneCh <- false
			return
		}
		session.MaybeGenerateTitle(model)
		streamDoneCh <- true
	} else {
		log.E("Model not found")
		streamDoneCh <- false
	}
}
