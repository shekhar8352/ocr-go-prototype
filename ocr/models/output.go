// Package models defines the strict output types for the OCR package.
// All structs map directly to the mandatory JSON schema.
package models

// OCRResult is the top-level output of an OCR extraction.
// Every field is strictly typed and maps 1:1 to the required JSON schema.
type OCRResult struct {
	Source         Source         `json:"source"`
	Image          ImageInfo      `json:"image"`
	Metadata       Metadata       `json:"metadata"`
	Text           TextResult     `json:"text"`
	StructuredData StructuredData `json:"structured_data"`
	Summary        *string        `json:"summary"`
}

// Source describes how the image was provided.
type Source struct {
	Type     SourceType `json:"type"`
	Path     string     `json:"path"`
	Checksum string     `json:"checksum"`
}

// SourceType is an enum for source types.
type SourceType string

const (
	SourceTypeFile SourceType = "file"
	SourceTypeURL  SourceType = "url"
)

// ImageInfo holds metadata about the image itself.
type ImageInfo struct {
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	DPI       *int       `json:"dpi"`
	ColorMode ColorMode  `json:"color_mode"`
}

// ColorMode is an enum for color modes.
type ColorMode string

const (
	ColorModeRGB       ColorMode = "RGB"
	ColorModeGrayscale ColorMode = "Grayscale"
	ColorModeCMYK      ColorMode = "CMYK"
	ColorModeUnknown   ColorMode = "Unknown"
)

// Metadata holds document-level metadata inferred by the model.
type Metadata struct {
	Language        *string      `json:"language"`
	DocumentType    DocumentType `json:"document_type"`
	ConfidenceScore float64      `json:"confidence_score"`
}

// DocumentType is an enum for document types.
type DocumentType string

const (
	DocumentTypeInvoice  DocumentType = "invoice"
	DocumentTypeReceipt  DocumentType = "receipt"
	DocumentTypeIDCard   DocumentType = "id_card"
	DocumentTypeContract DocumentType = "contract"
	DocumentTypeUnknown  DocumentType = "unknown"
)

// TextResult holds the OCR text output.
type TextResult struct {
	Raw   string     `json:"raw"`
	Lines []TextLine `json:"lines"`
}

// TextLine is a single line detected during OCR.
type TextLine struct {
	Text        string       `json:"text"`
	BoundingBox *BoundingBox `json:"bounding_box"`
	Confidence  float64      `json:"confidence"`
}

// BoundingBox is a rectangular region in the image.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// StructuredData holds tables and key-value pairs extracted from the document.
type StructuredData struct {
	KeyValuePairs map[string]string `json:"key_value_pairs"`
	Tables        []Table           `json:"tables"`
}

// Table is a single table detected in the document.
type Table struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// OllamaVisionResponse is the intermediate struct for parsing the Ollama model's JSON response.
// It mirrors OCRResult but uses more forgiving types to handle model quirks before strict validation.
type OllamaVisionResponse struct {
	Metadata       *OllamaMetadata       `json:"metadata,omitempty"`
	Text           *OllamaTextResult     `json:"text,omitempty"`
	StructuredData *OllamaStructuredData `json:"structured_data,omitempty"`
	Summary        *string               `json:"summary,omitempty"`
	Image          *OllamaImageInfo      `json:"image,omitempty"`
}

// OllamaMetadata is the forgiving metadata from Ollama.
type OllamaMetadata struct {
	Language        *string `json:"language,omitempty"`
	DocumentType    string  `json:"document_type,omitempty"`
	ConfidenceScore float64 `json:"confidence_score,omitempty"`
}

// OllamaTextResult is the forgiving text result from Ollama.
type OllamaTextResult struct {
	Raw   string           `json:"raw,omitempty"`
	Lines []OllamaTextLine `json:"lines,omitempty"`
}

// OllamaTextLine is a forgiving text line from Ollama.
type OllamaTextLine struct {
	Text        string       `json:"text,omitempty"`
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"`
	Confidence  float64      `json:"confidence,omitempty"`
}

// OllamaStructuredData is the forgiving structured data from Ollama.
type OllamaStructuredData struct {
	KeyValuePairs map[string]string `json:"key_value_pairs,omitempty"`
	Tables        []Table           `json:"tables,omitempty"`
}

// OllamaImageInfo is the forgiving image info from Ollama.
type OllamaImageInfo struct {
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	DPI       *int   `json:"dpi,omitempty"`
	ColorMode string `json:"color_mode,omitempty"`
}
