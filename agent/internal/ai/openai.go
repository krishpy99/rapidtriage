// filepath: /Users/krish/projects/rapidtriage/agent/internal/ai/openai.go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Default configuration values for OpenAI
const (
	defaultOpenAIEndpoint    = "https://api.openai.com/v1"
	defaultOpenAIModel       = "gpt-4o"
	defaultOpenAIMaxTokens   = 4096
	defaultOpenAITimeout     = 60 // seconds
	defaultOpenAITemperature = 0.7
)

// Supported OpenAI models
const (
	GPT4o      = "gpt-4o"
	GPT4oMini  = "gpt-4o-mini"
	GPT4Turbo  = "gpt-4-turbo"
	GPT4       = "gpt-4"
	GPT35Turbo = "gpt-3.5-turbo"
)

// OpenAIModel represents an implementation of the Model interface for OpenAI's API
type OpenAIModel struct {
	config       ModelConfig
	client       *http.Client
	modelName    string
	baseEndpoint string
}

// Register the OpenAI model factory
func init() {
	RegisterModel(ModelGPT4, NewOpenAIModel)
}

// NewOpenAIModel creates a new instance of the OpenAI model
func NewOpenAIModel(config ModelConfig) (Model, error) {
	// Set default values if not provided
	if config.Endpoint == "" {
		config.Endpoint = defaultOpenAIEndpoint
	}

	if config.ModelName == "" {
		config.ModelName = defaultOpenAIModel
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = defaultOpenAIMaxTokens
	}

	if config.Timeout == 0 {
		config.Timeout = defaultOpenAITimeout
	}

	if config.Temperature == 0.0 {
		config.Temperature = defaultOpenAITemperature
	}

	// Validate configuration
	if config.APIKey == "" {
		return nil, fmt.Errorf("%w: APIKey is required", ErrInvalidConfiguration)
	}

	// Create HTTP client with appropriate timeouts
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return &OpenAIModel{
		config:       config,
		client:       client,
		modelName:    config.ModelName,
		baseEndpoint: config.Endpoint,
	}, nil
}

// Name returns the name of the model
func (m *OpenAIModel) Name() string {
	return m.modelName
}

// Type returns the type of model
func (m *OpenAIModel) Type() ModelType {
	return ModelGPT4
}

// SupportedRequestTypes returns the types of requests this model supports
func (m *OpenAIModel) SupportedRequestTypes() []RequestType {
	// GPT-4o and newer models support multimodal inputs
	if strings.Contains(m.modelName, "gpt-4") || strings.Contains(m.modelName, "gpt-4o") {
		return []RequestType{TextRequest, ImageRequest, AudioRequest, MultimodalRequest}
	}
	return []RequestType{TextRequest}
}

// -- Request/Response Structures --

type OpenAIMessage struct {
	Role         string      `json:"role"`
	Content      interface{} `json:"content,omitempty"` // Can be string or array of content parts
	Name         string      `json:"name,omitempty"`
	FunctionCall *struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function_call,omitempty"`
}

type OpenAITextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type OpenAIImageContent struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type OpenAIChatRequest struct {
	Model        string          `json:"model"`
	Messages     []OpenAIMessage `json:"messages"`
	MaxTokens    int             `json:"max_tokens,omitempty"`
	Temperature  float64         `json:"temperature,omitempty"`
	Functions    interface{}     `json:"functions,omitempty"`     // Renamed from Tools
	FunctionCall interface{}     `json:"function_call,omitempty"` // Renamed from ToolChoice
}

type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      OpenAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

