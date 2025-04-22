package tools

import (
	"sync"

	"agent/internal/models"
)

// DefaultToolRegistry implements the ToolRegistry interface
type DefaultToolRegistry struct {
	tools []EmergencyTool
	mu    sync.RWMutex
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *DefaultToolRegistry {
	return &DefaultToolRegistry{
		tools: make([]EmergencyTool, 0),
	}
}

// Register adds a tool to the registry
func (r *DefaultToolRegistry) Register(tool EmergencyTool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = append(r.tools, tool)
	return nil
}

// GetAll returns all registered tools
func (r *DefaultToolRegistry) GetAll() []EmergencyTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]EmergencyTool, len(r.tools))
	copy(result, r.tools)

	return result
}

// GetApplicable returns tools applicable to the given emergency situation
func (r *DefaultToolRegistry) GetApplicable(situation *models.EmergencySituation) []EmergencyTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var applicable []EmergencyTool

	for _, tool := range r.tools {
		if tool.IsApplicable(situation) {
			applicable = append(applicable, tool)
		}
	}

	return applicable
}
