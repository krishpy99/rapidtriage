package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Default configuration values for Claude
const (
	defaultClaudeEndpoint    = "https://api.anthropic.com/v1/messages"
	defaultClaudeModel       = "claude-3-opus-20240229"
	defaultClaudeMaxTokens   = 4096
	defaultClaudeTimeout     = 60 // seconds
	defaultClaudeTemperature = 0.7
)

// Supported Claude models
const (
	ClaudeOpus   = "claude-3-opus-20240229"
	ClaudeSonnet = "claude-3-sonnet-20240229"
	ClaudeHaiku  = "claude-3-haiku-20240307"
	Claude2      = "claude-2.1"
)

// ClaudeModel represents an implementation of the Model interface for Anthropic's Claude API
type ClaudeModel struct {
	config    ModelConfig
	client    *http.Client
	modelName string
}

// Register the Claude model factory
func init() {
	RegisterModel(ModelClaude, NewClaudeModel)
}

// NewClaudeModel creates a new instance of the Claude model
func NewClaudeModel(config ModelConfig) (Model, error) {
	// Set default values if not provided
	if config.Endpoint == "" {
		config.Endpoint = defaultClaudeEndpoint
	}

	if config.ModelName == "" {
		config.ModelName = defaultClaudeModel
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = defaultClaudeMaxTokens
	}

	if config.Timeout == 0 {
		config.Timeout = defaultClaudeTimeout
	}

	if config.Temperature == 0 {
		config.Temperature = defaultClaudeTemperature
	}

	// Validate configuration
	if config.APIKey == "" {
		return nil, ErrInvalidConfiguration
	}

	// Create HTTP client with appropriate timeouts
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return &ClaudeModel{
		config:    config,
		client:    client,
		modelName: config.ModelName,
	}, nil
}

// Name returns the name of the model
func (m *ClaudeModel) Name() string {
	return m.modelName
}

// Type returns the type of model
func (m *ClaudeModel) Type() ModelType {
	return ModelClaude
}

// SupportedRequestTypes returns the types of requests this model supports
func (m *ClaudeModel) SupportedRequestTypes() []RequestType {
	// Claude-3 supports multimodal input (text and images)
	if strings.HasPrefix(m.modelName, "claude-3") {
		return []RequestType{TextRequest, ImageRequest, MultimodalRequest}
	}
	return []RequestType{TextRequest}
}

// ProcessText processes a text prompt and returns a standardized response
func (m *ClaudeModel) ProcessText(ctx context.Context, prompt string) (*ModelResponse, error) {
	// Create the request payload
	payload := map[string]interface{}{
		"model": m.modelName,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"max_tokens":  m.config.MaxTokens,
		"temperature": m.config.Temperature,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.Endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", m.config.APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	// Send the request
	resp, err := m.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrContextDeadlineExceeded
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors in the response status code
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				return nil, ErrRateLimitExceeded
			case http.StatusServiceUnavailable:
				return nil, ErrModelUnavailable
			default:
				return nil, fmt.Errorf("%w: %s", ErrAPICallFailed, errorResponse.Error.Message)
			}
		}

		return nil, fmt.Errorf("%w: status code %d", ErrAPICallFailed, resp.StatusCode)
	}

	// Parse the response
	var response struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the generated text
	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	var sb strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			sb.WriteString(content.Text)
		}
	}

	// Create standardized response
	modelResponse := &ModelResponse{
		Content: sb.String(),
		Raw:     response,
		Format:  FormatText,
		Metadata: map[string]interface{}{
			"model":         response.Model,
			"stop_reason":   response.StopReason,
			"input_tokens":  response.Usage.InputTokens,
			"output_tokens": response.Usage.OutputTokens,
			"message_id":    response.ID,
		},
	}

	return modelResponse, nil
}

// ProcessAudio is not natively supported by Claude, so we would need to use a speech-to-text service first
func (m *ClaudeModel) ProcessAudio(ctx context.Context, input *AudioInput, prompt string) (*ModelResponse, error) {
	return nil, ErrUnsupportedRequestType
}

// ProcessTextWithJson processes a text prompt and returns structured JSON
func (m *ClaudeModel) ProcessTextWithJson(ctx context.Context, prompt string, jsonSchema string) (*ModelResponse, error) {
	// Create a combined prompt that instructs Claude to respond with valid JSON
	systemPrompt := fmt.Sprintf(`You are a helpful assistant that always responds with valid JSON. 
Your response must follow this schema: %s

Respond only with JSON, no preamble or additional text.`, jsonSchema)

	// Create the request payload
	payload := map[string]interface{}{
		"model": m.modelName,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"max_tokens":  m.config.MaxTokens,
		"temperature": 0.2, // Lower temperature for more deterministic JSON generation
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.Endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", m.config.APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	// Send the request
	resp, err := m.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrContextDeadlineExceeded
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrAPICallFailed, resp.StatusCode)
	}

	// Parse the response
	var response struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the generated text
	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	var sb strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			sb.WriteString(content.Text)
		}
	}

	rawResponse := sb.String()

	// Extract JSON from the response (remove markdown code blocks if present)
	jsonStr := extractJSONFromText(rawResponse)

	// Verify that the response is valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		return nil, fmt.Errorf("%w: response is not valid JSON: %s", ErrInvalidJSONSchema, err.Error())
	}

	// Create standardized response
	modelResponse := &ModelResponse{
		Content: jsonStr,
		Raw:     response,
		Format:  FormatJSON,
		Metadata: map[string]interface{}{
			"model":         response.Model,
			"stop_reason":   response.StopReason,
			"input_tokens":  response.Usage.InputTokens,
			"output_tokens": response.Usage.OutputTokens,
			"message_id":    response.ID,
		},
	}

	return modelResponse, nil
}
