package ai

import (
	"fmt"
)

// Provider manages AI model instances
type Provider struct {
	defaultModel Model
	models       map[string]Model
}

// NewProvider creates a new AI model provider
func NewProvider(defaultModelType ModelType, config ModelConfig) (*Provider, error) {
	// Create the default model
	defaultModel, err := GetModel(defaultModelType, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create default model: %w", err)
	}

	// Create the provider with the default model
	provider := &Provider{
		defaultModel: defaultModel,
		models:       make(map[string]Model),
	}

	// Add the default model to the models map
	provider.models[string(defaultModelType)] = defaultModel

	return provider, nil
}

// DefaultModel returns the default model
func (p *Provider) DefaultModel() Model {
	return p.defaultModel
}

// Model returns a specific model by type or the default model if not found
func (p *Provider) Model(modelType ModelType) Model {
	if model, ok := p.models[string(modelType)]; ok {
		return model
	}
	return p.defaultModel
}

// AddModel adds a new model to the provider
func (p *Provider) AddModel(modelType ModelType, config ModelConfig) error {
	// Check if the model already exists
	if _, ok := p.models[string(modelType)]; ok {
		return fmt.Errorf("model %s already exists", modelType)
	}

	// Create the new model
	model, err := GetModel(modelType, config)
	if err != nil {
		return err
	}

	// Add the model to the provider
	p.models[string(modelType)] = model
	return nil
}

// WithDefaultModel returns a new provider with a different default model
func (p *Provider) WithDefaultModel(modelType ModelType) (*Provider, error) {
	if model, ok := p.models[string(modelType)]; ok {
		return &Provider{
			defaultModel: model,
			models:       p.models,
		}, nil
	}
	return nil, fmt.Errorf("model %s not found", modelType)
}

// DetectMIMEType attempts to detect the MIME type from the audio format
func DetectMIMEType(format string) string {
	switch format {
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	case "flac":
		return "audio/flac"
	case "m4a":
		return "audio/mp4"
	default:
		return "audio/mpeg" // Default to MP3
	}
}
