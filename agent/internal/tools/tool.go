package tools

import (
	"context"

	"agent/internal/models"
)

// ToolResponse represents the response from an emergency tool
type ToolResponse struct {
	ToolName  string            `json:"tool_name"`
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data,omitempty"`
	Timestamp string            `json:"timestamp"`
}

// EmergencyTool defines the interface for emergency response tools
type EmergencyTool interface {
	// Name returns the name of the tool
	Name() string

	// IsApplicable determines if this tool is applicable for the given emergency situation
	IsApplicable(situation *models.EmergencySituation) bool

	// Execute runs the tool's logic for the given emergency situation
	Execute(ctx context.Context, situation *models.EmergencySituation) (*ToolResponse, error)
}

// ToolRegistry maintains a registry of available emergency tools
type ToolRegistry interface {
	// Register adds a tool to the registry
	Register(tool EmergencyTool) error

	// GetAll returns all registered tools
	GetAll() []EmergencyTool

	// GetApplicable returns tools applicable to the given emergency situation
	GetApplicable(situation *models.EmergencySituation) []EmergencyTool
}