type OpenAIAudioTranscriptionRequest struct {
	File        []byte  `json:"file"`
	Model       string  `json:"model"`
	Language    string  `json:"language,omitempty"`
	Prompt      string  `json:"prompt,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type OpenAIAudioTranscriptionResponse struct {
	Text string `json:"text"`
}

// -- Helper function for API calls --

func (m *OpenAIModel) doRequest(ctx context.Context, url string, method string, body io.Reader, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create %s request to %s: %w", method, url, err)
	}

	// Set default Authorization header
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("DEBUG: Sending %s request to: %s\n", method, url)
	if method == "POST" && headers["Content-Type"] == "application/json" && body != nil {
		// Log JSON body carefully (could be large)
		buf := new(bytes.Buffer)
		if _, readErr := buf.ReadFrom(req.Body); readErr == nil {
			// Truncate large payloads in logs
			bodyStr := buf.String()
			// Restore the body for the actual request
			req.Body = io.NopCloser(bytes.NewBufferString(bodyStr))
		}
	}

	resp, err := m.client.Do(req)
	if err != nil {
		fmt.Printf("DEBUG: HTTP request error: %v\n", err)
		if ctx.Err() == context.DeadlineExceeded {
			return nil, nil, ErrContextDeadlineExceeded
		}
		return nil, nil, fmt.Errorf("failed to send %s request to %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: Received response with status code: %d\n", resp.StatusCode)

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return resp, respBodyBytes, nil
}

// -- Model Methods --

// ProcessText processes a text prompt and returns a text response
func (m *OpenAIModel) ProcessText(ctx context.Context, prompt string) (*ModelResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", m.baseEndpoint)

	payload := OpenAIChatRequest{
		Model: m.modelName,
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   m.config.MaxTokens,
		Temperature: m.config.Temperature,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, bodyBytes, err := m.doRequest(ctx, url, "POST", bytes.NewBuffer(jsonPayload), headers)
	if err != nil {
		return nil, err // Error already formatted by doRequest
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenAIErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				return nil, fmt.Errorf("%w: %s", ErrRateLimitExceeded, errorResponse.Error.Message)
			case http.StatusServiceUnavailable:
				return nil, fmt.Errorf("%w: %s", ErrModelUnavailable, errorResponse.Error.Message)
			default:
				return nil, fmt.Errorf("%w: %s (status: %d)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode)
			}
		}
		// Fallback error
		return nil, fmt.Errorf("%w: status code %d from %s", ErrAPICallFailed, resp.StatusCode, url)
	}

	var response OpenAIChatResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %w. Body: %s", err, string(bodyBytes))
	}

	// Check for empty response
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty or unexpected response structure from model: no choices found")
	}

	// Extract the content from the response
	content := response.Choices[0].Message.Content

	var textContent string

	// Handle different response content formats
	switch v := content.(type) {
	case string:
		textContent = v
	case []interface{}:
		// Handle array of content parts
		var result strings.Builder
		for _, part := range v {
			if contentMap, ok := part.(map[string]interface{}); ok {
				if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
					if text, ok := contentMap["text"].(string); ok {
						result.WriteString(text)
					}
				}
			}
		}
		textContent = result.String()
	default:
		return nil, fmt.Errorf("unexpected content format in response: %T", content)
	}

	// Create standardized response
	modelResponse := &ModelResponse{
		Content: textContent,
		Raw:     response,
		Format:  FormatText,
		Metadata: map[string]interface{}{
			"model":             response.Model,
			"finish_reason":     response.Choices[0].FinishReason,
			"prompt_tokens":     response.Usage.PromptTokens,
			"completion_tokens": response.Usage.CompletionTokens,
			"total_tokens":      response.Usage.TotalTokens,
		},
	}

	return modelResponse, nil
}

// ProcessAudio processes audio input and returns a text response
func (m *OpenAIModel) ProcessAudio(ctx context.Context, input *AudioInput, prompt string) (*ModelResponse, error) {
	// Read the entire audio file
	audioData, err := io.ReadAll(input.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// Step 1: First use OpenAI's Audio API for transcription
	transcription, err := m.transcribeAudio(ctx, audioData, input.MIMEType, input.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Step 2: Detect emotions and tone from the transcribed text
	emotionAnalysisResp, err := m.analyzeEmotionsAndTone(ctx, transcription)
	if err != nil {
		// Log the error but continue with the process
		fmt.Printf("Warning: emotion detection failed: %v\n", err)
		// Use empty emotion analysis if detection fails
		emotionAnalysisResp = "No emotional analysis available."
	}

	// Step 3: Generate a structured JSON response based on transcription and emotion analysis
	return m.generateEmergencyResponse(ctx, transcription, emotionAnalysisResp, prompt)
}

// analyzeEmotionsAndTone uses the completions API to analyze emotions and tone from transcribed text
func (m *OpenAIModel) analyzeEmotionsAndTone(ctx context.Context, transcription string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", m.baseEndpoint)

	// Create a prompt specifically for emotion and tone analysis
	analysisPrompt := fmt.Sprintf(`
Analyze the emotional state, tone, and urgency in this emergency call transcription:

"%s"

Focus only on detectable emotions like:
- Fear or panic
- Pain level
- Confusion or disorientation
- Distress level
- Calmness or composure
- Urgency in their voice
- Any signs of shock

Rate each detected emotion on a scale of 0-10 and explain your reasoning briefly.
`, transcription)

	payload := OpenAIChatRequest{
		Model: m.modelName,
		Messages: []OpenAIMessage{
			{Role: "user", Content: analysisPrompt},
		},
		MaxTokens:   1024, // Lower token count for emotion analysis
		Temperature: 0.3,  // Lower temperature for more consistent analysis
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal emotion analysis request: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, bodyBytes, err := m.doRequest(ctx, url, "POST", bytes.NewBuffer(jsonPayload), headers)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenAIErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return "", fmt.Errorf("%w: %s (status: %d)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode)
		}
		return "", fmt.Errorf("%w: status code %d from %s", ErrAPICallFailed, resp.StatusCode, url)
	}

	var response OpenAIChatResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to parse emotion analysis response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("empty response from emotion analysis")
	}

	content := response.Choices[0].Message.Content
	switch v := content.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("unexpected content format in emotion analysis response")
	}
}

// generateEmergencyResponse creates a structured response based on transcription and emotion analysis
func (m *OpenAIModel) generateEmergencyResponse(ctx context.Context, transcription, emotionAnalysis, prompt string) (*ModelResponse, error) {

	// Create a comprehensive prompt that includes all available information
	responsePrompt := fmt.Sprintf(`
You are analyzing an emergency call. Here is the relevant information:

TRANSCRIPTION:
%s

EMOTIONAL ANALYSIS:
%s

INSTRUCTION:
%s

Based on this information, provide a comprehensive emergency response with appropriate categorization, urgency assessment, and recommended actions.
`, transcription, emotionAnalysis, prompt)

	// Define a JSON schema for structured output
	jsonSchema := `{
		"emergency_type": {
			"type": "string",
			"description": "Type of emergency (Medical, Fire, Crime, Accident, etc.)"
		},
		"triage_code": {
			"type": "string",
			"enum": ["RED", "YELLOW", "GREEN", "UNKNOWN"],
			"description": "Triage code based on severity (RED: life-threatening, YELLOW: urgent, GREEN: non-urgent)"
		},
		"confidence": {
			"type": "number",
			"description": "Confidence level of assessment (0.0-1.0)"
		},
		"emotional_state": {
			"type": "object",
			"properties": {
				"distress": {"type": "number"},
				"panic": {"type": "number"},
				"pain": {"type": "number"},
				"confusion": {"type": "number"},
				"clarity": {"type": "number"}
			},
			"description": "Emotional markers detected in caller's voice (0.0-1.0)"
		},
		"keywords": {
			"type": "array",
			"items": {"type": "string"},
			"description": "Key medical or emergency terms extracted"
		},
		"summary": {
			"type": "string", 
			"description": "Brief summary of the emergency situation"
		},
		"recommended_actions": {
			"type": "array",
			"items": {"type": "string"},
			"description": "List of recommended immediate actions"
		}
	}`

	// Use ProcessTextWithJson to get structured output
	return m.ProcessTextWithJson(ctx, responsePrompt, jsonSchema)
}

// transcribeAudio uses OpenAI's Audio API to convert speech to text
func (m *OpenAIModel) transcribeAudio(ctx context.Context, audioData []byte, mimeType string, language string) (string, error) {
	url := fmt.Sprintf("%s/audio/transcriptions", m.baseEndpoint)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file part
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add other fields
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("failed to add model field: %w", err)
	}

	if language != "" {
		if err := writer.WriteField("language", language); err != nil {
			return "", fmt.Errorf("failed to add language field: %w", err)
		}
	}

	if err := writer.WriteField("temperature", fmt.Sprintf("%.1f", m.config.Temperature)); err != nil {
		return "", fmt.Errorf("failed to add temperature field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Set the content type header
	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	resp, bodyBytes, err := m.doRequest(ctx, url, "POST", body, headers)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenAIErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return "", fmt.Errorf("%w: %s (status: %d)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode)
		}
		return "", fmt.Errorf("%w: status code %d from %s", ErrAPICallFailed, resp.StatusCode, url)
	}

	var transcription OpenAIAudioTranscriptionResponse
	if err := json.Unmarshal(bodyBytes, &transcription); err != nil {
		return "", fmt.Errorf("failed to parse transcription response: %w", err)
	}

	return transcription.Text, nil
}

// ProcessTextWithJson processes a text prompt and returns structured JSON
func (m *OpenAIModel) ProcessTextWithJson(ctx context.Context, prompt string, jsonSchema string) (*ModelResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", m.baseEndpoint)

	// Create function specification with correct format
	functions := []map[string]interface{}{
		{
			"name":        "generate_structured_data",
			"description": "Generate structured data according to the provided schema",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": json.RawMessage(jsonSchema),
			},
		},
	}

	// Instruct the model to use the function
	instructedPrompt := fmt.Sprintf("Your task is to generate structured data based on this input: %s", prompt)

	payload := OpenAIChatRequest{
		Model: m.modelName,
		Messages: []OpenAIMessage{
			{Role: "user", Content: instructedPrompt},
		},
		Functions: functions,
		FunctionCall: map[string]string{
			"name": "generate_structured_data",
		},
		Temperature: 0.2, // Lower temperature for more predictable JSON
		MaxTokens:   m.config.MaxTokens,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request payload: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, bodyBytes, err := m.doRequest(ctx, url, "POST", bytes.NewBuffer(jsonPayload), headers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse OpenAIErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return nil, fmt.Errorf("%w: %s (status: %d)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("%w: status code %d from %s", ErrAPICallFailed, resp.StatusCode, url)
	}

	var response OpenAIChatResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %w", err)
	}

	// Debug output to help diagnose issues
	fmt.Printf("DEBUG: Response from Model: %+v\n", response)

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty response from model when expecting function call")
	}

	// Extract the function call arguments
	fc := response.Choices[0].Message.FunctionCall
	if fc == nil {
		return nil, fmt.Errorf("model did not call the function as expected")
	}

	jsonStr := fc.Arguments

	// Basic validation: Check if it's valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		fmt.Printf("DEBUG: Failed JSON validation. String was: %s\n", jsonStr)
		return nil, fmt.Errorf("%w: model response is not valid JSON: %s", ErrInvalidJSONSchema, err.Error())
	}

	// Create standardized response
	modelResponse := &ModelResponse{
		Content: jsonStr,
		Raw:     response,
		Format:  FormatJSON,
		Metadata: map[string]interface{}{
			"model":             response.Model,
			"finish_reason":     response.Choices[0].FinishReason,
			"prompt_tokens":     response.Usage.PromptTokens,
			"completion_tokens": response.Usage.CompletionTokens,
			"total_tokens":      response.Usage.TotalTokens,
			"function_name":     fc.Name,
		},
	}

	return modelResponse, nil
}
