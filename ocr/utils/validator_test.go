package utils

import (
	"testing"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

func TestValidateOCRResult_Valid(t *testing.T) {
	result := &models.OCRResult{
		Source: models.Source{
			Type:     models.SourceTypeFile,
			Path:     "/path/to/image.png",
			Checksum: "abc123",
		},
		Image: models.ImageInfo{
			Width:     800,
			Height:    600,
			DPI:       nil,
			ColorMode: models.ColorModeRGB,
		},
		Metadata: models.Metadata{
			Language:        strPtr("en"),
			DocumentType:    models.DocumentTypeInvoice,
			ConfidenceScore: 0.95,
		},
		Text: models.TextResult{
			Raw: "Sample text",
			Lines: []models.TextLine{
				{
					Text:       "Sample text",
					Confidence: 0.9,
				},
			},
		},
		StructuredData: models.StructuredData{
			KeyValuePairs: map[string]string{"key": "value"},
			Tables:        []models.Table{},
		},
		Summary: strPtr("A test document"),
	}

	if err := ValidateOCRResult(result); err != nil {
		t.Errorf("ValidateOCRResult: unexpected error: %v", err)
	}
}

func TestValidateOCRResult_Nil(t *testing.T) {
	if err := ValidateOCRResult(nil); err == nil {
		t.Fatal("expected error for nil result")
	}
}

func TestValidateOCRResult_InvalidSourceType(t *testing.T) {
	result := validResult()
	result.Source.Type = "invalid"

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for invalid source type")
	}
}

func TestValidateOCRResult_EmptyPath(t *testing.T) {
	result := validResult()
	result.Source.Path = ""

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestValidateOCRResult_EmptyChecksum(t *testing.T) {
	result := validResult()
	result.Source.Checksum = ""

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for empty checksum")
	}
}

func TestValidateOCRResult_InvalidDocumentType(t *testing.T) {
	result := validResult()
	result.Metadata.DocumentType = "banana"

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for invalid document type")
	}
}

func TestValidateOCRResult_ConfidenceOutOfRange(t *testing.T) {
	result := validResult()
	result.Metadata.ConfidenceScore = 1.5

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for confidence > 1")
	}
}

func TestValidateOCRResult_InvalidColorMode(t *testing.T) {
	result := validResult()
	result.Image.ColorMode = "INVALID"

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for invalid color mode")
	}
}

func TestValidateOCRResult_NilKeyValuePairs(t *testing.T) {
	result := validResult()
	result.StructuredData.KeyValuePairs = nil

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for nil key_value_pairs")
	}
}

func TestValidateOCRResult_NilTables(t *testing.T) {
	result := validResult()
	result.StructuredData.Tables = nil

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for nil tables")
	}
}

func TestValidateOCRResult_LineConfidenceOutOfRange(t *testing.T) {
	result := validResult()
	result.Text.Lines[0].Confidence = -0.1

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for line confidence < 0")
	}
}

func TestValidateOCRResult_EmptyLineText(t *testing.T) {
	result := validResult()
	result.Text.Lines[0].Text = ""

	if err := ValidateOCRResult(result); err == nil {
		t.Fatal("expected error for empty line text")
	}
}

func TestParseAndValidateJSON_Valid(t *testing.T) {
	input := `{
		"metadata": {
			"language": "en",
			"document_type": "invoice",
			"confidence_score": 0.95
		},
		"text": {
			"raw": "Hello World",
			"lines": [
				{
					"text": "Hello World",
					"bounding_box": null,
					"confidence": 0.9
				}
			]
		},
		"structured_data": {
			"key_value_pairs": {"key": "value"},
			"tables": []
		},
		"summary": "A test"
	}`

	resp, err := ParseAndValidateJSON(input)
	if err != nil {
		t.Fatalf("ParseAndValidateJSON: %v", err)
	}

	if resp.Metadata == nil {
		t.Fatal("metadata is nil")
	}
	if resp.Metadata.DocumentType != "invoice" {
		t.Errorf("document_type = %q, want %q", resp.Metadata.DocumentType, "invoice")
	}
	if resp.Text == nil {
		t.Fatal("text is nil")
	}
	if resp.Text.Raw != "Hello World" {
		t.Errorf("raw = %q, want %q", resp.Text.Raw, "Hello World")
	}
}

func TestParseAndValidateJSON_WithMarkdownFences(t *testing.T) {
	input := "```json\n{\"metadata\": {\"document_type\": \"receipt\", \"confidence_score\": 0.8}, \"text\": {\"raw\": \"test\", \"lines\": []}}\n```"

	resp, err := ParseAndValidateJSON(input)
	if err != nil {
		t.Fatalf("ParseAndValidateJSON with markdown: %v", err)
	}

	if resp.Metadata == nil || resp.Metadata.DocumentType != "receipt" {
		t.Error("failed to parse JSON wrapped in markdown fences")
	}
}

func TestParseAndValidateJSON_Invalid(t *testing.T) {
	_, err := ParseAndValidateJSON("not json at all")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "markdown json fence",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "markdown fence",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "extra text before JSON",
			input:    "Here is the result:\n{\"key\": \"value\"}",
			expected: `{"key": "value"}`,
		},
		{
			name:     "extra text after JSON",
			input:    "{\"key\": \"value\"}\nDone!",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanJSONResponse(tt.input)
			if got != tt.expected {
				t.Errorf("CleanJSONResponse(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// Helper functions

func validResult() *models.OCRResult {
	return &models.OCRResult{
		Source: models.Source{
			Type:     models.SourceTypeFile,
			Path:     "/test.png",
			Checksum: "abc123",
		},
		Image: models.ImageInfo{
			Width:     800,
			Height:    600,
			DPI:       nil,
			ColorMode: models.ColorModeRGB,
		},
		Metadata: models.Metadata{
			Language:        strPtr("en"),
			DocumentType:    models.DocumentTypeInvoice,
			ConfidenceScore: 0.95,
		},
		Text: models.TextResult{
			Raw: "Test",
			Lines: []models.TextLine{
				{
					Text:       "Test",
					Confidence: 0.9,
				},
			},
		},
		StructuredData: models.StructuredData{
			KeyValuePairs: map[string]string{},
			Tables:        []models.Table{},
		},
		Summary: nil,
	}
}

func strPtr(s string) *string {
	return &s
}
