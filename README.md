# OCR Go Prototype

A production-ready Go package for performing OCR (Optical Character Recognition) and document image understanding using locally running Ollama vision models.

## Features

- **Single function API** — call `ocr.Extract()` with a file path or URL
- **Strict JSON output** — every response conforms to a deterministic schema
- **Local-only processing** — no cloud APIs, no external services
- **Multi-format support** — PNG, JPG, JPEG, PDF (page-by-page)
- **Configurable** — functional options for model, timeout, feature flags
- **Production-ready** — typed errors, structured logging, request tracing, input validation
- **SSRF protection** — URL sanitization blocks private/internal networks
- **Retry logic** — automatic retry on JSON parse failure

## Prerequisites

- **Go 1.21+**
- **[Ollama](https://ollama.ai)** running locally at `http://localhost:11434`
- A vision-capable model pulled in Ollama:

```bash
ollama pull llama3.2-vision
```

- **Optional**: `pdftoppm` (from `poppler-utils`) for multi-page PDF support

```bash
# macOS
brew install poppler

# Ubuntu/Debian
sudo apt-get install poppler-utils
```

## Installation

```bash
go get github.com/sudhanshushekhar/ocr-go-prototype/ocr
```

## Quick Start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/sudhanshushekhar/ocr-go-prototype/ocr"
)

func main() {
    ctx := context.Background()

    result, err := ocr.Extract(ctx, "/path/to/invoice.png",
        ocr.WithSummary(true),
        ocr.WithModel("llama3.2-vision"),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
        os.Exit(1)
    }

    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    enc.Encode(result)
}
```

## API

### `ocr.Extract`

```go
func Extract(
    ctx context.Context,
    source string,       // Local file path OR remote URL
    opts ...Option,      // Functional options
) (*models.OCRResult, error)
```

### Options

| Option                           | Description                           | Default           |
| -------------------------------- | ------------------------------------- | ----------------- |
| `WithModel(string)`              | Ollama model name                     | `llama3.2-vision` |
| `WithTimeout(time.Duration)`     | Request timeout                       | `120s`            |
| `WithSummary(bool)`              | Include natural language summary      | `false`           |
| `WithLanguageDetection(bool)`    | Detect document language              | `true`            |
| `WithStructuredExtraction(bool)` | Extract tables + key-value pairs      | `true`            |
| `WithBoundingBoxes(bool)`        | Include bounding box coordinates      | `true`            |
| `WithConfidenceScores(bool)`     | Include OCR confidence scores         | `true`            |
| `WithOllamaURL(string)`          | Custom Ollama API endpoint            | `localhost:11434` |
| `WithTemperature(float64)`       | Model temperature (0 = deterministic) | `0.1`             |
| `WithMaxFileSize(int64)`         | Maximum file size in bytes            | `50 MB`           |

### Output Schema

Every response is strictly typed and conforms to this JSON structure:

```json
{
  "source": {
    "type": "file | url",
    "path": "string",
    "checksum": "sha256"
  },
  "image": {
    "width": 0,
    "height": 0,
    "dpi": null,
    "color_mode": "RGB | Grayscale | CMYK | Unknown"
  },
  "metadata": {
    "language": "string | null",
    "document_type": "invoice | receipt | id_card | contract | unknown",
    "confidence_score": 0.0
  },
  "text": {
    "raw": "string",
    "lines": [
      {
        "text": "string",
        "bounding_box": {
          "x": 0,
          "y": 0,
          "width": 0,
          "height": 0
        },
        "confidence": 0.0
      }
    ]
  },
  "structured_data": {
    "key_value_pairs": {},
    "tables": []
  },
  "summary": "string | null"
}
```

## Package Structure

```
ocr/
├── client/
│   └── ollama.go           # Ollama HTTP client
│   └── ollama_test.go
├── engine/
│   └── vision.go           # OCR orchestration + retry logic
├── models/
│   └── output.go           # Strict output structs
├── prompt/
│   └── ocr_prompt.go       # Versioned prompt templates
│   └── ocr_prompt_test.go
├── utils/
│   ├── hash.go             # SHA-256 checksums
│   ├── hash_test.go
│   ├── image.go            # Image loading, validation, SSRF protection
│   ├── image_test.go
│   ├── pdf.go              # PDF-to-image conversion
│   ├── validator.go        # JSON + schema validation
│   └── validator_test.go
├── config.go               # Configuration with defaults
├── errors.go               # Typed errors
├── errors_test.go
├── ocr.go                  # Public API (Extract function)
├── options.go              # Functional options
└── options_test.go
```

## Running Tests

```bash
go test ./... -v
```

## Running the Example

```bash
# Ensure Ollama is running with a vision model
ollama pull llama3.2-vision

# Run the example
go run examples/basic/main.go /path/to/image.png

# With a specific model
go run examples/basic/main.go /path/to/receipt.jpg minicpm-v
```

## Supported Models

Any vision-capable Ollama model works. Recommended:

| Model             | Size   | Notes                                 |
| ----------------- | ------ | ------------------------------------- |
| `llama3.2-vision` | ~4.7GB | Best overall quality                  |
| `minicpm-v`       | ~5.5GB | Good for structured documents         |
| `moondream`       | ~1.7GB | Lightweight, faster but less accurate |

## Error Handling

All errors are typed and can be inspected:

```go
result, err := ocr.Extract(ctx, source)
if err != nil {
    var ocrErr *ocr.OCRError
    if errors.As(err, &ocrErr) {
        fmt.Println("Operation:", ocrErr.Op)
        fmt.Println("Request ID:", ocrErr.RequestID)
        fmt.Println("Cause:", ocrErr.Unwrap())
    }
}
```

Sentinel errors: `ErrUnsupportedFormat`, `ErrFileTooLarge`, `ErrInvalidURL`, `ErrFileNotFound`, `ErrOllamaUnavailable`, `ErrInvalidJSONResponse`, and more.

## Logging

Structured JSON logs are written to stderr with:

- `request_id` — unique per extraction call
- `model` — which Ollama model was used
- `latency` — total processing time
- `prompt_eval_count` / `eval_count` — token counts
- No sensitive data (file contents, extracted text) is logged

## License

MIT
