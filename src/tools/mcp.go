// Package tools implements features to work with external functions, tools, mcp servers
package tools

import (
	"agentsmith/src/logger"
	"database/sql"
	"encoding/json"
	"os"
)

var log = logger.Logger("tools", 1, 1, 1)

type MCPTransport string

const (
	MCPTransportStdio = "stdio"
	MCPTransportSSE   = "sse"
)

type MCPServer struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Transport MCPTransport `json:"transport"`
	URL       string       `json:"url"`
	Command   string       `json:"command"`
	Args      []string     `json:"args"`
	Tools     []Tool       `json:"tools"`
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

		// Append the successfully loaded MCP server to the slice
		mcpServers = append(mcpServers, &mcpServer)
	}

	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		log.E("Error iterating MCP server rows: %v", err)
	}

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
