package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io" // Note: ioutil is deprecated, but keeping for consistency with existing code
	"net/http"
	"strings"
	"time"
)

// Default configuration values for Gemini
const (
	// Updated endpoint to v1beta as it's needed for features like file processing
	defaultGeminiEndpoint    = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiModel       = "gemini-2.5-pro-latest" // Use specific latest version
	defaultGeminiMaxTokens   = 8192                    // Increased default for 1.5 pro
	defaultGeminiTimeout     = 120                     // Increased timeout for potentially longer processing
	defaultGeminiTemperature = 0.7

	// Base host for file uploads
	fileUploadHost = "https://generativelanguage.googleapis.com"
)

// Supported Gemini models
const (
	Gemini25ProLatest   = "gemini-2.5-pro-latest"   // Gemini 2.5 Pro latest
	Gemini25FlashLatest = "gemini-2.5-flash-latest" // Gemini 2.5 Flash latest
	Gemini15ProLatest   = "gemini-1.5-pro-latest"
	Gemini15FlashLatest = "gemini-1.5-flash-latest"
	GeminiUltra         = "gemini-1.5-ultra"
	GeminiFlash         = "gemini-1.5-flash"
)

// GeminiModel represents an implementation of the Model interface for Google's Gemini API
type GeminiModel struct {
	config       ModelConfig
	client       *http.Client
	modelName    string
	baseEndpoint string // Store the base endpoint for constructing URLs
}

// Register the Gemini model factory
func init() {
	RegisterModel(ModelGemini, NewGeminiModel)
}

// NewGeminiModel creates a new instance of the Gemini model
func NewGeminiModel(config ModelConfig) (Model, error) {
	// Set default values if not provided
	if config.Endpoint == "" {
		config.Endpoint = defaultGeminiEndpoint
	}
	// Ensure the endpoint uses v1beta by default if unspecified or v1
	if !strings.Contains(config.Endpoint, "/v1beta") {
		config.Endpoint = strings.Replace(config.Endpoint, "/v1", "/v1beta", 1)
		if !strings.Contains(config.Endpoint, "/v1beta") { // If "/v1" wasn't present
			// Attempt to append /v1beta, assuming a base URL was given
			if !strings.HasSuffix(config.Endpoint, "/") {
				config.Endpoint += "/"
			}
			config.Endpoint += "v1beta" // Fallback logic
		}
		fmt.Printf("INFO: Forcing endpoint to v1beta: %s\n", config.Endpoint)
	}

	if config.ModelName == "" {
		config.ModelName = defaultGeminiModel
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = defaultGeminiMaxTokens
	}

	if config.Timeout == 0 {
		config.Timeout = defaultGeminiTimeout
	}

	if config.Temperature == 0.0 {
		config.Temperature = defaultGeminiTemperature
	}

	// Validate configuration
	if config.APIKey == "" {
		return nil, fmt.Errorf("%w: APIKey is required", ErrInvalidConfiguration)
	}

	// Create HTTP client with appropriate timeouts
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return &GeminiModel{
		config:       config,
		client:       client,
		modelName:    config.ModelName,
		baseEndpoint: config.Endpoint,
	}, nil
}

// Name returns the name of the model
func (m *GeminiModel) Name() string {
	return m.modelName
}

// Type returns the type of model
func (m *GeminiModel) Type() ModelType {
	return ModelGemini
}

// SupportedRequestTypes returns the types of requests this model supports
func (m *GeminiModel) SupportedRequestTypes() []RequestType {
	// Gemini models generally support multimodal input including audio
	switch m.modelName {
	case Gemini25ProLatest, Gemini25FlashLatest, Gemini15ProLatest, Gemini15FlashLatest, GeminiUltra, GeminiFlash:
		return []RequestType{TextRequest, ImageRequest, AudioRequest, MultimodalRequest}
	default:
		// If it contains 1.5 or 2.5, assume it supports audio
		if strings.Contains(m.modelName, "1.5") || strings.Contains(m.modelName, "2.5") {
			return []RequestType{TextRequest, AudioRequest, ImageRequest, MultimodalRequest}
		}
		return []RequestType{TextRequest}
	}
}

// -- Request/Response Structures --

type GeminiGenerateRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"` // "user" or "model"
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart can be text or file data
type GeminiPart struct {
	Text     string          `json:"text,omitempty"`
	FileData *GeminiFileData `json:"file_data,omitempty"` // Correct key: file_data
}

// GeminiFileData references an uploaded file for generateContent
type GeminiFileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"` // The reference (e.g., "files/xyz") from upload
}

type GeminiGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type GeminiGenerateResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
}

