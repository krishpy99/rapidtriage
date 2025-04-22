package ai

import "errors"

// Standard errors for AI model operations
var (
	// ErrUnsupportedModel is returned when an unsupported model type is requested
	ErrUnsupportedModel = errors.New("unsupported model type")

	// ErrUnsupportedRequestType is returned when the model doesn't support the request type
	ErrUnsupportedRequestType = errors.New("unsupported request type for this model")

	// ErrInvalidConfiguration is returned when the model configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid model configuration")

	// ErrAPICallFailed is returned when the API call to the model fails
	ErrAPICallFailed = errors.New("API call to model failed")

	// ErrContextDeadlineExceeded is returned when the context deadline is exceeded
	ErrContextDeadlineExceeded = errors.New("context deadline exceeded")

	// ErrInvalidJSONSchema is returned when the JSON schema is invalid
	ErrInvalidJSONSchema = errors.New("invalid JSON schema")

	// ErrInvalidAudioFormat is returned when the audio format is not supported
	ErrInvalidAudioFormat = errors.New("invalid audio format")

	// ErrModelUnavailable is returned when the model is unavailable
	ErrModelUnavailable = errors.New("model temporarily unavailable")

	// ErrRateLimitExceeded is returned when the API rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
