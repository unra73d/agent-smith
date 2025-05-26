package mcptools

import (
	"bytes"
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func GetBuiltinTools() []*Tool {
	res := make([]*Tool, 0, 4)

	// tool, err := NewToolFromJSON(`{
	// 	"Name": "builtin_dynamic_ai_agent",
	// 	"Description": "Use this tool to dynamically create AI agent with custom system prompts and subset of the tools you have. Use the combination of the system prompt, query and available tools to obtain the result from AI model. For example you can instruct model to analyze query and create a task execution plan for you. Or you can instruct model to validate or classify something. Use your best judgement tone determine the purpose of this tool and if you can utilize it to achieve your objective.",
	// 	"Required": ["sysPrompt", "query"],
	// 	"Params": [
	// 		{
	// 			"Name": "sysPrompt",
	// 			"Type": "string",
	// 			"Description": "System prompt that AI agent will set to model. Include here any context information you want AI model to be aware of. Model will not have any other context outside of this."
	// 		},
	// 		{
	// 			"Name": "query",
	// 			"Type": "string",
	// 			"Description": "Query text to be sent to the AI agent."
	// 		},
	// 		{
	// 			"Name": "tools",
	// 			"Type": "array",
	// 			"Description": "Array of tool names that can be used by this agent. Specify here any tools from the ones you have that you think the AI agent will need to use.",
	// 		},
	// 	],
	// }`)

	tool, err := NewToolFromJSON(`{
		"Name": "lua_code_runner",
		"Description": "Execute lua code and get the result. Use this tool when you need to perform any math or calculations. Any print with print() will be accumulated and returned as result. Last value on stack will be appended to the result. Don't write complex functions, write direct code and finish it with return statement to get result.",
		"Required": ["code"],
		"Params": [
			{
				"Name": "code",
				"Type": "string",
				"Description": "Lua5.1 code to be executed."
			}
		]
	}`)

	if err == nil {
		res = append(res, tool)
	}

	return res
}

func RunLua(callRequest *ToolCallRequest) (toolResult string) {
	L := lua.NewState()
	defer L.Close() // Ensure Lua state is closed even on error

	var stdoutBuf bytes.Buffer // Buffer to capture print output

	// Override the Lua print function to write to our buffer
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			stdoutBuf.WriteString(L.ToStringMeta(L.Get(i)).String())
			if i < top {
				stdoutBuf.WriteString("\t") // Separate arguments with tabs
			}
		}
		stdoutBuf.WriteString("\n") // Add newline at the end
		return 0                    // No return values
	}))

	codeStr := callRequest.Params["code"].(string)

	err := L.DoString(codeStr)
	if err == nil {
		// Capture print output
		capturedOutput := stdoutBuf.String()

		// Capture the last returned value
		lv := L.Get(-1)
		// returnedValue := ""
		// if str, ok := lv.(lua.LString); ok && str != "" {
		// 	returnedValue = string(str)
		// }
		goValue := luaValueToGoInterface(lv)
		resultMap := map[string]interface{}{
			"result": goValue,
		}
		jsonOutput, err := json.MarshalIndent(resultMap, "", "  ") // Using MarshalIndent for pretty print
		log.CheckW(err, "error marshalling tool result to json")
		returnedValue := string(jsonOutput)

		// Combine results: prioritize print output, then add returned value if available
		if capturedOutput != "" {
			toolResult = capturedOutput
			if returnedValue != "" {
				toolResult += "\nLast returned value: " + returnedValue
			}
		} else if returnedValue != "" {
			toolResult = returnedValue
		}
		// If neither print output nor returned value, toolResult remains an empty string.
	}
	return
}

func luaValueToGoInterface(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		// Check if it's an array-like table or a map-like table
		// Gopher-lua uses 1-based indexing for arrays internally.
		// If Len() > 0 and RawGetInt(1) exists, it's likely an array.
		// However, a more robust way is to check if all keys are sequential integers starting from 1.
		// For simplicity, we'll try to determine based on `Len()` and iterating.

		// First, check if it's primarily an array
		isSequential := true
		if v.Len() > 0 { // Len() gives count of numerical keys starting from 1
			for i := 1; i <= v.Len(); i++ {
				if v.RawGetInt(i) == lua.LNil {
					isSequential = false
					break
				}
			}
		} else { // if Len() is 0, it might still be a map or an empty array
			isSequential = false
			// We need to iterate non-numeric keys
			v.ForEach(func(key, _ lua.LValue) {
				if key.Type() != lua.LTNumber {
					isSequential = false
				}
			})
		}

		if isSequential && v.Len() > 0 { // It's an array
			arr := make([]interface{}, 0, v.Len())
			for i := 1; i <= v.Len(); i++ {
				arr = append(arr, luaValueToGoInterface(v.RawGetInt(i)))
			}
			return arr
		} else { // It's a map (or mixed/empty table)
			m := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				m[key.String()] = luaValueToGoInterface(value)
			})
			return m
		}
	case *lua.LFunction, *lua.LUserData, *lua.LChannel, *lua.LState:
		// These types typically cannot be marshaled directly to JSON
		return fmt.Sprintf("unsupported_type: %T", v)
	default:
		return nil // Fallback for any other unexpected type
	}
}
