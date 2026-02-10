package ocr

import "time"

const (
	// DefaultOllamaURL is the default Ollama API endpoint.
	DefaultOllamaURL = "http://localhost:11434"

	// DefaultModel is the default vision-capable Ollama model.
	DefaultModel = "llama3.2-vision"

	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 120 * time.Second

	// DefaultTemperature is the deterministic temperature for OCR.
	DefaultTemperature = 0.1

	// DefaultMaxFileSize is the maximum allowed file size (50 MB).
	DefaultMaxFileSize = 50 * 1024 * 1024

	// DefaultMaxImageDimension is the maximum allowed image dimension (pixels) per side.
	DefaultMaxImageDimension = 8192

	// MaxRetries is the number of retries if JSON parsing fails.
	MaxRetries = 1
)

// Config holds all configuration for an OCR extraction request.
type Config struct {
	// OllamaURL is the base URL for the Ollama API.
	OllamaURL string

	// Model is the Ollama model to use.
	Model string

	// Timeout is the request timeout.
	Timeout time.Duration

	// Temperature controls randomness (0 = deterministic).
	Temperature float64

	// MaxFileSize is the maximum file size in bytes.
	MaxFileSize int64

	// MaxImageDimension is the max width/height in pixels.
	MaxImageDimension int

	// Feature flags
	WithSummary              bool
	WithLanguageDetection    bool
	WithStructuredExtraction bool
	WithBoundingBoxes        bool
	WithConfidenceScores     bool
}

// DefaultConfig returns a Config with all defaults applied.
func DefaultConfig() *Config {
	return &Config{
		OllamaURL:                DefaultOllamaURL,
		Model:                    DefaultModel,
		Timeout:                  DefaultTimeout,
		Temperature:              DefaultTemperature,
		MaxFileSize:              DefaultMaxFileSize,
		MaxImageDimension:        DefaultMaxImageDimension,
		WithSummary:              false,
		WithLanguageDetection:    true,
		WithStructuredExtraction: true,
		WithBoundingBoxes:        true,
		WithConfidenceScores:     true,
	}
}
