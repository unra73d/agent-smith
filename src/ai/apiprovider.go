// Package ai implements connectivity with the AI models
package ai

import (
	"agentsmith/src/logger"
	"agentsmith/src/mcptools"
	"agentsmith/src/util"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

type APIProvider struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	APIURL      string       `json:"url"`
	APIKey      string       `json:"apiKey"`
	APIType     APIType      `json:"type"`
	RateLimit   int          `json:"rateLimit"`
	Models      []*Model     `json:"models"`
	rateLimiter *rateLimiter `json:"-"`
}

type rateLimiter struct {
	mu         sync.Mutex
	timestamps []time.Time
}

func (self *APIProvider) WaitForAllowance() {
	if self.RateLimit == 0 {
		return
	}
	if self.rateLimiter == nil {
		self.rateLimiter = &rateLimiter{}
	}
	rl := self.rateLimiter
	for {
		rl.mu.Lock()
		now := time.Now()
		oneMinuteAgo := now.Add(-1 * time.Minute)
		// Remove timestamps older than 1 minute
		i := 0
		for ; i < len(rl.timestamps); i++ {
			if rl.timestamps[i].After(oneMinuteAgo) {
				break
			}
		}
		rl.timestamps = rl.timestamps[i:]
		if len(rl.timestamps) < self.RateLimit {
			rl.timestamps = append(rl.timestamps, now)
			rl.mu.Unlock()
			return
		}
		// Wait until the oldest timestamp is out of the window
		wait := rl.timestamps[0].Add(time.Minute).Sub(now)
		rl.mu.Unlock()
		if wait < time.Millisecond*100 {
			wait = time.Millisecond * 100
		}
		time.Sleep(wait)
	}
}

func NewProvider(id string, apiType APIType, name string, url string, apiKey string, rateLimit int) (provider *APIProvider, err error) {
	provider = &APIProvider{id, name, url, apiKey, apiType, rateLimit, make([]*Model, 0, 16), nil}
	err = provider.LoadModels()
	return
}

