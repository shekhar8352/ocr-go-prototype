//go:build integration

package ocr

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestExtractIntegration runs against a real Ollama instance.
// Run with: go test -tags=integration ./ocr -run TestExtractIntegration -v
func TestExtractIntegration(t *testing.T) {
	imagePath := os.Getenv("OCR_TEST_IMAGE")
	if imagePath == "" {
		t.Skip("set OCR_TEST_IMAGE to run integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := Extract(ctx, imagePath)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if result.Text.Raw == "" {
		t.Error("expected non-empty text from integration run")
	}
}
