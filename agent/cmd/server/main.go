package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent/internal/ai"
	"agent/internal/api"
	"agent/internal/config"
	"agent/internal/tools"
	"agent/internal/tools/ambulance"
	"agent/internal/tools/booking"
	"agent/internal/tools/hospital"
	"agent/internal/tools/location"
	"agent/internal/triage"
)

// Configuration constants
const (
	defaultPort       = 8080
	defaultAPITimeout = 30 * time.Second
	maxAudioSize      = 20 * 1024 * 1024 // 20MB
)

func main() {
	log.Println("Starting RapidTriage API server...")

	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil && err != config.ErrEnvFileNotFound {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Set up a context that will be canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create components
	components, err := setupComponents()
	if err != nil {
		log.Fatalf("Failed to set up components: %v", err)
	}

	// Create HTTP server
	port := config.GetInt("PORT", defaultPort)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      components.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create a deadline for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// Components holds all the application components
type Components struct {
	mux              *http.ServeMux
	toolRegistry     *tools.DefaultToolRegistry
	locationTool     *location.LocationTool
	audioProcessor   *api.AudioProcessor
	emergencyHandler *api.EmergencyHandler
}

// setupComponents initializes all application components
func setupComponents() (*Components, error) {
	// Create HTTP client (simplified for this implementation)
	httpClient := &mockHTTPClient{}

	// Create tool registry
	toolRegistry := tools.NewToolRegistry()

	// Create and register tools
	locationTool := createLocationTool(httpClient)
	hospitalTool := createHospitalTool(httpClient)
	ambulanceTool := createAmbulanceTool(httpClient)
	bookingTool := createBookingTool(httpClient)

	// Register tools with registry
	if err := toolRegistry.Register(locationTool); err != nil {
		return nil, fmt.Errorf("failed to register location tool: %w", err)
	}
	if err := toolRegistry.Register(hospitalTool); err != nil {
		return nil, fmt.Errorf("failed to register hospital tool: %w", err)
	}
	if err := toolRegistry.Register(ambulanceTool); err != nil {
		return nil, fmt.Errorf("failed to register ambulance tool: %w", err)
	}
	if err := toolRegistry.Register(bookingTool); err != nil {
		return nil, fmt.Errorf("failed to register booking tool: %w", err)
	}

	// Create classifier
	classifierConfig := triage.ClassifierConfig{
		Threshold:    0.5,
		FallbackCode: "YELLOW", // Default to YELLOW if unsure
	}
	classifier := triage.NewRuleBasedClassifier(classifierConfig)

	// Create audio processor with AI model configuration
	audioProcessor, err := createAudioProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to create audio processor: %w", err)
	}

	// Create text processor with AI model configuration
	textProcessor, err := createTextProcessor()
	if err != nil {
		return nil, fmt.Errorf("failed to create text processor: %w", err)
	}

	// Create summary generator
	summaryGenerator := &api.DefaultSummaryGenerator{}

	// Create coordinator
	coordinatorConfig := api.CoordinatorConfig{
		MaxConcurrentTools: config.GetInt("MAX_CONCURRENT_TOOLS", 5),
		Notifications: api.NotificationConfig{
			EnableSMS:     config.GetBool("ENABLE_SMS_NOTIFICATIONS", true),
			EnableEmail:   config.GetBool("ENABLE_EMAIL_NOTIFICATIONS", true),
			EnablePush:    config.GetBool("ENABLE_PUSH_NOTIFICATIONS", true),
			RetryAttempts: 3,
			RetryInterval: 5 * time.Second,
		},
		DefaultTimeout: time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
	}
	coordinator := api.NewEmergencyCoordinator(
		classifier,
		toolRegistry,
		locationTool,
		summaryGenerator,
		coordinatorConfig,
	)

	// Create API handler with both audio and text processors
	maxSize := config.GetInt("MAX_AUDIO_SIZE_MB", 20) * 1024 * 1024
	emergencyHandler := api.NewEmergencyHandler(audioProcessor, textProcessor, coordinator, int64(maxSize))

	// Create and configure HTTP mux
	mux := http.NewServeMux()
	emergencyHandler.RegisterRoutes(mux)

	return &Components{
		mux:              mux,
		toolRegistry:     toolRegistry,
		locationTool:     locationTool,
		audioProcessor:   audioProcessor,
		emergencyHandler: emergencyHandler,
	}, nil
}

// createAudioProcessor creates and configures an audio processor with AI models
func createAudioProcessor() (*api.AudioProcessor, error) {
	// Get model configuration from environment
	modelTypeStr := config.Get("AI_MODEL_TYPE", "gemini")
	var modelType ai.ModelType
	switch modelTypeStr {
	case "gemini", "GEMINI":
		modelType = ai.ModelGemini
	case "claude", "CLAUDE":
		modelType = ai.ModelClaude
	case "gpt4", "GPT4", "openai", "OPENAI":
		modelType = ai.ModelGPT4
	case "llama", "LLAMA":
		modelType = ai.ModelLlama
	default:
		modelType = ai.ModelGemini
	}

	// Set up audio processor configuration
	modelConfig := api.AudioProcessorConfig{
		ModelEndpoint:  config.Get("AI_MODEL_ENDPOINT", ""),
		APIKey:         config.Get("AI_MODEL_API_KEY", ""),
		ModelType:      modelType,
		ModelName:      config.Get("AI_MODEL_NAME", ""),
		Timeout:        time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		MaxAudioLength: 600, // 10 minutes
		Temperature:    0.7,
		MaxTokens:      4096,
	}

	// Use model-specific environment variables if the general ones aren't set
	if modelConfig.ModelEndpoint == "" {
		switch modelType {
		case ai.ModelGemini:
			modelConfig.ModelEndpoint = config.Get("GEMINI_ENDPOINT", "https://generativelanguage.googleapis.com/v1")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("GEMINI_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("GEMINI_MODEL", "gemini-1.5-pro")
			}
		case ai.ModelClaude:
			modelConfig.ModelEndpoint = config.Get("CLAUDE_ENDPOINT", "https://api.anthropic.com/v1/messages")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("CLAUDE_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("CLAUDE_MODEL", "claude-3-opus-20240229")
			}
		case ai.ModelGPT4:
			modelConfig.ModelEndpoint = config.Get("OPENAI_ENDPOINT", "https://api.openai.com/v1")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("OPENAI_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("OPENAI_MODEL", "gpt-4o")
			}
		}
	}

	return api.NewAudioProcessor(modelConfig)
}

// createTextProcessor creates and configures a text processor with AI models
func createTextProcessor() (*api.TextProcessor, error) {
	// Get model configuration from environment (reusing same config as audio processor)
	modelTypeStr := config.Get("AI_MODEL_TYPE", "gemini")
	var modelType ai.ModelType
	switch modelTypeStr {
	case "gemini", "GEMINI":
		modelType = ai.ModelGemini
	case "claude", "CLAUDE":
		modelType = ai.ModelClaude
	case "gpt4", "GPT4", "openai", "OPENAI":
		modelType = ai.ModelGPT4
	case "llama", "LLAMA":
		modelType = ai.ModelLlama
	default:
		modelType = ai.ModelGemini
	}

	// Set up text processor configuration
	modelConfig := api.TextProcessorConfig{
		ModelEndpoint: config.Get("AI_MODEL_ENDPOINT", ""),
		APIKey:        config.Get("AI_MODEL_API_KEY", ""),
		ModelType:     modelType,
		ModelName:     config.Get("AI_MODEL_NAME", ""),
		Timeout:       time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		Temperature:   0.7,
		MaxTokens:     4096,
	}

	// Use model-specific environment variables if the general ones aren't set
	if modelConfig.ModelEndpoint == "" {
		switch modelType {
		case ai.ModelGemini:
			modelConfig.ModelEndpoint = config.Get("GEMINI_ENDPOINT", "https://generativelanguage.googleapis.com/v1")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("GEMINI_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("GEMINI_MODEL", "gemini-1.5-pro")
			}
		case ai.ModelClaude:
			modelConfig.ModelEndpoint = config.Get("CLAUDE_ENDPOINT", "https://api.anthropic.com/v1/messages")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("CLAUDE_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("CLAUDE_MODEL", "claude-3-opus-20240229")
			}
		case ai.ModelGPT4:
			modelConfig.ModelEndpoint = config.Get("OPENAI_ENDPOINT", "https://api.openai.com/v1")
			if modelConfig.APIKey == "" {
				modelConfig.APIKey = config.Get("OPENAI_API_KEY", "")
			}
			if modelConfig.ModelName == "" {
				modelConfig.ModelName = config.Get("OPENAI_MODEL", "gpt-4o")
			}
		}
	}

	return api.NewTextProcessor(modelConfig)
}

// createLocationTool creates and configures a location tool
func createLocationTool(client *mockHTTPClient) *location.LocationTool {
	config := location.Config{
		APIEndpoint:   config.Get("LOCATION_API_ENDPOINT", "https://api.location.example.com"),
		APIKey:        config.Get("LOCATION_API_KEY", "mock-location-api-key"),
		Timeout:       time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		RetryAttempts: 3,
		MaxResults:    5,
		MaxDistance:   50.0, // 50km radius
	}

	// Create adapter to bridge universal client with tool-specific interface
	adapter := &location.UniversalClientAdapter{
		UniversalClient: client,
	}

	return location.NewLocationTool(config, adapter)
}

// createHospitalTool creates and configures a hospital tool
func createHospitalTool(client *mockHTTPClient) *hospital.HospitalTool {
	config := hospital.Config{
		APIEndpoint:   config.Get("HOSPITAL_API_ENDPOINT", "https://api.hospitals.example.com"),
		APIKey:        config.Get("HOSPITAL_API_KEY", "mock-hospital-api-key"),
		Timeout:       time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		RetryAttempts: 3,
	}

	// Create adapter to bridge universal client with tool-specific interface
	adapter := &hospital.UniversalClientAdapter{
		UniversalClient: client,
	}

	return hospital.NewHospitalTool(config, adapter)
}

// createAmbulanceTool creates and configures an ambulance tool
func createAmbulanceTool(client *mockHTTPClient) *ambulance.AmbulanceTool {
	config := ambulance.Config{
		APIEndpoint:   config.Get("AMBULANCE_API_ENDPOINT", "https://api.ambulance.example.com"),
		APIKey:        config.Get("AMBULANCE_API_KEY", "mock-ambulance-api-key"),
		Timeout:       time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		RetryAttempts: 3,
	}

	// Create adapter to bridge universal client with tool-specific interface
	adapter := &ambulance.UniversalClientAdapter{
		UniversalClient: client,
	}

	return ambulance.NewAmbulanceTool(config, adapter)
}

// createBookingTool creates and configures a booking tool
func createBookingTool(client *mockHTTPClient) *booking.BookingTool {
	config := booking.Config{
		APIEndpoint:   config.Get("BOOKING_API_ENDPOINT", "https://api.booking.example.com"),
		APIKey:        config.Get("BOOKING_API_KEY", "mock-booking-api-key"),
		Timeout:       time.Duration(config.GetInt("API_TIMEOUT_SECONDS", 30)) * time.Second,
		RetryAttempts: 3,
	}

	// Create adapter to bridge universal client with tool-specific interface
	adapter := &booking.UniversalClientAdapter{
		UniversalClient: client,
	}

	return booking.NewBookingTool(config, adapter)
}

// mockHTTPClient is a placeholder implementation for demonstration
type mockHTTPClient struct{}

// Do implements a unified interface method for all tool HTTP clients
func (c *mockHTTPClient) Do(req interface{}) (interface{}, error) {
	switch req.(type) {
	case *location.HTTPRequest:
		return &location.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`[{"id":"hospital-1","name":"General Hospital","type":"hospital","latitude":37.7749,"longitude":-122.4194,"address":"123 Main St"}]`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	case *hospital.HTTPRequest:
		return &hospital.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`{"success":true,"hospital_id":"hospital-1","eta":"5 minutes"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	case *ambulance.HTTPRequest:
		return &ambulance.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`{"success":true,"ambulance_id":"ambulance-1","eta":"3 minutes"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	case *booking.HTTPRequest:
		return &booking.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`{"success":true,"booking_id":"booking-1","status":"confirmed"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported request type: %T", req)
	}
}
