package mcptools

import (
	"encoding/json"
)

type ToolParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Value       string `json:"value"`
}

type Tool struct {
	Name           string       `json:"name"`
	Params         []*ToolParam `json:"params"`
	Description    string       `json:"description"`
	Server         *MCPServer   `json:"-"`
	RequiredParams []string     `json:"requiredParams"`
}

func NewToolFromJSON(jsonStr string) (*Tool, error) {
	var tool Tool
	err := json.Unmarshal([]byte(jsonStr), &tool)
	if tool.RequiredParams == nil {
		tool.RequiredParams = make([]string, 0)
	}

	return &tool, err
}
