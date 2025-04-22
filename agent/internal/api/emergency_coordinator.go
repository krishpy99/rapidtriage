package api

import (
	"context"
	"fmt"
	"time"

	"agent/internal/models"
	"agent/internal/tools"
	"agent/internal/tools/location"
)

// EmergencyCoordinator manages the emergency response process
type EmergencyCoordinator struct {
	classifier         Classifier
	toolRegistry       tools.ToolRegistry
	locationTool       *location.LocationTool
	summaryGenerator   SummaryGenerator
	notificationConfig NotificationConfig
}

// Classifier defines the interface for emergency classification
type Classifier interface {
	Classify(ctx context.Context, situation *models.EmergencySituation) (models.TriageCode, float64, error)
}

// SummaryGenerator generates emergency summaries for responders
type SummaryGenerator interface {
	GenerateSummary(ctx context.Context, situation *models.EmergencySituation, responses []*tools.ToolResponse) (string, error)
}

// NotificationConfig contains settings for emergency notifications
type NotificationConfig struct {
	EnableSMS     bool
	EnableEmail   bool
	EnablePush    bool
	RetryAttempts int
	RetryInterval time.Duration
}

// CoordinatorConfig contains configuration for the emergency coordinator
type CoordinatorConfig struct {
	MaxConcurrentTools int
	Notifications      NotificationConfig
	DefaultTimeout     time.Duration
}

// NewEmergencyCoordinator creates a new emergency coordinator
func NewEmergencyCoordinator(
	classifier Classifier,
	toolRegistry tools.ToolRegistry,
	locationTool *location.LocationTool,
	summaryGenerator SummaryGenerator,
	config CoordinatorConfig,
) *EmergencyCoordinator {
	if config.MaxConcurrentTools == 0 {
		config.MaxConcurrentTools = 5
	}

	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}

	return &EmergencyCoordinator{
		classifier:         classifier,
		toolRegistry:       toolRegistry,
		locationTool:       locationTool,
		summaryGenerator:   summaryGenerator,
		notificationConfig: config.Notifications,
	}
}

// ProcessEmergency processes an emergency situation
func (c *EmergencyCoordinator) ProcessEmergency(ctx context.Context, situation *models.EmergencySituation) (*EmergencyResponse, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Classify the emergency if not already classified
	if situation.Code == models.CodeUnknown {
		code, confidence, err := c.classifier.Classify(ctx, situation)
		if err != nil {
			return nil, fmt.Errorf("failed to classify emergency: %w", err)
		}
		situation.SetTriageCode(code, confidence)
	}

	// Find nearby hospitals and ambulances
	var nearestHospitals []location.Facility
	var nearestAmbulances []location.Facility
	var toolResponses []*tools.ToolResponse
	var responseErr error

	if situation.Location != nil {
		// Get nearest hospitals
		nearestHospitals, responseErr = c.locationTool.GetNearestHospitals(ctx, situation.Location, 3)
		if responseErr != nil {
			// Log but continue
			fmt.Printf("Warning: failed to get nearest hospitals: %v\n", responseErr)
		}

		// Get nearest ambulances for higher priority cases
		if situation.Code == models.CodeRed || situation.Code == models.CodeYellow {
			nearestAmbulances, responseErr = c.locationTool.GetNearestAmbulances(ctx, situation.Location, 2)
			if responseErr != nil {
				// Log but continue
				fmt.Printf("Warning: failed to get nearest ambulances: %v\n", responseErr)
			}
		}
	}

	// Get applicable tools
	applicableTools := c.toolRegistry.GetApplicable(situation)

	// Execute tools and collect responses
	for _, tool := range applicableTools {
		toolResponse, err := tool.Execute(ctx, situation)
		if err != nil {
			// Log error but continue with other tools
			fmt.Printf("Warning: tool %s failed: %v\n", tool.Name(), err)
			continue
		}

		toolResponses = append(toolResponses, toolResponse)
	}

	// Generate a summary for responders
	summary, err := c.summaryGenerator.GenerateSummary(ctx, situation, toolResponses)
	if err != nil {
		// Use a simplified summary if generator fails
		summary = fmt.Sprintf("Emergency: %s (Code %s). Confidence: %.2f",
			situation.Description, situation.Code, situation.Confidence)
	}

	// Create emergency response
	response := &EmergencyResponse{
		EmergencyID:   situation.ID,
		Code:          situation.Code,
		Summary:       summary,
		Timestamp:     time.Now().Format(time.RFC3339),
		ToolResponses: toolResponses,
	}

	if len(nearestHospitals) > 0 {
		response.NearestHospitals = nearestHospitals
	}

	if len(nearestAmbulances) > 0 {
		response.NearestAmbulances = nearestAmbulances
	}

	return response, nil
}

