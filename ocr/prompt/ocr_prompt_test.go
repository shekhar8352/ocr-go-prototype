package prompt

import (
	"strings"
	"testing"
)

func TestBuildOCRPrompt_AllEnabled(t *testing.T) {
	cfg := PromptConfig{
		WithSummary:              true,
		WithLanguageDetection:    true,
		WithStructuredExtraction: true,
		WithBoundingBoxes:        true,
		WithConfidenceScores:     true,
	}

	prompt := BuildOCRPrompt(cfg)

	requiredPhrases := []string{
		"Respond ONLY with valid JSON",
		"No markdown",
		"No explanations",
		"document_type",
		"key_value_pairs",
		"tables",
		"bounding_box",
		"confidence",
		"summary",
		"language",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(prompt, phrase) {
			t.Errorf("prompt missing required phrase: %q", phrase)
		}
	}
}

func TestBuildOCRPrompt_AllDisabled(t *testing.T) {
	cfg := PromptConfig{
		WithSummary:              false,
		WithLanguageDetection:    false,
		WithStructuredExtraction: false,
		WithBoundingBoxes:        false,
		WithConfidenceScores:     false,
	}

	prompt := BuildOCRPrompt(cfg)

	// Should still contain JSON instruction
	if !strings.Contains(prompt, "Respond ONLY with valid JSON") {
		t.Error("prompt missing JSON instruction even with all flags disabled")
	}

	// Summary should be null
	if !strings.Contains(prompt, `"summary": null`) {
		t.Error("prompt should include null summary when WithSummary is false")
	}
}

func TestBuildOCRPrompt_ContainsSchema(t *testing.T) {
	cfg := PromptConfig{
		WithSummary:              true,
		WithLanguageDetection:    true,
		WithStructuredExtraction: true,
		WithBoundingBoxes:        true,
		WithConfidenceScores:     true,
	}

	prompt := BuildOCRPrompt(cfg)

	// Must contain key schema elements
	schemaElements := []string{
		"metadata",
		"text",
		"raw",
		"lines",
		"structured_data",
	}

	for _, elem := range schemaElements {
		if !strings.Contains(prompt, elem) {
			t.Errorf("prompt missing schema element: %q", elem)
		}
	}
}

func TestPromptVersion(t *testing.T) {
	if PromptVersion == "" {
		t.Fatal("PromptVersion is empty")
	}
}
