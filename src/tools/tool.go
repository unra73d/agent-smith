package tools

type ToolParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Value       string `json:"value"`
}

type Tool struct {
	Name        string      `json:"name"`
	Params      []ToolParam `json:"params"`
	Description string      `json:"description"`
}
