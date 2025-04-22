package triage

import (
	"context"

	"agent/internal/models"
)

// Classifier defines the interface for emergency situation classification
type Classifier interface {
	// Classify analyzes an emergency description and returns a triage code and confidence level
	Classify(ctx context.Context, situation *models.EmergencySituation) (models.TriageCode, float64, error)
}

// ClassifierConfig contains configuration options for the classifier
type ClassifierConfig struct {
	ModelPath    string
	Threshold    float64
	UseFallback  bool
	FallbackCode models.TriageCode
}
