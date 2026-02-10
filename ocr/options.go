package ocr

import "time"

// Option is a functional option for configuring OCR extraction.
type Option func(*Config)

// WithSummary enables or disables natural language summary in the output.
func WithSummary(enabled bool) Option {
	return func(c *Config) {
		c.WithSummary = enabled
	}
}

// WithLanguageDetection enables or disables language detection.
func WithLanguageDetection(enabled bool) Option {
	return func(c *Config) {
		c.WithLanguageDetection = enabled
	}
}

// WithStructuredExtraction enables or disables tables + key-value pair extraction.
func WithStructuredExtraction(enabled bool) Option {
	return func(c *Config) {
		c.WithStructuredExtraction = enabled
	}
}

// WithBoundingBoxes enables or disables bounding box coordinates for text lines.
func WithBoundingBoxes(enabled bool) Option {
	return func(c *Config) {
		c.WithBoundingBoxes = enabled
	}
}

// WithConfidenceScores enables or disables confidence scores for text lines.
func WithConfidenceScores(enabled bool) Option {
	return func(c *Config) {
		c.WithConfidenceScores = enabled
	}
}

// WithModel sets the Ollama model to use for OCR.
func WithModel(model string) Option {
	return func(c *Config) {
		if model != "" {
			c.Model = model
		}
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.Timeout = d
		}
	}
}

// WithOllamaURL sets a custom Ollama API endpoint.
func WithOllamaURL(url string) Option {
	return func(c *Config) {
		if url != "" {
			c.OllamaURL = url
		}
	}
}

// WithTemperature sets the model temperature.
func WithTemperature(t float64) Option {
	return func(c *Config) {
		if t >= 0 && t <= 2 {
			c.Temperature = t
		}
	}
}

// WithMaxFileSize sets the maximum allowed file size in bytes.
func WithMaxFileSize(size int64) Option {
	return func(c *Config) {
		if size > 0 {
			c.MaxFileSize = size
		}
	}
}
