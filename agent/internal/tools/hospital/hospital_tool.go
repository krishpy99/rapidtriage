package hospital

import (
	"context"
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
	// Hospital tool is applicable for urgent (RED/YELLOW) cases
	return situation.Code == models.CodeRed || situation.Code == models.CodeYellow
}

// Execute sends the emergency information to the hospital
func (t *HospitalTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	// For now, just return a placeholder message as requested
	return &tools.ToolResponse{
		ToolName:  t.Name(),
		Success:   true,
		Message:   "Called Hospital Communication Tool",
		Data:      map[string]string{},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