type GeminiCandidate struct {
	Content       GeminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	Index         int                  `json:"index"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings"`
}

type GeminiPromptFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings"`
}

type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"` // e.g., NEGLIGIBLE, LOW, MEDIUM, HIGH
}

// -- File API Structures --

type GeminiFileUploadResponse struct {
	File GeminiFileInfo `json:"file"`
}

type GeminiFileInfo struct {
	Name        string `json:"name"` // IMPORTANT: This (e.g., "files/xyz") is the URI used in generateContent
	URI         string `json:"uri"`  // The full resource URI
	MimeType    string `json:"mimeType"`
	SizeBytes   string `json:"sizeBytes"` // Often string in Google APIs
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
	Sha256Hash  string `json:"sha256Hash"`
	DisplayName string `json:"displayName"`
}

type GeminiErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// -- Helper function for API calls --

func (m *GeminiModel) doRequest(ctx context.Context, url string, method string, body io.Reader, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create %s request to %s: %w", method, url, err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Printf("DEBUG: Sending %s request to: %s\n", method, url)
	if method == "POST" && headers["Content-Type"] == "application/json" && body != nil {
		// Log JSON body carefully (could be large)
		buf := new(bytes.Buffer)
		if _, readErr := buf.ReadFrom(req.Body); readErr == nil {
			fmt.Printf("DEBUG: Request Body (JSON): %s\n", buf.String())
			// Restore the body for the actual request
			req.Body = io.NopCloser(buf)
		} else {
			// If reading fails, log that and proceed
			fmt.Println("DEBUG: Could not read request body for logging.")
			// Ensure req.Body is still valid if it was a simple buffer initially
			if origBody, ok := body.(*bytes.Buffer); ok {
				req.Body = io.NopCloser(origBody)
			} else {
				// Fix: removed incorrect type assertion with io.NopCloser
				// Just ensure we have a valid body for the request
				if body != nil {
					req.Body = io.NopCloser(bytes.NewBuffer([]byte{}))
				}
			}
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

	respBodyBytes, err := io.ReadAll(resp.Body) // Use io.ReadAll directly
	if err != nil {
		return resp, nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	fmt.Printf("DEBUG: Response body: %s\n", string(respBodyBytes))

	return resp, respBodyBytes, nil
}

// -- Model Methods --

// ProcessText processes a text prompt and returns a text response
func (m *GeminiModel) ProcessText(ctx context.Context, prompt string) (*ModelResponse, error) {
	// Ensure we use the v1beta endpoint for consistency
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		m.baseEndpoint, m.modelName, m.config.APIKey)

	payload := GeminiGenerateRequest{
		Contents: []GeminiContent{
			{
				Role: "user",
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     m.config.Temperature,
			MaxOutputTokens: m.config.MaxTokens,
			TopP:            0.95,
			TopK:            40,
		},
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
		var errorResponse GeminiErrorResponse
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

	var response GeminiGenerateResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse successful response: %w. Body: %s", err, string(bodyBytes))
	}

	// Check for blocking first
	if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("request blocked by API, reason: %s", response.PromptFeedback.BlockReason)
	}

	// Extract the generated text
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty or unexpected response structure from model: no candidates or parts found")
	}

	// Get the content
	textContent := response.Candidates[0].Content.Parts[0].Text

	// Create standardized response
	metadata := map[string]interface{}{
		"model":         m.modelName,
		"finish_reason": response.Candidates[0].FinishReason,
	}

	// Add safety ratings to metadata
	if len(response.Candidates[0].SafetyRatings) > 0 {
		safetyRatings := make(map[string]string)
		for _, rating := range response.Candidates[0].SafetyRatings {
			safetyRatings[rating.Category] = rating.Probability
		}
		metadata["safety_ratings"] = safetyRatings
	}

	return &ModelResponse{
		Content:  textContent,
		Raw:      response,
		Format:   FormatText,
		Metadata: metadata,
	}, nil
}

// ProcessAudio processes audio input and returns a text response
func (m *GeminiModel) ProcessAudio(ctx context.Context, input *AudioInput, prompt string) (*ModelResponse, error) {
	// Read the entire audio file
	audioData, err := io.ReadAll(input.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// Determine MIME type if not provided
	mimeType := input.MIMEType
	if mimeType == "" {
		// Try to infer from AudioFormat
		switch strings.ToLower(input.AudioFormat) {
		case "mp3":
			mimeType = "audio/mpeg"
		case "wav":
			mimeType = "audio/wav"
		case "ogg":
			mimeType = "audio/ogg"
		case "flac":
			mimeType = "audio/flac"
		case "m4a":
			mimeType = "audio/m4a"
		case "aac":
			mimeType = "audio/aac"
		case "opus":
			mimeType = "audio/opus"
		default:
			// If format is unknown, cannot reliably guess MIME type
			return nil, fmt.Errorf("unknown audio format '%s', please provide a MIME type", input.AudioFormat)
		}
		fmt.Printf("DEBUG: Inferred MIME type '%s' from format '%s'\n", mimeType, input.AudioFormat)
	}

	// Step 1: Upload the audio file to get a file reference
	fileInfo, err := m.uploadAudioFile(ctx, audioData, mimeType)
	if err != nil {
		// Error already includes context from uploadAudioFile
		return nil, fmt.Errorf("failed to upload audio file: %w", err)
	}

	// Step 2: Send the analysis request with the file reference (fileInfo.Name)
	return m.generateContentFromFileUri(ctx, fileInfo.Name, mimeType, prompt)
}

// uploadAudioFile uploads an audio file to the Gemini Files API
func (m *GeminiModel) uploadAudioFile(ctx context.Context, audioData []byte, mimeType string) (*GeminiFileInfo, error) {
	// Construct the correct URL for the File API upload endpoint
	uploadUrl := fmt.Sprintf("%s/upload/v1beta/files?key=%s", fileUploadHost, m.config.APIKey)

	headers := map[string]string{
		"Content-Type":     mimeType,
		"x-goog-file-name": fmt.Sprintf("audio-upload-%d.tmp", time.Now().UnixNano()), // Temporary unique name
	}

	resp, bodyBytes, err := m.doRequest(ctx, uploadUrl, "POST", bytes.NewBuffer(audioData), headers)
	if err != nil {
		return nil, err // Error already formatted
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse GeminiErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return nil, fmt.Errorf("%w: %s (status: %d, url: %s)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode, uploadUrl)
		}
		return nil, fmt.Errorf("%w: status code %d from %s. Response: %s", ErrAPICallFailed, resp.StatusCode, uploadUrl, string(bodyBytes))
	}

	// Parse the successful response
	var fileResponse GeminiFileUploadResponse
	if err := json.Unmarshal(bodyBytes, &fileResponse); err != nil {
		return nil, fmt.Errorf("failed to parse successful file upload response: %w. Body: %s", err, string(bodyBytes))
	}

	// The 'Name' field (e.g., "files/xyz") is the reference needed for generateContent
	if fileResponse.File.Name == "" {
		return nil, fmt.Errorf("file upload response did not contain a file reference ('name'). Response: %+v", fileResponse)
	}

	fmt.Printf("DEBUG: Successfully uploaded file. File reference: %s\n", fileResponse.File.Name)
	return &fileResponse.File, nil
}

// generateContentFromFileUri sends a request to analyze audio using a file URI
func (m *GeminiModel) generateContentFromFileUri(ctx context.Context, fileRef string, mimeType string, prompt string) (*ModelResponse, error) {
	fmt.Printf("DEBUG: Starting content generation with file reference: %s\n", fileRef)

	// Use the v1beta endpoint for generateContent when using file input
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		m.baseEndpoint,
		m.modelName,
		m.config.APIKey)

	fmt.Printf("DEBUG: Content generation URL: %s\n", url)

	// Create the request payload using structs for clarity and correctness
	payload := GeminiGenerateRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt}, // Text part first
					{ // File part second
						FileData: &GeminiFileData{
							MimeType: mimeType,
							FileURI:  fileRef, // Use the file reference (e.g., "files/xyz") here
						},
					},
				},
			},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     m.config.Temperature,
			MaxOutputTokens: m.config.MaxTokens,
			TopP:            0.95,
			TopK:            40,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal generate content request payload: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, bodyBytes, err := m.doRequest(ctx, url, "POST", bytes.NewBuffer(jsonPayload), headers)
	if err != nil {
		return nil, err // Error already formatted
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse GeminiErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			switch resp.StatusCode {
			case http.StatusTooManyRequests:
				return nil, fmt.Errorf("%w: %s", ErrRateLimitExceeded, errorResponse.Error.Message)
			case http.StatusServiceUnavailable:
				return nil, fmt.Errorf("%w: %s", ErrModelUnavailable, errorResponse.Error.Message)
			default:
				// Include the specific error message from the API
				return nil, fmt.Errorf("%w: %s (status: %d, url: %s)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode, url)
			}
		}
		// Fallback if JSON error parsing fails
		return nil, fmt.Errorf("%w: status code %d from %s. Response: %s", ErrAPICallFailed, resp.StatusCode, url, string(bodyBytes))
	}

	// Parse the successful response
	var response GeminiGenerateResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse successful generate content response: %w. Body: %s", err, string(bodyBytes))
	}

	// Check for blocking first
	if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("request blocked by API, reason: %s", response.PromptFeedback.BlockReason)
	}

	// Extract the generated text
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		fmt.Printf("DEBUG: Empty or unexpected response structure. Full response: %+v\n", response)
		return nil, fmt.Errorf("empty or unexpected response structure from model: no candidates or text parts found")
	}

	// Get the content
	textContent := response.Candidates[0].Content.Parts[0].Text

	// Create standardized response
	metadata := map[string]interface{}{
		"model":           m.modelName,
		"finish_reason":   response.Candidates[0].FinishReason,
		"audio_mime_type": mimeType,
		"file_reference":  fileRef,
	}

	// Add safety ratings to metadata
	if len(response.Candidates[0].SafetyRatings) > 0 {
		safetyRatings := make(map[string]string)
		for _, rating := range response.Candidates[0].SafetyRatings {
			safetyRatings[rating.Category] = rating.Probability
		}
		metadata["safety_ratings"] = safetyRatings
	}

	return &ModelResponse{
		Content:  textContent,
		Raw:      response,
		Format:   FormatText,
		Metadata: metadata,
	}, nil
}

