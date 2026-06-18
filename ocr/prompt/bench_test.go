package prompt

import "testing"

func BenchmarkBuildOCRPrompt(b *testing.B) {
	cfg := PromptConfig{
		WithSummary:              true,
		WithLanguageDetection:    true,
		WithStructuredExtraction: true,
		WithBoundingBoxes:        true,
		WithConfidenceScores:     true,
	}

	for b.Loop() {
		BuildOCRPrompt(cfg)
	}
}

func BenchmarkOCRJSONSchemaBytes(b *testing.B) {
	cfg := PromptConfig{
		WithSummary:              true,
		WithStructuredExtraction: true,
	}

	for b.Loop() {
		OCRJSONSchemaBytes(cfg)
	}
}
