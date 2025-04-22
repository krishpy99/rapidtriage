package ai

import (
	"context"
	"io"
)

// ModelType represents the type of AI model
type ModelType string

const (
	// ModelGemini represents Google's Gemini model
	ModelGemini ModelType = "gemini"

	// ModelClaude represents Anthropic's Claude model
	ModelClaude ModelType = "claude"

	// ModelGPT4 represents OpenAI's GPT-4 model
	ModelGPT4 ModelType = "gpt4"

	// ModelLlama represents Meta's Llama model (open source)
	ModelLlama ModelType = "llama"
)

// RequestType defines the type of request being made
type RequestType string

const (
	// TextRequest for text-only input
	TextRequest RequestType = "text"

	// AudioRequest for audio input
	AudioRequest RequestType = "audio"

	// ImageRequest for image input
	ImageRequest RequestType = "image"

	// MultimodalRequest for mixed input types
	MultimodalRequest RequestType = "multimodal"
)

// ModelResponse represents a standardized response from an AI model
type ModelResponse struct {
	// Content contains the primary text response from the model
	Content string

	// Raw contains the original raw response from the model if additional processing is needed
	Raw interface{}

	// Metadata stores any additional information about the response
	Metadata map[string]interface{}

	// Format indicates whether the response is plain text, structured JSON, etc.
	Format string
}

// Common response formats
const (
	FormatText = "text"
	FormatJSON = "json"
)

// ModelConfig contains configuration for AI models
type ModelConfig struct {
	APIKey      string
	Endpoint    string
	ModelName   string
	MaxTokens   int
	Temperature float64
	Timeout     int // Timeout in seconds
}

// AudioInput represents an audio input to be processed
type AudioInput struct {
	Audio       io.Reader
	MIMEType    string
	Language    string
	SampleRate  int
	AudioFormat string
}

// Model defines the interface for all AI model implementations
type Model interface {
	// Name returns the name of the model implementation
	Name() string

	// Type returns the type of model
	Type() ModelType

	// SupportedRequestTypes returns the types of requests this model supports
	SupportedRequestTypes() []RequestType

	// ProcessText processes a text prompt and returns a standardized response
	ProcessText(ctx context.Context, prompt string) (*ModelResponse, error)

	// ProcessAudio processes audio input and returns a standardized response
	ProcessAudio(ctx context.Context, input *AudioInput, prompt string) (*ModelResponse, error)

	// ProcessTextWithJson processes a text prompt and returns structured JSON as a standardized response
	ProcessTextWithJson(ctx context.Context, prompt string, jsonSchema string) (*ModelResponse, error)
}

// Factory function type for creating models
type ModelFactory func(config ModelConfig) (Model, error)

// Registry of model factories
var modelFactories = make(map[ModelType]ModelFactory)

// RegisterModel registers a model factory for a given model type
func RegisterModel(modelType ModelType, factory ModelFactory) {
	modelFactories[modelType] = factory
}

// GetModel returns a model instance for the specified model type
func GetModel(modelType ModelType, config ModelConfig) (Model, error) {
	factory, exists := modelFactories[modelType]
	if !exists {
		return nil, ErrUnsupportedModel
	}
	return factory(config)
}
