package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"agent/internal/ai"
	"agent/internal/models"
)

// AudioProcessor is responsible for processing audio data and extracting emergency information
type AudioProcessor struct {
	modelProvider *ai.Provider
	config        AudioProcessorConfig
}

// AudioProcessorConfig contains configuration for the audio processor
type AudioProcessorConfig struct {
	ModelEndpoint  string
	APIKey         string
	ModelType      ai.ModelType
	ModelName      string
	Timeout        time.Duration
	MaxAudioLength int // Maximum audio length in seconds
	Temperature    float64
	MaxTokens      int
}

// NewAudioProcessor creates a new audio processor
func NewAudioProcessor(config AudioProcessorConfig) (*AudioProcessor, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxAudioLength == 0 {
		config.MaxAudioLength = 300 // Default to 5 minutes
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

	return &AudioProcessor{
		modelProvider: provider,
		config:        config,
	}, nil
}

// ProcessEmergencyAudio processes audio data to extract emergency information
func (p *AudioProcessor) ProcessEmergencyAudio(ctx context.Context, audioData io.Reader) (*models.EmergencySituation, error) {
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// Prepare more comprehensive prompt for model to capture emotional tone
	prompt := `
Analyze this emergency call audio recording and provide a detailed assessment including:

1. Emergency description: Precisely what is the medical emergency situation?
2. Severity indicators: What symptoms or signs indicate the urgency level?
3. Emotional state: Assess the caller's emotional state, tone of voice, and stress level.
4. Key medical details: Extract any relevant medical history, allergies, or medications.
5. Environmental factors: Identify any contextual factors that might impact response.

Provide a comprehensive analysis that will help emergency responders prioritize and prepare for this situation.`

	// Prepare audio input
	audioInput := &ai.AudioInput{
		Audio:       audioData,
		MIMEType:    "audio/mpeg", // Default, can be overridden
		Language:    "en",         // Default to English
		AudioFormat: "mp3",        // Default format
	}

	// Process audio with model
	model := p.modelProvider.DefaultModel()
	response, err := model.ProcessAudio(ctx, audioInput, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio with model: %w", err)
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
func (p *AudioProcessor) extractStructuredInfo(ctx context.Context, description string, structuredInfo interface{}) error {
	// Define the JSON schema for structured extraction
	jsonSchema := `{
		"emergency_type": {
			"type": "string",
			"description": "Type of emergency (Medical, Fire, Crime, etc.)"
		},
		"triage_code": {
			"type": "string", 
			"enum": ["RED", "YELLOW", "GREEN", "UNKNOWN"],
			"description": "RED for life-threatening, YELLOW for urgent, GREEN for non-urgent, UNKNOWN if can't determine"
		},
		"confidence": {
			"type": "number",
			"description": "Confidence in assessment from 0.0 to 1.0"
		},
		"emotional_state": {
			"type": "object",
			"properties": {
				"distress": {"type": "number"},
				"pain": {"type": "number"}, 
				"confusion": {"type": "number"},
				"panic": {"type": "number"}
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

	// Get the model and process the text to get structured JSON
	model := p.modelProvider.DefaultModel()
	response, err := model.ProcessTextWithJson(ctx, prompt, jsonSchema)
	if err != nil {
		return fmt.Errorf("failed to extract structured info: %w", err)
	}

	// Parse JSON response into the provided structuredInfo interface
	if err := json.Unmarshal([]byte(response.Content), structuredInfo); err != nil {
		return fmt.Errorf("failed to parse structured info: %w", err)
	}

	return nil
}
