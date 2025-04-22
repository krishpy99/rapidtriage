package triage

import (
	"context"
	"strings"

	"agent/internal/models"
)

// RuleBasedClassifier implements a simple rule-based classifier
type RuleBasedClassifier struct {
	redKeywords    []string
	yellowKeywords []string
	greenKeywords  []string
	threshold      float64
	fallbackCode   models.TriageCode
}

// NewRuleBasedClassifier creates a new rule-based classifier
func NewRuleBasedClassifier(config ClassifierConfig) *RuleBasedClassifier {
	if config.Threshold == 0 {
		config.Threshold = 0.5 // Default threshold
	}

	return &RuleBasedClassifier{
		// These are very simplified examples - in a real system, these would be much more comprehensive
		redKeywords: []string{
			"not breathing", "heart attack", "stroke", "unconscious", "severe bleeding",
			"choking", "drowning", "seizure", "anaphylaxis", "overdose",
		},
		yellowKeywords: []string{
			"broken bone", "deep cut", "burn", "concussion", "severe pain",
			"high fever", "difficulty breathing", "chest pain", "allergic reaction",
		},
		greenKeywords: []string{
			"minor cut", "sprain", "mild fever", "rash", "cold symptoms",
			"ear pain", "sore throat", "minor burn", "minor headache",
		},
		threshold:    config.Threshold,
		fallbackCode: config.FallbackCode,
	}
}

// Classify implements the Classifier interface
func (c *RuleBasedClassifier) Classify(ctx context.Context, situation *models.EmergencySituation) (models.TriageCode, float64, error) {
	desc := strings.ToLower(situation.Description)

	// Check for red keywords (highest priority)
	redScore := c.calculateScore(desc, c.redKeywords)
	if redScore >= c.threshold {
		return models.CodeRed, redScore, nil
	}

	// Check for yellow keywords
	yellowScore := c.calculateScore(desc, c.yellowKeywords)
	if yellowScore >= c.threshold {
		return models.CodeYellow, yellowScore, nil
	}

	// Check for green keywords
	greenScore := c.calculateScore(desc, c.greenKeywords)
	if greenScore >= c.threshold {
		return models.CodeGreen, greenScore, nil
	}

	// If no clear classification, use fallback or return unknown
	if c.fallbackCode != "" {
		return c.fallbackCode, 0.3, nil // Low confidence
	}

	return models.CodeUnknown, 0.0, nil
}

// calculateScore computes a simple relevance score based on keyword matches
func (c *RuleBasedClassifier) calculateScore(text string, keywords []string) float64 {
	matches := 0

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			matches++
		}
	}

	if len(keywords) == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(keywords))
}
