// Package ai implements connectivity with the AI models
package ai

import (
	"agentsmith/src/logger"
	"agentsmith/src/tools"
	"agentsmith/src/util"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tmaxmax/go-sse"
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
	Models() []*Model

	Test() error
	LoadModels() error
	ChatCompletion(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool) (string, error)
	ChatCompletionStream(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool, writeCh chan string) error
}

type APIProvider struct {
	name    string
	apiURL  string
	apiKey  string
	apiType APIType
	models  []*Model
}

func (self *APIProvider) Name() string     { return self.name }
func (self *APIProvider) URL() string      { return self.apiURL }
func (self *APIProvider) APIKey() string   { return self.apiKey }
func (self *APIProvider) Type() APIType    { return self.apiType }
func (self *APIProvider) Models() []*Model { return self.models }

type OpenAIProvider struct {
	APIProvider
}

type GoogleAIProvider struct {
	APIProvider
}

func NewProvider(apiType APIType, name string, url string, apiKey string) (IAPIProvider, error) {
	basicProvider := APIProvider{name, url, apiKey, apiType, make([]*Model, 0, 16)}

	var provider IAPIProvider
	switch apiType {
	case APITypeOpenAI, APITypeLMStudio, APITypeOpenAICompatible, APITypeOllama:
		provider = &OpenAIProvider{basicProvider}
	case APITypeGoogle:
		provider = &GoogleAIProvider{basicProvider}
	}
	err := provider.LoadModels()
	return provider, err
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

	var signal sync.WaitGroup

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

		signal.Add(1)
		go func() {
			defer signal.Done()
			provider, err := NewProvider(
				APIType(providerTypeStr.String),
				name.String,
				apiURL.String,
				apiKey.String,
			)
			if err != nil {
				log.W("Error creating provider '%s' from DB data: %v", name.String, err)
				return
			}
			providers = append(providers, provider)
		}()
	}

	signal.Wait()

	log.D("Loaded providers from DB:", len(providers))
	return providers
}

func (self *OpenAIProvider) Test() error { return nil }

type OpenAIModelListRes struct {
	Data []map[string]any `json:"data"`
}

func (self *OpenAIProvider) LoadModels() error {
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
		return err
	}

	self.models = make([]*Model, len(list.Data))
	for i, config := range list.Data {
		self.models[i] = &Model{
			ID:       config["id"].(string),
			Name:     self.name + ": " + config["id"].(string),
			Provider: self,
		}
	}

	return nil
}

type OpenAIChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatCompletionChoice struct {
	Index        int                         `json:"index"`
	FinishReason string                      `json:"finish_reason"`
	Message      OpenAIChatCompletionMessage `json:"message"`
}

type OpenAIChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatCompletionRes struct {
	ID                string                       `json:"id"`
	Created           int64                        `json:"created"`
	Model             string                       `json:"model"`
	Choices           []OpenAIChatCompletionChoice `json:"choices"`
	Usage             OpenAIChatCompletionUsage    `json:"usage"`
	SystemFingerprint string                       `json:"system_fingerprint"`
}

func (self *OpenAIProvider) ChatCompletion(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool) (string, error) {
	log.D("OpenAI chat completion")
	url := self.apiURL + "/chat/completions"

	c := resty.New()
	defer c.Close()
	r := c.R()

	if self.apiKey != "" && self.apiType != APITypeOllama && self.apiType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.apiKey)
	}

	r.SetBody(map[string]any{
		"model":    model.ID,
		"messages": prepareMessages(messages, sysPrompt),
	})
	res := &OpenAIChatCompletionRes{}
	r.SetResult(res)
	_, err := r.Post(url)

	if err != nil || len(res.Choices) == 0 {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}

type OpenAIStreamChatResponseChoice struct {
	Index int `json:"index"`
	Delta struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason interface{} `json:"finish_reason"`
}

type OpenAIStreamChatResponse struct {
	ID                string                           `json:"id"`
	Object            string                           `json:"object"`
	Created           int                              `json:"created"`
	Model             string                           `json:"model"`
	SystemFingerprint string                           `json:"system_fingerprint"`
	Choices           []OpenAIStreamChatResponseChoice `json:"choices"`
}

func (self *OpenAIProvider) ChatCompletionStream(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool, writeCh chan string) (err error) {
	log.D("OpenAI chat completion streaming")
	log.D("System prompt:", sysPrompt)
	url := self.apiURL + "/chat/completions"

	body := map[string]any{
		"model":    model.ID,
		"messages": prepareMessages(messages, sysPrompt),
		"stream":   true,
		"tools":    prepareTools(tools),
	}
	bodyJSON, err := json.Marshal(body)
	log.CheckW(err, "Failed to pack chat content into json")

	ctx, cancel := context.WithCancel(context.Background())
	r, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJSON))
	r = r.WithContext(ctx)

	if self.apiKey != "" && self.apiType != APITypeOllama && self.apiType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.apiKey)
	}
	r.Header.Add("Content-Type", "application/json")

	conn := sse.NewConnection(r)

	conn.SubscribeMessages(func(e sse.Event) {
		logger.BreakOnError()
		if e.Data != "[DONE]" {
			var response OpenAIStreamChatResponse
			err := json.Unmarshal([]byte(e.Data), &response)
			log.CheckE(err, nil, "Failed to parse JSON response from chat completion_tokens")
			writeCh <- response.Choices[0].Delta.Content
		} else {
			log.D("Provider closing streaming")
			cancel()
		}
	})

	err = conn.Connect()
	log.CheckW(err, "Failed to call completions api")

	return err
}

func (self *GoogleAIProvider) Test() error { return nil }

func (self *GoogleAIProvider) LoadModels() error {
	return nil
}

func (self *GoogleAIProvider) ChatCompletion(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool) (string, error) {
	return "Message received", nil
}

func (self *GoogleAIProvider) ChatCompletionStream(messages []*Message, sysPrompt string, model *Model, tools []*tools.Tool, writeCh chan string) error {
	return nil
}

func prepareMessages(messages []*Message, sysPrompt string) *[]map[string]string {
	bodyMessages := make([]map[string]string, len(messages)+1)
	bodyMessages[0] = map[string]string{
		"role":    "system",
		"content": sysPrompt,
	}
	for i, message := range messages {
		bodyMessages[i+1] = map[string]string{
			"role":    string(message.Origin),
			"content": strings.TrimSpace(util.CutThinking(message.Text)),
		}
	}
	return &bodyMessages
}

func prepareTools(tools []*tools.Tool) *[]map[string]any {
	bodyTools := make([]map[string]any, len(tools))

	for i, tool := range tools {

		paramMap := make(map[string]any)
		for _, param := range tool.Params {
			paramMap[param.Name] = map[string]string{
				"type":        param.Type,
				"description": param.Description,
			}
		}

		bodyTools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters": map[string]any{
					"type":       "object",
					"properties": paramMap,
					"required":   tool.RequiredParams,
				},
			},
		}
	}

	return &bodyTools
}