func LoadProviders() []*APIProvider {
	log.D("Loading providers from", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	providers := make([]*APIProvider, 0, 16)

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open agent db for loading providers")
	defer db.Close()

	query := "SELECT id, name, api_url, api_key, provider, rate_limit FROM providers;"
	rows, err := db.Query(query)
	log.CheckE(err, nil, "Failed to select providers from DB")
	defer rows.Close()

	var signal sync.WaitGroup

	for rows.Next() {
		var id, name, apiURL, apiKey, providerTypeStr sql.NullString
		var rateLimit sql.NullInt16

		err = rows.Scan(&id, &name, &apiURL, &apiKey, &providerTypeStr, &rateLimit)
		if err != nil {
			log.W("Failed to scan provider row:", err)
			continue
		}

		if !id.Valid || !name.Valid || !providerTypeStr.Valid {
			log.W("Skipping provider row due to missing name or provider type")
			continue
		}

		signal.Add(1)
		go func() {
			defer signal.Done()
			provider, err := NewProvider(
				id.String,
				APIType(providerTypeStr.String),
				name.String,
				apiURL.String,
				apiKey.String,
				int(rateLimit.Int16),
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

func (self *APIProvider) Save() (err error) {
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	query := `
	INSERT INTO providers (id, name, api_url, api_key, provider, rate_limit)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name=excluded.name,
		api_url=excluded.api_url,
		api_key=excluded.api_key,
		provider=excluded.provider,
		rate_limit=excluded.rate_limit;
	`

	_, err = db.Exec(query, self.ID, self.Name, self.APIURL, self.APIKey, self.APIType, self.RateLimit)
	log.CheckW(err, "Failed to update provider DB")

	log.D("Saved provider", self.Name)
	return
}

func (self *APIProvider) Delete() (err error) {
	log.D("Deleting provider from ", os.Getenv("AS_AGENT_DB_FILE"))
	defer logger.BreakOnError()

	db, err := sql.Open("sqlite3", os.Getenv("AS_AGENT_DB_FILE"))
	log.CheckE(err, nil, "Failed to open DB")
	defer db.Close()

	query := "DELETE FROM providers WHERE id=?"
	_, err = db.Exec(query, self.ID)
	log.CheckW(err, "Failed to delete provider from DB")

	return
}

func (self *APIProvider) Test() bool {
	return self.LoadModels() == nil
}

type OpenAIModelListRes struct {
	Data  []map[string]any `json:"data"`
	Error string           `json:"error"`
}

func (self *APIProvider) LoadModels() (err error) {
	defer logger.BreakOnError()
	log.D("Loading OpenAI models")
	url := self.APIURL + "/models"

	c := resty.New()
	defer c.Close()
	r := c.R()

	if self.APIKey != "" && self.APIType != APITypeOllama && self.APIType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.APIKey)
	}

	list := &OpenAIModelListRes{}
	r.SetResult(list)
	r.SetTimeout(10 * time.Second)
	_, err = r.Get(url)
	log.CheckE(err, nil, "failed to list models for provider: ", self.Name)
	if list.Error != "" {
		return errors.New("bad api call")
	}

	self.Models = make([]*Model, len(list.Data))
	for i, config := range list.Data {
		self.Models[i] = &Model{
			ID:       config["id"].(string),
			Name:     config["id"].(string),
			Provider: self,
		}
	}

	log.D("loaded", len(self.Models), " models for provider: ", self.Name)
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

func (self *APIProvider) ChatCompletion(messages []*Message, sysPrompt string, model *Model, tools []*mcptools.Tool) (string, error) {
	log.D("OpenAI chat completion")
	url := self.APIURL + "/chat/completions"

	c := resty.New()
	defer c.Close()
	r := c.R()

	if self.APIKey != "" && self.APIType != APITypeOllama && self.APIType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.APIKey)
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

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIToolCall struct {
	Index    int                `json:"-"`
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

type OpenAIFunctionCallChunk struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type OpenAIToolCallChunk struct {
	Index    int                     `json:"index"`
	ID       string                  `json:"id,omitempty"`
	Type     string                  `json:"type,omitempty"`
	Function OpenAIFunctionCallChunk `json:"function,omitempty"`
}

type OpenAIDelta struct {
	Role      string                `json:"role,omitempty"`
	Content   string                `json:"content,omitempty"`
	ToolCalls []OpenAIToolCallChunk `json:"tool_calls,omitempty"`
}

type OpenAIStreamChatResponseChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason *string     `json:"finish_reason"`
}

type OpenAIStreamChatResponse struct {
	ID                string                           `json:"id"`
	Object            string                           `json:"object"`
	Created           int                              `json:"created"`
	Model             string                           `json:"model"`
	SystemFingerprint string                           `json:"system_fingerprint"`
	Choices           []OpenAIStreamChatResponseChoice `json:"choices"`
}

func (self *APIProvider) ChatCompletionStream(
	ctx context.Context,
	messages []*Message,
	sysPrompt string,
	model *Model,
	tools []*mcptools.Tool,
	writeCh chan string,
	toolCh chan []*mcptools.ToolCallRequest,
) (err error) {
	defer logger.BreakOnError()
	log.D("OpenAI chat completion streaming")

	// log.D("System prompt:", sysPrompt)
	url := self.APIURL + "/chat/completions"

	body := map[string]any{
		"model":    model.ID,
		"messages": prepareMessages(messages, sysPrompt),
		"stream":   true,
	}

	if len(tools) > 0 {
		body["tools"] = prepareTools(tools)
	}

	var bodyJSON []byte
	bodyJSON, err = json.Marshal(body)
	log.CheckE(err, nil, "failed to marshal request body")

	// log.D(string(bodyJSON))

	apiCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	r, err := http.NewRequestWithContext(apiCtx, http.MethodPost, url, bytes.NewBuffer(bodyJSON))
	log.CheckE(err, nil, "failed to create request")

	if self.APIKey != "" && self.APIType != APITypeOllama && self.APIType != APITypeLMStudio {
		r.Header.Add("Authorization", "Bearer "+self.APIKey)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "text/event-stream")
	r.Header.Add("Cache-Control", "no-cache")
	r.Header.Add("Connection", "keep-alive")

	var toolCallsMutex sync.Mutex
	toolCallBuilders := make(map[int]map[string]any)

	conn := sse.NewConnection(r)

	conn.SubscribeMessages(func(e sse.Event) {
		defer logger.BreakOnError()
		eventData := string(e.Data)
		// log.D("Received SSE Data:", eventData) // Log raw data for debugging

		if eventData == "[DONE]" {
			log.D("Provider closing streaming ([DONE] received)")
			cancel()
			return
		}

		var response OpenAIStreamChatResponse
		err := json.Unmarshal([]byte(eventData), &response)
		log.CheckE(err, nil, "failed to parse OpenAI JSON chunk")

		if len(response.Choices) == 0 {
			log.D("Received chunk with no choices, skipping.")
			return
		}

		choice := response.Choices[0]
		delta := choice.Delta

		// 1. Send Text Content
		if delta.Content != "" {
			// log.D("Sending content chunk:", delta.Content) // Debug log
			select {
			case writeCh <- delta.Content:
			case <-ctx.Done():
				log.W("Context cancelled, could not send content chunk")
				return
			}
		}

		// 2. Accumulate Tool Calls
		if len(delta.ToolCalls) > 0 {
			toolCallsMutex.Lock()
			for _, toolChunk := range delta.ToolCalls {
				index := toolChunk.Index

				_, exists := toolCallBuilders[index]
				if !exists {
					toolCallBuilders[index] = map[string]any{
						"id":     "",
						"name":   "",
						"params": "",
					}
				}

				// Update fields (only update if not empty in the chunk)
				if toolChunk.ID != "" {
					toolCallBuilders[index]["id"] = toolCallBuilders[index]["id"].(string) + toolChunk.ID
				}
				if toolChunk.Function.Name != "" {
					toolCallBuilders[index]["name"] = toolCallBuilders[index]["name"].(string) + toolChunk.Function.Name
				}
				if toolChunk.Function.Arguments != "" {
					toolCallBuilders[index]["params"] = toolCallBuilders[index]["params"].(string) + toolChunk.Function.Arguments
				}
			}
			toolCallsMutex.Unlock()
		}

		// 3. Check Finish Reason (optional, for logging or early exit)
		if choice.FinishReason != nil {
			log.D("Stream finished with reason:", *choice.FinishReason)
		}
	})

	log.D("Connecting to SSE stream...")
	err = conn.Connect()

	// Check connection error type
	if err != nil {
		// SSE library might return specific errors on context cancellation or normal closure
		// Check if the error is due to context cancellation (which is expected on [DONE])
		if err == context.Canceled {
			log.D("SSE connection closed gracefully by context cancellation.")
			err = nil
		} else {
			log.E("SSE connection error:", err)
		}
	} else {
		log.D("SSE connection closed without error.")
	}

	toolRequests := make([]*mcptools.ToolCallRequest, len(toolCallBuilders))
	for i, toolCall := range toolCallBuilders {
		var params map[string]any
		var rawParams string
		paramKeys := []string{"params", "arguments", "args"}
		for _, key := range paramKeys {
			if p, ok := toolCall[key].(string); ok && p != "" {
				rawParams = p
				break
			}
		}

		err = json.Unmarshal([]byte(rawParams), &params)
		if err != nil {
			log.E("failed to parse tool params json: %v", err)
		}

		toolID := toolCall["id"].(string)
		if len(toolID) == 0 {
			toolID = uuid.NewString()
		}
		toolRequests[i] = &mcptools.ToolCallRequest{
			ID:     toolID,
			Name:   toolCall["name"].(string),
			Params: params,
		}
	}

	if toolCh != nil && len(toolRequests) > 0 {
		toolCh <- toolRequests
	}

	log.D("Finished processing stream. Accumulated tool calls:", len(toolRequests))
	return
}

func prepareMessages(messages []*Message, sysPrompt string) *[]map[string]any {
	bodyMessages := make([]map[string]any, len(messages)+1)
	bodyMessages[0] = map[string]any{
		"role":    "system",
		"content": sysPrompt,
	}
	for i, message := range messages {
		bodyMessages[i+1] = map[string]any{
			"role":    string(message.Origin),
			"content": "<no response>",
		}
		content := strings.TrimSpace(util.CutThinking(message.Text))
		if len(content) > 0 {
			bodyMessages[i+1]["content"] = content
		}

		if message.Origin == MessageOriginAI && len(message.ToolRequests) > 0 {
			paramJSON, _ := json.Marshal(message.ToolRequests[0].Params)
			toolCalls := []map[string]any{}
			toolCalls = append(toolCalls, map[string]any{
				"id":   message.ToolRequests[0].ID,
				"type": "function",
				"function": map[string]string{
					"name":      message.ToolRequests[0].Name,
					"arguments": string(paramJSON),
				},
			})
			bodyMessages[i+1]["tool_calls"] = toolCalls
		} else if message.Origin == MessageOriginTool && len(message.ToolRequests) > 0 {
			bodyMessages[i+1]["name"] = message.ToolRequests[0].Name
			bodyMessages[i+1]["tool_call_id"] = message.ToolRequests[0].ID
		}
	}
	return &bodyMessages
}

func prepareTools(tools []*mcptools.Tool) *[]map[string]any {
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
