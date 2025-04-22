package location

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"agent/internal/models"
	"agent/internal/tools"
)

// Facility represents a medical facility or ambulance
type Facility struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"` // "hospital" or "ambulance"
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
	Distance  float64 `json:"distance,omitempty"` // Distance in kilometers
}

// Config contains configuration for the location tool
type Config struct {
	APIEndpoint   string
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
	MaxResults    int
	MaxDistance   float64 // Maximum distance in kilometers
}

// LocationTool implements functionality to find nearby medical facilities
type LocationTool struct {
	config     Config
	client     HTTPClient
	cache      map[string][]Facility // Simple in-memory cache
	cacheTTL   time.Duration
	lastUpdate time.Time
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

// UniversalClientAdapter adapts a universal HTTP client to the LocationTool's HTTPClient interface
type UniversalClientAdapter struct {
	UniversalClient interface {
		Do(req interface{}) (interface{}, error)
	}
}

// Do implements the location.HTTPClient interface
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

// NewLocationTool creates a new location tool
func NewLocationTool(config Config, client HTTPClient) *LocationTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	if config.MaxResults == 0 {
		config.MaxResults = 5
	}

	if config.MaxDistance == 0 {
		config.MaxDistance = 50.0 // Default to 50km
	}

	return &LocationTool{
		config:     config,
		client:     client,
		cache:      make(map[string][]Facility),
		cacheTTL:   30 * time.Minute,
		lastUpdate: time.Now(),
	}
}

// Name returns the name of the tool
func (t *LocationTool) Name() string {
	return "Location Services Tool"
}

// IsApplicable determines if this tool is applicable for the given emergency
func (t *LocationTool) IsApplicable(situation *models.EmergencySituation) bool {
	// This tool is applicable if the emergency situation has location information
	return situation.Location != nil
}

// Execute finds nearby hospitals or ambulances
func (t *LocationTool) Execute(ctx context.Context, situation *models.EmergencySituation) (*tools.ToolResponse, error) {
	if situation.Location == nil {
		return nil, fmt.Errorf("location information missing")
	}

	// Try to get facilities from cache first, if not too old
	cacheKey := fmt.Sprintf("%.4f:%.4f", situation.Location.Latitude, situation.Location.Longitude)
	if facilities, ok := t.cache[cacheKey]; ok && time.Since(t.lastUpdate) < t.cacheTTL {
		return t.createResponse(situation, facilities)
	}

	// Prepare API request payload
	payload := map[string]interface{}{
		"latitude":       situation.Location.Latitude,
		"longitude":      situation.Location.Longitude,
		"max_distance":   t.config.MaxDistance,
		"max_results":    t.config.MaxResults,
		"emergency_code": string(situation.Code),
	}

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Prepare request
	req := &HTTPRequest{
		Method: "POST",
		URL:    t.config.APIEndpoint + "/facilities/nearby",
		Body:   body,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + t.config.APIKey,
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
		return nil, fmt.Errorf("failed to communicate with location service: %w", lastErr)
	}

	// Parse response
	var facilities []Facility
	if err := json.Unmarshal(resp.Body, &facilities); err != nil {
		return nil, fmt.Errorf("failed to parse location service response: %w", err)
	}

	// Calculate distances and sort by distance
	for i := range facilities {
		facilities[i].Distance = calculateDistance(
			situation.Location.Latitude,
			situation.Location.Longitude,
			facilities[i].Latitude,
			facilities[i].Longitude,
		)
	}

	sort.Slice(facilities, func(i, j int) bool {
		return facilities[i].Distance < facilities[j].Distance
	})

	// Update cache
	t.cache[cacheKey] = facilities
	t.lastUpdate = time.Now()

	return t.createResponse(situation, facilities)
}

// FilterByType filters facilities by type and returns the closest ones
func (t *LocationTool) FilterByType(facilities []Facility, facilityType string, maxResults int) []Facility {
	var filtered []Facility

	for _, facility := range facilities {
		if facility.Type == facilityType {
			filtered = append(filtered, facility)
		}
	}

	if maxResults > 0 && len(filtered) > maxResults {
		filtered = filtered[:maxResults]
	}

	return filtered
}

// GetNearestHospitals returns the nearest hospitals
func (t *LocationTool) GetNearestHospitals(ctx context.Context, location *models.Location, maxResults int) ([]Facility, error) {
	// Create a temporary situation to use with Execute
	situation := &models.EmergencySituation{
		ID:        "temp",
		Location:  location,
		Timestamp: time.Now(),
	}

	response, err := t.Execute(ctx, situation)
	if err != nil {
		return nil, err
	}

	// Parse facilities from response data
	var allFacilities []Facility
	facilitiesData, _ := json.Marshal(response.Data["facilities"])
	if err := json.Unmarshal(facilitiesData, &allFacilities); err != nil {
		return nil, fmt.Errorf("failed to parse facilities: %w", err)
	}

	return t.FilterByType(allFacilities, "hospital", maxResults), nil
}

// GetNearestAmbulances returns the nearest ambulances
func (t *LocationTool) GetNearestAmbulances(ctx context.Context, location *models.Location, maxResults int) ([]Facility, error) {
	// Create a temporary situation to use with Execute
	situation := &models.EmergencySituation{
		ID:        "temp",
		Location:  location,
		Timestamp: time.Now(),
	}

	response, err := t.Execute(ctx, situation)
	if err != nil {
		return nil, err
	}

	// Parse facilities from response data
	var allFacilities []Facility
	facilitiesData, _ := json.Marshal(response.Data["facilities"])
	if err := json.Unmarshal(facilitiesData, &allFacilities); err != nil {
		return nil, fmt.Errorf("failed to parse facilities: %w", err)
	}

	return t.FilterByType(allFacilities, "ambulance", maxResults), nil
}

// createResponse formats the tool response with nearby facilities
func (t *LocationTool) createResponse(situation *models.EmergencySituation, facilities []Facility) (*tools.ToolResponse, error) {
	// Convert facilities to JSON
	facilitiesJSON, err := json.Marshal(facilities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal facilities: %w", err)
	}

	data := map[string]string{
		"facilities":       string(facilitiesJSON),
		"num_facilities":   fmt.Sprintf("%d", len(facilities)),
		"source_latitude":  fmt.Sprintf("%.6f", situation.Location.Latitude),
		"source_longitude": fmt.Sprintf("%.6f", situation.Location.Longitude),
	}

	return &tools.ToolResponse{
		ToolName:  t.Name(),
		Success:   true,
		Message:   fmt.Sprintf("Found %d nearby medical facilities", len(facilities)),
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// calculateDistance uses the Haversine formula to calculate distance between coordinates in kilometers
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
