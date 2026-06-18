package ocr_test

import (
	"context"
	"fmt"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr"
)

func ExampleExtract() {
	ctx := context.Background()

	result, err := ocr.Extract(ctx, "/path/to/document.png",
		ocr.WithSummary(true),
		ocr.WithModel("llama3.2-vision"),
	)
	if err != nil {
		fmt.Printf("OCR failed: %v\n", err)
		return
	}

	fmt.Printf("Extracted %d characters\n", len(result.Text.Raw))
	// Output is environment-dependent; this example documents usage only.
}