// ProcessTextWithJson processes a text prompt and returns structured JSON
func (m *GeminiModel) ProcessTextWithJson(ctx context.Context, prompt string, jsonSchema string) (*ModelResponse, error) {
	// Ensure we use the v1beta endpoint
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		m.baseEndpoint, m.modelName, m.config.APIKey)

	// Instruct the model to output JSON matching the schema
	instructedPrompt := fmt.Sprintf("Your response MUST be a valid JSON object adhering strictly to the following JSON schema:\n```json\n%s\n```\nBased on the following request, generate the JSON object:\n%s", jsonSchema, prompt)

	payload := GeminiGenerateRequest{
		Contents: []GeminiContent{
			{
				Role: "user",
				Parts: []GeminiPart{
					{Text: instructedPrompt},
				},
			},
		},
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     0.2, // Lower temperature for more predictable JSON
			MaxOutputTokens: m.config.MaxTokens,
		},
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
		// Handle errors same way as ProcessText/ProcessAudio
		var errorResponse GeminiErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResponse); err == nil && errorResponse.Error.Message != "" {
			return nil, fmt.Errorf("%w: %s (status: %d)", ErrAPICallFailed, errorResponse.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("%w: status code %d from %s", ErrAPICallFailed, resp.StatusCode, url)
	}

	// Parse the response
	var response GeminiGenerateResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse successful JSON response: %w. Body: %s", err, string(bodyBytes))
	}

	// Check for blocking
	if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
		return nil, fmt.Errorf("request blocked by API, reason: %s", response.PromptFeedback.BlockReason)
	}

	// Extract the generated text
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from model when expecting JSON")
	}

	rawResponse := response.Candidates[0].Content.Parts[0].Text

	// Extract JSON, removing potential markdown fences
	jsonStr := extractJSONFromText(rawResponse)

	// Basic validation: Check if it's valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err != nil {
		// Log the raw string that failed validation
		fmt.Printf("DEBUG: Failed JSON validation. String was: %s\n", jsonStr)
		return nil, fmt.Errorf("%w: model response is not valid JSON: %s", ErrInvalidJSONSchema, err.Error())
	}

	// Create standardized response
	metadata := map[string]interface{}{
		"model":         m.modelName,
		"finish_reason": response.Candidates[0].FinishReason,
	}

	// Add safety ratings to metadata
	if len(response.Candidates[0].SafetyRatings) > 0 {
		safetyRatings := make(map[string]string)
		for _, rating := range response.Candidates[0].SafetyRatings {
			safetyRatings[rating.Category] = rating.Probability
		}
		metadata["safety_ratings"] = safetyRatings
	}

	return &ModelResponse{
		Content:  jsonStr,
		Raw:      response,
		Format:   FormatJSON,
		Metadata: metadata,
	}, nil
}

// extractJSONFromText extracts JSON string, removing markdown code fences if present.
func extractJSONFromText(text string) string {
	text = strings.TrimSpace(text)
	// Handle ```json ... ```
	if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
		return strings.TrimSpace(text[7 : len(text)-3])
	}
	// Handle ``` ... ``` (generic code block)
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		return strings.TrimSpace(text[3 : len(text)-3])
	}
	// Assume it's already JSON if it starts with { or [
	if (strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}")) ||
		(strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]")) {
		return text
	}

	// If none of the above, return the text as is, validation will catch it later if it's not JSON
	fmt.Printf("WARN: Could not extract JSON from code block, returning raw text: %s\n", text)
	return text
}
