package server

import (
	"agentsmith/src/logger"
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func InitDebugRoutes(router *gin.Engine, server *http.Server) {
	router.GET("/debug/quit", func(c *gin.Context) {
		stopServer(server)
	})
	router.GET("/debug/initdb", initDB)
}

func initDB(c *gin.Context) {
	log.D("Initializing sqlite DB")
	defer logger.BreakOnError()

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, func() { c.Status(500) }, "Cant open DB")

	defer db.Close()

	// Create the tables
	// Create the sessions table
	createSessionsTableSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		date DATETIME,
		summary TEXT,
		data TEXT
	);`

	// Create the AI providers table
	createAIProvidersTableSQL := `
	CREATE TABLE IF NOT EXISTS providers (
		name TEXT PRIMARY KEY,
		api_url TEXT,
		api_key TEXT,
		provider TEXT,
		additional_params TEXT
	);`

	// Create the roles table
	createRolesTableSQL := `
	CREATE TABLE IF NOT EXISTS roles (
		id TEXT PRIMARY KEY,
		data TEXT
	);`

	// Execute the SQL statements to create the tables
	_, err = db.Exec(createSessionsTableSQL)
	log.CheckW(err, "Failed to create sessions table")

	_, err = db.Exec(createAIProvidersTableSQL)
	log.CheckW(err, "Failed to create AI providers table")

	_, err = db.Exec(createRolesTableSQL)
	log.CheckW(err, "Failed to create roles table")

	log.D("SQLite DB initialized")
	c.JSON(200, map[string]string{"error": ""})
}

func stopServer(server *http.Server) {
	log.D("Stopping server")
	defer logger.BreakOnError()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	err := server.Shutdown(ctx)
	log.CheckE(err, nil, "Server shutdown failed")
	log.D("Server stopped")
}
