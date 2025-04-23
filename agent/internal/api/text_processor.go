package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agent/internal/ai"
	"agent/internal/models"
)

// TextProcessor is responsible for processing text data and extracting emergency information
type TextProcessor struct {
	modelProvider *ai.Provider
	config        TextProcessorConfig
}

// TextProcessorConfig contains configuration for the text processor
type TextProcessorConfig struct {
	ModelEndpoint string
	APIKey        string
	ModelType     ai.ModelType
	ModelName     string
	Timeout       time.Duration
	Temperature   float64
	MaxTokens     int
}

// NewTextProcessor creates a new text processor
func NewTextProcessor(config TextProcessorConfig) (*TextProcessor, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.ModelType == "" {
		config.ModelType = ai.ModelGemini // Default to Gemini
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	// Create model configuration
	modelConfig := ai.ModelConfig{
		APIKey:      config.APIKey,
		Endpoint:    config.ModelEndpoint,
		ModelName:   config.ModelName,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
		Timeout:     int(config.Timeout.Seconds()),
	}

	// Create AI provider with default model
	provider, err := ai.NewProvider(config.ModelType, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	return &TextProcessor{
		modelProvider: provider,
		config:        config,
	}, nil
}

// ProcessEmergencyText processes text data to extract emergency information
func (p *TextProcessor) ProcessEmergencyText(ctx context.Context, text string) (*models.EmergencySituation, error) {
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// Prepare prompt for the model to analyze the emergency text
	prompt := `
Analyze this emergency text description and provide a detailed assessment including:

1. Emergency description: Precisely what is the medical emergency situation?
2. Severity indicators: What symptoms or signs indicate the urgency level?
3. Emotional state: Assess the emotional state based on the text.
4. Key medical details: Extract any relevant medical history, allergies, or medications.
5. Environmental factors: Identify any contextual factors that might impact response.

Provide a comprehensive analysis that will help emergency responders prioritize and prepare for this situation.`

	// Process text with model
	model := p.modelProvider.DefaultModel()
	response, err := model.ProcessText(ctx, prompt + "\n\nText: " + text)
	if err != nil {
		return nil, fmt.Errorf("failed to process text with model: %w", err)
	}

	// Parse the structured JSON response
	var structuredInfo struct {
		EmergencyType      string             `json:"emergency_type"`
		TriageCode         string             `json:"triage_code"`
		Confidence         float64            `json:"confidence"`
		EmotionalState     map[string]float64 `json:"emotional_state"`
		Keywords           []string           `json:"keywords"`
		Summary            string             `json:"summary"`
		RecommendedActions []string           `json:"recommended_actions"`
	}

	if response.Format == ai.FormatJSON {
		// The response is already in JSON format
		if err := json.Unmarshal([]byte(response.Content), &structuredInfo); err != nil {
			return nil, fmt.Errorf("failed to parse structured response: %w", err)
		}
	} else {
		// For text format, try to extract structured information
		if err := p.extractStructuredInfo(ctx, response.Content, &structuredInfo); err != nil {
			return nil, fmt.Errorf("failed to extract structured info from text response: %w", err)
		}
	}

	// Create a new emergency situation with the extracted description
	situation := models.NewEmergencySituation(structuredInfo.Summary)

	// Map the triage code from the response
	var triageCode models.TriageCode
	switch structuredInfo.TriageCode {
	case "RED":
		triageCode = models.CodeRed
	case "YELLOW":
		triageCode = models.CodeYellow
	case "GREEN":
		triageCode = models.CodeGreen
	default:
		triageCode = models.CodeUnknown
	}

	// Set triage code and confidence
	situation.SetTriageCode(triageCode, structuredInfo.Confidence)

	// Set keywords and emotional markers
	situation.Keywords = structuredInfo.Keywords
	situation.EmotionalMarkers = structuredInfo.EmotionalState

	// Add metadata for emergency type and recommended actions
	situation.Metadata["emergency_type"] = structuredInfo.EmergencyType
	situation.Metadata["model_used"] = model.Name()

	// If available, add model-specific metadata
	if response.Metadata != nil {
		for key, value := range response.Metadata {
			metaKey := fmt.Sprintf("model_meta_%s", key)
			metaValue := fmt.Sprintf("%v", value)
			situation.Metadata[metaKey] = metaValue
		}
	}

	if len(structuredInfo.RecommendedActions) > 0 {
		actionsJSON, err := json.Marshal(structuredInfo.RecommendedActions)
		if err == nil {
			situation.Metadata["recommended_actions"] = string(actionsJSON)
		}
	}

	return situation, nil
}

// extractStructuredInfo uses the AI model to extract structured information from the text
func (p *TextProcessor) extractStructuredInfo(ctx context.Context, description string, structuredInfo interface{}) error {
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
			"description": "Emotional states from 0.0 to 1.0"
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
			"description": "Recommended immediate actions"
		}
	}`

	// Prepare prompt for structured extraction
	prompt := fmt.Sprintf(`
Based on this emergency description: "%s"

Please extract and format the information as structured JSON according to the provided schema.
Include only information that can be clearly inferred from the emergency description.
`, description)

	// Get structured JSON from model
	model := p.modelProvider.DefaultModel()
	response, err := model.ProcessTextWithJson(ctx, prompt, jsonSchema)
	if err != nil {
		return fmt.Errorf("failed to extract structured information: %w", err)
	}

	if err := json.Unmarshal([]byte(response.Content), structuredInfo); err != nil {
		return fmt.Errorf("failed to parse structured information: %w", err)
	}

	return nil
}