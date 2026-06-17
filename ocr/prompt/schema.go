package prompt

import "encoding/json"

// OCRJSONSchema returns the JSON schema object for Ollama structured output.
func OCRJSONSchema(cfg PromptConfig) map[string]any {
	properties := map[string]any{
		"metadata": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"language": map[string]any{
					"type": []string{"string", "null"},
				},
				"document_type": map[string]any{
					"type": "string",
					"enum": []string{"invoice", "receipt", "id_card", "contract", "unknown"},
				},
				"confidence_score": map[string]any{
					"type":    "number",
					"minimum": 0,
					"maximum": 1,
				},
			},
			"required": []string{"document_type", "confidence_score"},
		},
		"text": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"raw": map[string]any{"type": "string"},
				"lines": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"text":       map[string]any{"type": "string"},
							"confidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1},
							"bounding_box": map[string]any{
								"type": []string{"object", "null"},
								"properties": map[string]any{
									"x":      map[string]any{"type": "number"},
									"y":      map[string]any{"type": "number"},
									"width":  map[string]any{"type": "number"},
									"height": map[string]any{"type": "number"},
								},
							},
						},
						"required": []string{"text"},
					},
				},
			},
			"required": []string{"raw", "lines"},
		},
		"structured_data": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key_value_pairs": map[string]any{"type": "object"},
				"tables": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"headers": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							"rows": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							},
						},
					},
				},
			},
			"required": []string{"key_value_pairs", "tables"},
		},
	}

	if cfg.WithSummary {
		properties["summary"] = map[string]any{
			"type": []string{"string", "null"},
		}
	} else {
		properties["summary"] = map[string]any{
			"type": "null",
		}
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   []string{"metadata", "text", "structured_data", "summary"},
	}
}

// OCRJSONSchemaBytes returns the JSON schema as raw JSON for Ollama's format field.
func OCRJSONSchemaBytes(cfg PromptConfig) json.RawMessage {
	schema := OCRJSONSchema(cfg)
	data, err := json.Marshal(schema)
	if err != nil {
		return json.RawMessage(`"json"`)
	}
	return data
}
