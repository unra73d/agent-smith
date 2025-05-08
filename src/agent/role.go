package agent

import (
	"agentsmith/src/logger"
	"database/sql"
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type RoleConfig struct {
	Name               string `json:"name"`
	GeneralInstruction string `json:"generalInstruction"`
	Role               string `json:"role"`
	Style              string `json:"style"`
}

type Role struct {
	ID     string     `json:"id"`
	Config RoleConfig `json:"config"`
}

func LoadRoles() []*Role {
	log.D("Loading roles from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()
	roles := make([]*Role, 0, 32)

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open role db")
	defer db.Close()

	query := "SELECT id, data FROM roles;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select roles from DB")
	defer rows.Close()

	for rows.Next() {
		var role Role
		var dataJSON string

		// Scan the row data into variables
		err = rows.Scan(&role.ID, &dataJSON)
		if err != nil {
			log.W("Failed to scan role row:", err)
			continue
		}

		// Unmarshal the JSON data from the 'data' column into Config
		if dataJSON != "" {
			err = json.Unmarshal([]byte(dataJSON), &role.Config)
			if err != nil {
				log.W("Failed to unmarshal config for role:", role.ID, err)
				role.Config = RoleConfig{}
			}
		} else {
			role.Config = RoleConfig{}
		}

		// Append the successfully loaded role to the slice
		roles = append(roles, &role)
	}

	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		log.E("Error iterating role rows: %v", err)
	}

	log.D("Loaded roles from DB:", len(roles))
	return roles
}

func newRole() *Role {
	role := &Role{
		ID: uuid.NewString(),
		Config: struct {
			Name               string `json:"name"`
			GeneralInstruction string `json:"generalInstruction"`
			Role               string `json:"role"`
			Style              string `json:"style"`
		}{},
	}
	return role
}

func (self *Role) Save() (err error) {
	log.D("Saving role to ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	configJSON, err := json.Marshal(self.Config)
	log.CheckE(err, nil, "Failed to marshal config for role ", self.ID)

	// Use INSERT OR REPLACE (UPSERT) to handle both new and existing roles
	query := `
	INSERT INTO roles (id, data)
	VALUES (?, ?)
	ON CONFLICT(id) DO UPDATE SET
		data=excluded.data;
	`

	_, err = db.Exec(query, self.ID, string(configJSON))
	log.CheckW(err, "Failed to update role DB")

	log.D("Saved role", self.ID)
	return
}

func (self *Role) Delete() {
	log.D("Deleting role from ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	query := "DELETE FROM roles WHERE id=?"
	db.Exec(query, self.ID)
}
