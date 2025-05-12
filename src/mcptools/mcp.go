// Package tools implements features to work with external functions, tools, mcp servers
package mcptools

import (
	"agentsmith/src/logger"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

var log = logger.Logger("tools", 1, 1, 1)

type MCPTransport string

const (
	MCPTransportStdio = "stdio"
	MCPTransportSSE   = "sse"
)

type ToolCallRequest struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

type MCPServer struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Transport MCPTransport `json:"transport"`
	URL       string       `json:"url"`
	Command   string       `json:"command"`
	Args      []string     `json:"args"`
	Tools     []*Tool      `json:"tools"`
}

func LoadMCPServers() []*MCPServer {
	log.D("Loading MCP servers from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()
	mcpServers := make([]*MCPServer, 0, 8)

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open MCP server db")
	defer db.Close()

	query := "SELECT id, name, transport, url, command, args FROM mcp;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select MCP servers from DB")
	defer rows.Close()

	var signal sync.WaitGroup

	for rows.Next() {
		var mcpServer MCPServer
		var argsJSON sql.NullString
		var url sql.NullString
		var command sql.NullString

		err = rows.Scan(&mcpServer.ID, &mcpServer.Name, &mcpServer.Transport, &url, &command, &argsJSON)
		if err != nil {
			log.W("Failed to scan MCP server row:", err)
			continue
		}

		// Assign values from sql.NullString if they are valid
		if url.Valid {
			mcpServer.URL = url.String
		} else {
			mcpServer.URL = ""
		}

		if command.Valid {
			mcpServer.Command = command.String
		} else {
			mcpServer.Command = ""
		}

		signal.Add(1)
		go func() {
			defer signal.Done()
			// Unmarshal the JSON data from the 'args' column into Args
			if argsJSON.Valid && argsJSON.String != "" {
				err = json.Unmarshal([]byte(argsJSON.String), &mcpServer.Args)
				if err != nil {
					log.W("Failed to unmarshal args for MCP server:", mcpServer.ID, err)
					mcpServer.Args = make([]string, 0)
				}
			} else {
				mcpServer.Args = make([]string, 0)
			}

			mcpServer.LoadTools()

			// Append the successfully loaded MCP server to the slice
			mcpServers = append(mcpServers, &mcpServer)
		}()
	}

	signal.Wait()

	log.D("Loaded MCP servers from DB:", len(mcpServers))
	return mcpServers
}

func (self *MCPServer) Save() (err error) {
	log.D("Saving MCP server to ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	argsJSON, err := json.Marshal(self.Args)
	log.CheckE(err, nil, "Failed to marshal args for MCP server ", self.ID)

	// Use INSERT OR REPLACE (UPSERT) to handle both new and existing MCP servers
	query := `
	INSERT INTO mcp (id, name, transport, url, command, args)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name=excluded.name,
		transport=excluded.transport,
		url=excluded.url,
		command=excluded.command,
		args=excluded.args;
	`

	_, err = db.Exec(query, self.ID, self.Name, self.Transport, self.URL, self.Command, string(argsJSON))
	log.CheckW(err, "Failed to update MCP server DB")

	log.D("Saved MCP server", self.ID)
	return
}

func (self *MCPServer) Delete() {
	log.D("Deleting MCP server from ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	query := "DELETE FROM mcp WHERE id=?"
	db.Exec(query, self.ID)
}

func (self *MCPServer) connect() (context.Context, context.CancelFunc, *client.Client) {
	defer logger.BreakOnError()

	c, err := client.NewStdioMCPClient(
		self.Command,
		[]string{}, // Empty ENV
		self.Args...,
	)
	log.CheckE(err, nil, "Failed to create MCP client")

	// Create context with timeout
	ctx, cancel := context.WithCancel(context.Background())

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "Agent Smith MCP client",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initRequest)
	log.CheckE(err, nil, "Failed to initialize MCP request")

	return ctx, cancel, c
}

func (self *MCPServer) LoadTools() {
	defer logger.BreakOnError()

	ctx, cancel, c := self.connect()
	defer cancel()

	toolsRequest := mcp.ListToolsRequest{}
	mcpTools, err := c.ListTools(ctx, toolsRequest)
	log.CheckE(err, nil, "Failed to list tools from MCP: ", self.Name)

	self.Tools = make([]*Tool, 0, len(mcpTools.Tools))
	for _, tool := range mcpTools.Tools {
		params := make([]*ToolParam, 0, 8)
		for name, prop := range tool.InputSchema.Properties {
			// Ensure prop is a map[string]interface{}
			if propMap, ok := prop.(map[string]interface{}); ok {
				param := &ToolParam{
					Name:        name,
					Type:        propMap["type"].(string),
					Description: propMap["description"].(string),
				}
				params = append(params, param)

			} else {
				log.W("Invalid property format for:", name)
			}
		}
		self.Tools = append(self.Tools, &Tool{
			Name:           tool.Name,
			Description:    tool.Description,
			Params:         params,
			RequiredParams: tool.InputSchema.Required,
			Server:         self,
		})
	}
}

func (self *MCPServer) CallTool(callRequest *ToolCallRequest) (result string, err error) {
	defer logger.BreakOnError()

	ctx, cancel, c := self.connect()
	defer cancel()

	toolRequest := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}

	toolRequest.Params.Name = callRequest.Name
	toolRequest.Params.Arguments = callRequest.Params

	callResult, err := c.CallTool(ctx, toolRequest)
	log.CheckE(err, nil, "Failed to call MCP tool: ", callRequest.Name)

	for _, content := range callResult.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			result = textContent.Text
		} else {
			jsonBytes, _ := json.MarshalIndent(content, "", "  ")
			result = string(jsonBytes)
		}
	}

	return
}
