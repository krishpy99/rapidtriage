package hospital

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agent/internal/models"
	"agent/internal/tools"
)

// Config contains configuration for the hospital communication tool
type Config struct {
	APIEndpoint   string
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
}

// HospitalTool implements communication with hospital emergency departments
type HospitalTool struct {
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

// UniversalClientAdapter adapts a universal HTTP client to the HospitalTool's HTTPClient interface
type UniversalClientAdapter struct {
	UniversalClient interface {
		Do(req interface{}) (interface{}, error)
	}
}

// Do implements the hospital.HTTPClient interface
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

// NewHospitalTool creates a new hospital communication tool
func NewHospitalTool(config Config, client HTTPClient) *HospitalTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	return &HospitalTool{
		config: config,
		client: client,
	}
}

// Name returns the name of the tool
func (t *HospitalTool) Name() string {
	return "Hospital Communication Tool"
}

// IsApplicable determines if this tool is applicable for the given emergency
func (t *HospitalTool) IsApplicable(situation *models.EmergencySituation) bool {
	// This tool is applicable for all emergencies
	return true
}

// Execute sends the emergency information to the hospital
func (t *HospitalTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	// Create request payload
	payload := map[string]interface{}{
		"emergency_id": situation.ID,
		"code":         string(situation.Code),
		"description":  situation.Description,
		"timestamp":    situation.Timestamp.Format(time.RFC3339),
	}

	if situation.Location != nil {
		payload["location"] = situation.Location
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
		URL:    t.config.APIEndpoint + "/emergencies",
		Body:   body,
		Headers: map[string]string{
			"Content-Type":     "application/json",
			"Authorization":    "Bearer " + t.config.APIKey,
			"X-Emergency-Code": string(situation.Code),
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
		return nil, fmt.Errorf("failed to communicate with hospital API: %w", lastErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hospital API returned status code: %d", resp.StatusCode)
	}

	// Parse response
	var responseData map[string]string
	if err := json.Unmarshal(resp.Body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse hospital API response: %w", err)
	}

	return &tools.ToolResponse{
		ToolName:  t.Name(),
		Success:   true,
		Message:   "Successfully communicated emergency to hospital",
		Data:      responseData,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
