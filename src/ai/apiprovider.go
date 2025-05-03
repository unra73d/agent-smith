// Package ai implements connectivity with the AI models
package ai

import (
	"agentsmith/src/logger"
	"database/sql"
	"encoding/json"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"resty.dev/v3"
)

var log = logger.Logger("ai", 1, 1, 1)

type APIType string

const (
	APITypeOpenAI           = "openai"
	APITypeLMStudio         = "lmstudio"
	APITypeGoogle           = "google"
	APITypeMistral          = "mistral"
	APITypeOllama           = "ollama"
	APITypeAnthropic        = "anthropic"
	APITypeOpenAICompatible = "openaicompatible"
)

type IAPIProvider interface {
	Name() string
	URL() string
	APIKey() string
	Type() APIType

	Test() error
	ListModels() ([]Model, error)
	ChatCompletion(messages []Message, model *Model, toolUse bool) (*Message, error)
}

type APIProvider struct {
	name    string
	apiURL  string
	apiKey  string
	apiType APIType
}

func (self *APIProvider) Name() string   { return self.name }
func (self *APIProvider) URL() string    { return self.apiURL }
func (self *APIProvider) APIKey() string { return self.apiKey }
func (self *APIProvider) Type() APIType  { return self.apiType }

type OpenAIProvider struct {
	APIProvider
}

type GoogleAIProvider struct {
	APIProvider
}

func NewProvider(apiType APIType, name string, url string, apiKey string) (IAPIProvider, error) {
	basicProvider := APIProvider{name, url, apiKey, apiType}

	switch apiType {
	case APITypeOpenAI, APITypeLMStudio, APITypeOpenAICompatible, APITypeOllama:
		provider := &OpenAIProvider{basicProvider}
		return provider, provider.Test()
	case APITypeGoogle:
		provider := &GoogleAIProvider{basicProvider}
		return provider, provider.Test()
	}

	return nil, errors.New("unknown provider")
}

func LoadProviders() []IAPIProvider {
	log.D("Loading providers from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	providers := make([]IAPIProvider, 0, 16)

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open agent db for loading providers")
	defer db.Close()

	query := "SELECT name, api_url, api_key, provider FROM providers;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select providers from DB")
	defer rows.Close()

	for rows.Next() {
		var name, apiURL, apiKey, providerTypeStr sql.NullString

		err = rows.Scan(&name, &apiURL, &apiKey, &providerTypeStr)
		if err != nil {
			log.W("Failed to scan provider row:", err)
			continue
		}

		if !name.Valid || !providerTypeStr.Valid {
			log.W("Skipping provider row due to missing name or provider type")
			continue
		}

		provider, err := NewProvider(
			APIType(providerTypeStr.String),
			name.String,
			apiURL.String,
			apiKey.String,
		)
		if err != nil {
			log.W("Error creating provider '%s' from DB data: %v", name.String, err)
			continue
		}
		providers = append(providers, provider)
	}

	// Check for errors during row iteration
	if err = rows.Err(); err != nil {
		log.E("Error iterating provider rows: %v", err)
	}

	log.D("Loaded providers from DB:", len(providers))
	return providers
}

func LoadProvidersFromJSON() []IAPIProvider {
	log.D("loading providers")
	defer logger.BreakOnError()

	data, err := os.ReadFile(os.Getenv("AS_MODEL_CONFIG_FILE"))
	log.CheckE(err, nil, "Failed to open models config file")

	var loadedConfigs []APIProvider
	err = json.Unmarshal(data, &loadedConfigs)
	log.CheckE(err, nil, "Failed to parse models json")

	providers := make([]IAPIProvider, 0, 16)

	for _, config := range loadedConfigs {
		provider, err := NewProvider(config.apiType, config.name, config.apiURL, config.apiKey)
		if err != nil {
			log.E("Error recreating provider", config.name)
			continue
		}
		providers = append(providers, provider)
	}

	return providers
}

func (self *OpenAIProvider) Test() error { return nil }

type OpenAIModelListRes struct {
	Data []map[string]any `json:"data"`
}

func (self *OpenAIProvider) ListModels() ([]Model, error) {
	log.D("Loading OpenAI models")
	url := self.apiURL + "/models"

	c := resty.New()
	defer c.Close()
	r := c.R()

	if self.apiKey != "" && self.apiType != APITypeOllama && self.apiType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.apiKey)
	}

	list := &OpenAIModelListRes{}
	r.SetResult(list)
	_, err := r.Get(url)

	if err != nil {
		return nil, err
	}

	models := make([]Model, len(list.Data))
	for i, model := range list.Data {
		models[i].ID = model["id"].(string)
		models[i].Name = model["id"].(string)
		models[i].Provider = self
	}

	return models, nil
}

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	FinishReason string                `json:"finish_reason"`
	Message      ChatCompletionMessage `json:"message"`
}

type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatCompletionRes struct {
	ID                string                 `json:"id"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             ChatCompletionUsage    `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint"`
}

func (self *OpenAIProvider) ChatCompletion(messages []Message, model *Model, toolUse bool) (*Message, error) {
	log.D("OpenAI chat completion")
	url := self.apiURL + "/chat/completions"

	c := resty.New()
	defer c.Close()
	r := c.R()

	if self.apiKey != "" && self.apiType != APITypeOllama && self.apiType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.apiKey)
	}

	bodyMessages := make([]map[string]string, len(messages))
	for i, message := range messages {
		bodyMessages[i] = map[string]string{
			"role":    string(message.Origin),
			"content": message.Text,
		}
	}

	r.SetBody(map[string]any{
		"model":    model.ID,
		"messages": bodyMessages,
	})
	res := &OpenAIChatCompletionRes{}
	r.SetResult(res)
	_, err := r.Post(url)

	if err != nil || len(res.Choices) == 0 {
		return nil, err
	}

	return &Message{res.ID, MessageOriginAI, res.Choices[0].Message.Content}, nil
}

func (self *GoogleAIProvider) Test() error { return nil }

func (self *GoogleAIProvider) ListModels() ([]Model, error) {
	return []Model{}, nil
}

func (self *GoogleAIProvider) ChatCompletion(messages []Message, model *Model, toolUse bool) (*Message, error) {
	return &Message{"", MessageOriginAI, "Message received"}, nil
}
