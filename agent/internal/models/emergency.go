package models

import (
	"time"
)

// TriageCode represents the severity level of a medical emergency
type TriageCode string

const (
	// CodeRed represents life-threatening emergencies requiring immediate intervention
	CodeRed TriageCode = "RED"

	// CodeYellow represents urgent but not immediately life-threatening situations
	CodeYellow TriageCode = "YELLOW"

	// CodeGreen represents non-urgent cases requiring medical attention
	CodeGreen TriageCode = "GREEN"

	// CodeUnknown represents situations that could not be classified
	CodeUnknown TriageCode = "UNKNOWN"
)

// EmergencySituation represents a medical emergency situation
type EmergencySituation struct {
	ID               string             `json:"id"`
	Description      string             `json:"description"`
	Code             TriageCode         `json:"code"`
	Confidence       float64            `json:"confidence"`
	Location         *Location          `json:"location,omitempty"`
	Timestamp        time.Time          `json:"timestamp"`
	PatientInfo      *PatientInfo       `json:"patient_info,omitempty"`
	EmotionalMarkers map[string]float64 `json:"emotional_markers,omitempty"`
	Keywords         []string           `json:"keywords,omitempty"`
	Metadata         map[string]string  `json:"metadata,omitempty"`
}

// Location represents geolocation information
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// PatientInfo contains basic information about the patient
type PatientInfo struct {
	Name      string   `json:"name,omitempty"`
	Age       int      `json:"age,omitempty"`
	Gender    string   `json:"gender,omitempty"`
	Allergies []string `json:"allergies,omitempty"`
}

// NewEmergencySituation creates a new emergency situation with default values
func NewEmergencySituation(description string) *EmergencySituation {
	return &EmergencySituation{
		ID:          generateUUID(),
		Description: description,
		Code:        CodeUnknown,
		Confidence:  0.0,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]string),
	}
}

// Helper function to generate a UUID
func generateUUID() string {
	// In a real implementation, you'd use a proper UUID library
	return "emergency-" + time.Now().Format("20060102-150405.000")
}

// SetTriageCode sets the triage code and confidence level
func (e *EmergencySituation) SetTriageCode(code TriageCode, confidence float64) {
	e.Code = code
	e.Confidence = confidence
}

// IsLifeThreatening returns true if the emergency is classified as life-threatening
func (e *EmergencySituation) IsLifeThreatening() bool {
	return e.Code == CodeRed
}
