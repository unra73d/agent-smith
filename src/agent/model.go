package agent

import (
	"agentsmith/src/logger"
	"encoding/json"
	"os"
)

type APIType string

const (
	APITypeOpenAI = "openai"
)

type Model struct {
	Name        string
	APIUrl      string
	APIKey      string
	APIProvider APIType
}

func LoadModels() []Model {
	log.D("Loading models")
	defer logger.BreakOnError()

	models := make([]Model, 0)
	data, err := os.ReadFile(os.Getenv("AS_MODEL_CONFIG_FILE"))
	log.CheckE(err, nil, "Failed to open config file")

	err = json.Unmarshal(data, &models)
	log.CheckE(err, nil, "Failed to parse models json")

	return models
}
