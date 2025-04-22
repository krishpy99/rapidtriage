package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"agent/internal/models"
)

// EmergencyHandler handles emergency API requests
type EmergencyHandler struct {
	audioProcessor *AudioProcessor
	coordinator    *EmergencyCoordinator
	maxAudioSize   int64
}

// NewEmergencyHandler creates a new emergency API handler
func NewEmergencyHandler(audioProcessor *AudioProcessor, coordinator *EmergencyCoordinator, maxAudioSize int64) *EmergencyHandler {
	if maxAudioSize == 0 {
		maxAudioSize = 10 * 1024 * 1024 // Default to 10MB
	}

	return &EmergencyHandler{
		audioProcessor: audioProcessor,
		coordinator:    coordinator,
		maxAudioSize:   maxAudioSize,
	}
}

// RegisterRoutes registers the API routes
func (h *EmergencyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/emergency", h.HandleEmergency)
	mux.HandleFunc("/api/v1/health", h.HandleHealthCheck)
}

// HandleEmergency processes an incoming emergency request
func (h *EmergencyHandler) HandleEmergency(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType == "" || len(contentType) < 19 || contentType[:19] != "multipart/form-data" {
		http.Error(w, "Content-Type must be multipart/form-data", http.StatusBadRequest)
		return
	}

	// Parse multipart form with max size limit - letting Go parse the Content-Type header directly
	err := r.ParseMultipartForm(h.maxAudioSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get location data
	var location *models.Location
	locationData := r.FormValue("location")
	if locationData != "" {
		if err := json.Unmarshal([]byte(locationData), &location); err != nil {
			http.Error(w, fmt.Sprintf("Invalid location data: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Get audio file
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get audio file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Log incoming request
	log.Printf("Received emergency request with audio file: %s (size: %d bytes)",
		header.Filename, header.Size)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Process audio to extract emergency information
	situation, err := h.audioProcessor.ProcessEmergencyAudio(ctx, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process audio: %v", err), http.StatusInternalServerError)
		return
	}

	// Add location information if available
	if location != nil {
		situation.Location = location
	}

	// Process the emergency with the coordinator
	response, err := h.coordinator.ProcessEmergency(ctx, situation)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process emergency: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// HandleHealthCheck provides a basic health check endpoint
func (h *EmergencyHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health check response: %v", err)
	}
}
