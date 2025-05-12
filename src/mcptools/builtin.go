package mcptools

func GetBuiltinTools() []*Tool {
	res := make([]*Tool, 0, 4)

	tool, err := NewToolFromJSON(`{
		"Name": "builtin_dynamic_ai_agent",
		"Description": "Use this tool to dynamically create AI agent with custom system prompts and subset of the tools you have. Use the combination of the system prompt, query and available tools to obtain the result from AI model. For example you can instruct model to analyze query and create a task execution plan for you. Or you can instruct model to validate or classify something. Use your best judgement tone determine the purpose of this tool and if you can utilize it to achieve your objective.",
		"Required": ["sysPrompt", "query"],
		"Params": [
			{
				"Name": "sysPrompt",
				"Type": "string",
				"Description": "System prompt that AI agent will set to model. Include here any context information you want AI model to be aware of. Model will not have any other context outside of this."
			},
			{
				"Name": "query",
				"Type": "string",
				"Description": "Query text to be sent to the AI agent."
			},
			{
				"Name": "tools",
				"Type": "array",
				"Description": "Array of tool names that can be used by this agent. Specify here any tools from the ones you have that you think the AI agent will need to use.",
			},
		],
	}`)

	if err != nil {
		res = append(res, tool)
	}

	return res
}
