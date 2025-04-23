package booking

import (
	"context"
	"fmt"
	"time"

	"agent/internal/models"
	"agent/internal/tools"
)

// Config contains configuration for the booking tool
type Config struct {
	APIEndpoint   string
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
}

// BookingTool implements functionality to get booking URLs for non-urgent cases
type BookingTool struct {
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

// UniversalClientAdapter adapts a universal HTTP client to the BookingTool's HTTPClient interface
type UniversalClientAdapter struct {
	UniversalClient interface {
		Do(req interface{}) (interface{}, error)
	}
}

// Do implements the booking.HTTPClient interface
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

// NewBookingTool creates a new booking tool
func NewBookingTool(config Config, client HTTPClient) *BookingTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	return &BookingTool{
		config: config,
		client: client,
	}
}

// Name returns the name of the tool
func (t *BookingTool) Name() string {
	return "Hospital Booking Tool"
}

// IsApplicable determines if this tool is applicable for the given emergency
func (t *BookingTool) IsApplicable(situation *models.EmergencySituation) bool {
	// This tool is only applicable for non-urgent cases
	return situation.Code == models.CodeGreen
}

// Execute retrieves booking URLs for the nearest hospitals
func (t *BookingTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	// For now, just return a placeholder message
	return &tools.ToolResponse{
		ToolName: t.Name(),
		Success:  true,
		Message:  "Called Hospital Booking Tool",
		Data: map[string]string{
			"booking_url": "https://hospital-booking.example.com",
			"hospital_id": "nearest-hospital-123",
			"wait_time":   "30 minutes",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
