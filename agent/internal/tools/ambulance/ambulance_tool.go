package ambulance

import (
	"context"
	"fmt"
	"time"

	"agent/internal/models"
	"agent/internal/tools"
)

// Config contains configuration for the ambulance dispatch tool
type Config struct {
	APIEndpoint   string
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
}

// AmbulanceTool implements communication with ambulance dispatch services
type AmbulanceTool struct {
	config Config
	client HTTPClient
}

// HTTPClient defines the interface for HTTP clients
type HTTPClient interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest and HTTPResponse are simplified HTTP structures
type HTTPRequest struct {
	Method  string
	URL     string
	Body    []byte
	Headers map[string]string
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// UniversalClientAdapter adapts a universal HTTP client to the AmbulanceTool's HTTPClient interface
type UniversalClientAdapter struct {
	UniversalClient interface {
		Do(req interface{}) (interface{}, error)
	}
}

// Do implements the ambulance.HTTPClient interface
func (a *UniversalClientAdapter) Do(req *HTTPRequest) (*HTTPResponse, error) {
	resp, err := a.UniversalClient.Do(req)
	if err != nil {
		return nil, err
	}

	if httpResp, ok := resp.(*HTTPResponse); ok {
		return httpResp, nil
	}

	return nil, fmt.Errorf("unexpected response type: %T", resp)
}

// NewAmbulanceTool creates a new ambulance dispatch tool
func NewAmbulanceTool(config Config, client HTTPClient) *AmbulanceTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	return &AmbulanceTool{
		config: config,
		client: client,
	}
}

// Name returns the name of the tool
func (t *AmbulanceTool) Name() string {
	return "Ambulance Dispatch Tool"
}

// IsApplicable determines if this tool is applicable for the given emergency
func (t *AmbulanceTool) IsApplicable(situation *models.EmergencySituation) bool {
	// Ambulance is only applicable for critical (RED) cases
	return situation.Code == models.CodeRed && situation.Location != nil
}

// Execute dispatches an ambulance to the emergency location
func (t *AmbulanceTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	// For now, just return a placeholder message as requested
	return &tools.ToolResponse{
		ToolName:  t.Name(),
		Success:   true,
		Message:   "Called Ambulance Dispatch Tool",
		Data:      map[string]string{},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// getPriorityFromCode converts a triage code to a priority string
func getPriorityFromCode(code models.TriageCode) string {
	switch code {
	case models.CodeRed:
		return "HIGH"
	case models.CodeYellow:
		return "MEDIUM"
	case models.CodeGreen:
		return "LOW"
	default:
		return "MEDIUM" // Default to medium priority if unknown
	}
}
