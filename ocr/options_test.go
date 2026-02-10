package ocr

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.OllamaURL != DefaultOllamaURL {
		t.Errorf("OllamaURL = %q, want %q", cfg.OllamaURL, DefaultOllamaURL)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.Temperature != DefaultTemperature {
		t.Errorf("Temperature = %v, want %v", cfg.Temperature, DefaultTemperature)
	}
	if cfg.MaxFileSize != DefaultMaxFileSize {
		t.Errorf("MaxFileSize = %d, want %d", cfg.MaxFileSize, DefaultMaxFileSize)
	}
}

func TestOptions(t *testing.T) {
	cfg := DefaultConfig()

	opts := []Option{
		WithModel("minicpm-v"),
		WithTimeout(30 * time.Second),
		WithSummary(true),
		WithLanguageDetection(false),
		WithStructuredExtraction(false),
		WithBoundingBoxes(false),
		WithConfidenceScores(false),
		WithOllamaURL("http://custom:11434"),
		WithTemperature(0.0),
		WithMaxFileSize(1024),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Model != "minicpm-v" {
		t.Errorf("Model = %q, want %q", cfg.Model, "minicpm-v")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if !cfg.WithSummary {
		t.Error("WithSummary should be true")
	}
	if cfg.WithLanguageDetection {
		t.Error("WithLanguageDetection should be false")
	}
	if cfg.WithStructuredExtraction {
		t.Error("WithStructuredExtraction should be false")
	}
	if cfg.WithBoundingBoxes {
		t.Error("WithBoundingBoxes should be false")
	}
	if cfg.WithConfidenceScores {
		t.Error("WithConfidenceScores should be false")
	}
	if cfg.OllamaURL != "http://custom:11434" {
		t.Errorf("OllamaURL = %q, want %q", cfg.OllamaURL, "http://custom:11434")
	}
	if cfg.Temperature != 0.0 {
		t.Errorf("Temperature = %v, want %v", cfg.Temperature, 0.0)
	}
	if cfg.MaxFileSize != 1024 {
		t.Errorf("MaxFileSize = %d, want %d", cfg.MaxFileSize, 1024)
	}
}

func TestOptionEdgeCases(t *testing.T) {
	cfg := DefaultConfig()

	// Empty model should not override
	WithModel("")(cfg)
	if cfg.Model != DefaultModel {
		t.Error("empty model should not override default")
	}

	// Zero timeout should not override
	WithTimeout(0)(cfg)
	if cfg.Timeout != DefaultTimeout {
		t.Error("zero timeout should not override default")
	}

	// Negative timeout should not override
	WithTimeout(-1 * time.Second)(cfg)
	if cfg.Timeout != DefaultTimeout {
		t.Error("negative timeout should not override default")
	}

	// Invalid temperature should not override
	WithTemperature(-1)(cfg)
	if cfg.Temperature != DefaultTemperature {
		t.Error("negative temperature should not override default")
	}
	WithTemperature(3)(cfg)
	if cfg.Temperature != DefaultTemperature {
		t.Error("temperature > 2 should not override default")
	}

	// Zero max file size should not override
	WithMaxFileSize(0)(cfg)
	if cfg.MaxFileSize != DefaultMaxFileSize {
		t.Error("zero max file size should not override default")
	}
}
