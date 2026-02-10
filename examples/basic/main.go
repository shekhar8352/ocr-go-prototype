// Command example demonstrates the usage of the OCR package.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image-path-or-url> [model]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s /path/to/invoice.png\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s /path/to/receipt.jpg llama3.2-vision\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s https://example.com/document.png\n", os.Args[0])
		os.Exit(1)
	}

	source := os.Args[1]
	model := "llama3.2-vision"
	if len(os.Args) > 2 {
		model = os.Args[2]
	}

	ctx := context.Background()

	fmt.Fprintf(os.Stderr, "Processing: %s\n", source)
	fmt.Fprintf(os.Stderr, "Model: %s\n", model)
	fmt.Fprintf(os.Stderr, "---\n")

	result, err := ocr.Extract(ctx, source,
		ocr.WithModel(model),
		ocr.WithSummary(true),
		ocr.WithLanguageDetection(true),
		ocr.WithStructuredExtraction(true),
		ocr.WithBoundingBoxes(true),
		ocr.WithConfidenceScores(true),
		ocr.WithTimeout(180*time.Second),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Output strict JSON to stdout
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
