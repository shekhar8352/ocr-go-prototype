package prompt

import (
	"encoding/json"
	"testing"
)

func TestOCRJSONSchema(t *testing.T) {
	schema := OCRJSONSchema(PromptConfig{
		WithSummary:              true,
		WithStructuredExtraction: true,
	})

	if schema["type"] != "object" {
		t.Error("schema type should be object")
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	if _, ok := props["metadata"]; !ok {
		t.Error("metadata property missing")
	}
	if _, ok := props["text"]; !ok {
		t.Error("text property missing")
	}
}

func TestOCRJSONSchemaBytes(t *testing.T) {
	raw := OCRJSONSchemaBytes(PromptConfig{})
	if len(raw) == 0 {
		t.Fatal("schema bytes should not be empty")
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("schema bytes should be valid JSON: %v", err)
	}
}