// EmergencyResponse represents the coordinated emergency response
type EmergencyResponse struct {
	EmergencyID       string                `json:"emergency_id"`
	Code              models.TriageCode     `json:"code"`
	Summary           string                `json:"summary"`
	Timestamp         string                `json:"timestamp"`
	NearestHospitals  []location.Facility   `json:"nearest_hospitals,omitempty"`
	NearestAmbulances []location.Facility   `json:"nearest_ambulances,omitempty"`
	ToolResponses     []*tools.ToolResponse `json:"tool_responses,omitempty"`
}

// DefaultSummaryGenerator implements a basic summary generator
type DefaultSummaryGenerator struct{}

// GenerateSummary generates a human-readable summary of the emergency
func (g *DefaultSummaryGenerator) GenerateSummary(ctx context.Context, situation *models.EmergencySituation, responses []*tools.ToolResponse) (string, error) {
	// In a real implementation, this would use a language model to generate a cohesive summary
	// This is a simplified version

	priorityText := getPriorityText(situation.Code)
	summary := fmt.Sprintf("EMERGENCY ALERT: %s - %s\n\n", priorityText, situation.Code)
	summary += fmt.Sprintf("Description: %s\n", situation.Description)

	if situation.PatientInfo != nil {
		summary += "\nPATIENT INFO:\n"
		if situation.PatientInfo.Name != "" {
			summary += fmt.Sprintf("Name: %s\n", situation.PatientInfo.Name)
		}
		if situation.PatientInfo.Age > 0 {
			summary += fmt.Sprintf("Age: %d\n", situation.PatientInfo.Age)
		}
		if situation.PatientInfo.Gender != "" {
			summary += fmt.Sprintf("Gender: %s\n", situation.PatientInfo.Gender)
		}
		if len(situation.PatientInfo.Allergies) > 0 {
			summary += fmt.Sprintf("Allergies: %v\n", situation.PatientInfo.Allergies)
		}
	}

	if situation.Location != nil {
		summary += fmt.Sprintf("\nLOCATION: Lat %.6f, Long %.6f\n",
			situation.Location.Latitude, situation.Location.Longitude)
		if situation.Location.Address != "" {
			summary += fmt.Sprintf("Address: %s\n", situation.Location.Address)
		}
	}

	// Add timestamps
	summary += fmt.Sprintf("\nEmergency reported at: %s\n", situation.Timestamp.Format(time.RFC3339))
	summary += fmt.Sprintf("Alert generated at: %s\n", time.Now().Format(time.RFC3339))

	// Add confidence
	summary += fmt.Sprintf("\nAssessment confidence: %.1f%%\n", situation.Confidence*100)

	return summary, nil
}

// getPriorityText returns a descriptive text for the priority level
func getPriorityText(code models.TriageCode) string {
	switch code {
	case models.CodeRed:
		return "CRITICAL - IMMEDIATE RESPONSE REQUIRED"
	case models.CodeYellow:
		return "URGENT - PROMPT RESPONSE REQUIRED"
	case models.CodeGreen:
		return "NON-URGENT - STANDARD RESPONSE"
	default:
		return "UNCLASSIFIED EMERGENCY"
	}
}
