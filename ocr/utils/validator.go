package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

// ValidDocumentTypes is the set of allowed document type values.
var ValidDocumentTypes = map[models.DocumentType]bool{
	models.DocumentTypeInvoice:  true,
	models.DocumentTypeReceipt:  true,
	models.DocumentTypeIDCard:   true,
	models.DocumentTypeContract: true,
	models.DocumentTypeUnknown:  true,
}

// ValidColorModes is the set of allowed color mode values.
var ValidColorModes = map[models.ColorMode]bool{
	models.ColorModeRGB:       true,
	models.ColorModeGrayscale: true,
	models.ColorModeCMYK:      true,
	models.ColorModeUnknown:   true,
}

// ValidateOCRResult validates that an OCRResult conforms to the strict schema.
func ValidateOCRResult(result *models.OCRResult) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Validate source
	if result.Source.Type != models.SourceTypeFile && result.Source.Type != models.SourceTypeURL {
		return fmt.Errorf("invalid source type: %q", result.Source.Type)
	}
	if result.Source.Path == "" {
		return fmt.Errorf("source path is empty")
	}
	if result.Source.Checksum == "" {
		return fmt.Errorf("source checksum is empty")
	}

	// Validate metadata
	if !ValidDocumentTypes[result.Metadata.DocumentType] {
		return fmt.Errorf("invalid document type: %q", result.Metadata.DocumentType)
	}
	if result.Metadata.ConfidenceScore < 0 || result.Metadata.ConfidenceScore > 1 {
		return fmt.Errorf("confidence_score out of range [0, 1]: %f", result.Metadata.ConfidenceScore)
	}

	// Validate image
	if !ValidColorModes[result.Image.ColorMode] {
		return fmt.Errorf("invalid color mode: %q", result.Image.ColorMode)
	}

	// Validate text lines
	for i, line := range result.Text.Lines {
		if line.Text == "" {
			return fmt.Errorf("text line %d has empty text", i)
		}
		if line.Confidence < 0 || line.Confidence > 1 {
			return fmt.Errorf("text line %d confidence out of range [0, 1]: %f", i, line.Confidence)
		}
	}

	// Validate structured data
	if result.StructuredData.KeyValuePairs == nil {
		return fmt.Errorf("structured_data.key_value_pairs is nil (should be empty map)")
	}
	if result.StructuredData.Tables == nil {
		return fmt.Errorf("structured_data.tables is nil (should be empty slice)")
	}

	return nil
}

// ParseAndValidateJSON attempts to unmarshal raw JSON into an OllamaVisionResponse.
// It first strips any markdown code fences the model may have included.
func ParseAndValidateJSON(raw string) (*models.OllamaVisionResponse, error) {
	cleaned := CleanJSONResponse(raw)

	var resp models.OllamaVisionResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &resp, nil
}

// CleanJSONResponse strips markdown code fences and extraneous text from model output,
// extracting only the JSON object.
func CleanJSONResponse(raw string) string {
	s := strings.TrimSpace(raw)

	// Remove markdown code fences
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Find the first { and last } to extract the JSON object
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		s = s[start : end+1]
	}

	return s
}
