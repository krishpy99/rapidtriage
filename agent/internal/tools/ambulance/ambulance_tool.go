package ambulance

import (
	"context"
	"encoding/json"
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
	// Ambulance is applicable for urgent cases that require transport
	return situation.Code == models.CodeRed || situation.Code == models.CodeYellow
}

// Execute dispatches an ambulance to the emergency location
func (t *AmbulanceTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	// Check if location is available
	if situation.Location == nil {
		return nil, fmt.Errorf("cannot dispatch ambulance: location information missing")
	}

	// Prepare dispatch request
	payload := map[string]interface{}{
		"emergency_id": situation.ID,
		"code":         string(situation.Code),
		"location": map[string]interface{}{
			"latitude":  situation.Location.Latitude,
			"longitude": situation.Location.Longitude,
			"address":   situation.Location.Address,
		},
		"description": situation.Description,
		"timestamp":   situation.Timestamp.Format(time.RFC3339),
	}

	if situation.PatientInfo != nil {
		payload["patient_info"] = situation.PatientInfo
	}

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Prepare request
	req := &HTTPRequest{
		Method: "POST",
		URL:    t.config.APIEndpoint + "/dispatch",
		Body:   body,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + t.config.APIKey,
			"X-Priority":    getPriorityFromCode(situation.Code),
		},
	}

	// Send request with retries
	var resp *HTTPResponse
	var lastErr error

	for attempt := 0; attempt < t.config.RetryAttempts; attempt++ {
		resp, lastErr = t.client.Do(req)
		if lastErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break
		}

		// Exponential backoff
		time.Sleep(time.Duration(attempt*attempt) * 100 * time.Millisecond)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to communicate with ambulance service: %w", lastErr)
	}

	// Parse response
	var responseData map[string]string
	if err := json.Unmarshal(resp.Body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse ambulance service response: %w", err)
	}

	return &tools.ToolResponse{
		ToolName:  t.Name(),
		Success:   true,
		Message:   "Successfully dispatched ambulance to the location",
		Data:      responseData,
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
