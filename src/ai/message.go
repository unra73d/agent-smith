package ai

import "agentsmith/src/mcptools"

type MessageOrigin string

const (
	MessageOriginUser   = "user"
	MessageOriginAI     = "assistant"
	MessageOriginTool   = "tool"
	MessageOriginSystem = "system"
)

type Message struct {
	ID           string                      `json:"id"`
	Origin       MessageOrigin               `json:"origin"`
	Text         string                      `json:"text"`
	ToolRequests []*mcptools.ToolCallRequest `json:"toolRequests"`
}
