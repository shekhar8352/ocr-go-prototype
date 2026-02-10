// Package prompt provides versioned OCR prompt templates for Ollama vision models.
package prompt

import (
	"strings"
)

const (
	// PromptVersion is the current version of the OCR prompt template.
	PromptVersion = "1.0.0"
)

// PromptConfig controls what the prompt asks the model to extract.
type PromptConfig struct {
	WithSummary              bool
	WithLanguageDetection    bool
	WithStructuredExtraction bool
	WithBoundingBoxes        bool
	WithConfidenceScores     bool
}

// BuildOCRPrompt constructs the deterministic OCR prompt for Ollama vision models.
// The prompt strictly enforces JSON-only output with the exact required schema.
func BuildOCRPrompt(cfg PromptConfig) string {
	var sb strings.Builder

	sb.WriteString(`You are a precise OCR engine. Analyze the provided image and extract all text content.

CRITICAL INSTRUCTIONS:
- Respond ONLY with valid JSON.
- No markdown. No code fences. No explanations. No comments.
- Do NOT wrap the JSON in backticks or any markup.
- Output MUST start with { and end with }
- Every string value must be properly escaped.

You must return a JSON object with EXACTLY this structure:

{
  "metadata": {
    "language": `)

	if cfg.WithLanguageDetection {
		sb.WriteString(`"<detected ISO 639-1 language code or null if unknown>",`)
	} else {
		sb.WriteString(`null,`)
	}

	sb.WriteString(`
    "document_type": "<one of: invoice, receipt, id_card, contract, unknown>",
    "confidence_score": <float between 0.0 and 1.0 representing overall OCR confidence>
  },
  "text": {
    "raw": "<all extracted text as a single string, preserving line breaks with \\n>",
    "lines": [
      {
        "text": "<text content of this line>",`)

	if cfg.WithBoundingBoxes {
		sb.WriteString(`
        "bounding_box": {
          "x": <estimated x coordinate>,
          "y": <estimated y coordinate>,
          "width": <estimated width>,
          "height": <estimated height>
        },`)
	} else {
		sb.WriteString(`
        "bounding_box": null,`)
	}

	if cfg.WithConfidenceScores {
		sb.WriteString(`
        "confidence": <float between 0.0 and 1.0>`)
	} else {
		sb.WriteString(`
        "confidence": 0.0`)
	}

	sb.WriteString(`
      }
    ]
  },`)

	if cfg.WithStructuredExtraction {
		sb.WriteString(`
  "structured_data": {
    "key_value_pairs": {
      "<key>": "<value>"
    },
    "tables": [
      {
        "headers": ["<column header 1>", "<column header 2>"],
        "rows": [["<cell 1>", "<cell 2>"]]
      }
    ]
  },`)
	} else {
		sb.WriteString(`
  "structured_data": {
    "key_value_pairs": {},
    "tables": []
  },`)
	}

	if cfg.WithSummary {
		sb.WriteString(`
  "summary": "<brief natural language summary of the document content>"`)
	} else {
		sb.WriteString(`
  "summary": null`)
	}

	sb.WriteString(`
}

RULES:
1. Extract ALL visible text from the image, missing nothing.
2. "document_type" MUST be exactly one of: "invoice", "receipt", "id_card", "contract", "unknown".
3. If no tables are found, return "tables": [].
4. If no key-value pairs are found, return "key_value_pairs": {}.
5. "lines" must contain every line of text found, even if only one.`)

	if cfg.WithBoundingBoxes {
		sb.WriteString(`
6. Estimate bounding boxes as best as possible based on text position in the image.`)
	}

	if cfg.WithLanguageDetection {
		sb.WriteString(`
7. Detect the primary language of the document and use ISO 639-1 codes (e.g., "en", "fr", "de").`)
	}

	sb.WriteString(`

Remember: Output ONLY the JSON object. Nothing else.`)

	return sb.String()
}
