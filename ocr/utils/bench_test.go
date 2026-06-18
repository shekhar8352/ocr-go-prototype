package utils

import (
	"testing"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

func BenchmarkCleanJSONResponse(b *testing.B) {
	raw := "```json\n{\"metadata\":{\"document_type\":\"unknown\",\"confidence_score\":0.9},\"text\":{\"raw\":\"hello\",\"lines\":[]},\"structured_data\":{\"key_value_pairs\":{},\"tables\":[]},\"summary\":null}\n```"
	for b.Loop() {
		CleanJSONResponse(raw)
	}
}

func BenchmarkValidateOCRResult(b *testing.B) {
	result := &models.OCRResult{
		Source: models.Source{
			Type:     models.SourceTypeFile,
			Path:     "/tmp/test.png",
			Checksum: "abc123",
		},
		Image: models.ImageInfo{
			Width: 100, Height: 100, ColorMode: models.ColorModeRGB,
		},
		Metadata: models.Metadata{
			DocumentType: models.DocumentTypeUnknown, ConfidenceScore: 0.9,
		},
		Text: models.TextResult{
			Raw: "hello",
			Lines: []models.TextLine{
				{Text: "hello", Confidence: 0.9},
			},
		},
		StructuredData: models.StructuredData{
			KeyValuePairs: map[string]string{},
			Tables:        []models.Table{},
		},
	}

	for b.Loop() {
		_ = ValidateOCRResult(result)
	}
}

func BenchmarkEncodeBase64(b *testing.B) {
	data := make([]byte, 1024*1024)
	for b.Loop() {
		EncodeBase64(data)
	}
}
