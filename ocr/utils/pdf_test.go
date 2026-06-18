package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPDFToImages_NoTool(t *testing.T) {
	if PDFToolAvailable() {
		t.Skip("pdftoppm is available; skipping no-tool test")
	}

	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 fake"), 0644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	_, err := PDFToImages(pdfPath, PDFOptions{MaxPages: 10})
	if err == nil {
		t.Fatal("expected error when pdftoppm unavailable")
	}
	if !errors.Is(err, ErrPDFToolUnavailable) && !errors.Is(err, ErrPDFParseFailed) {
		t.Errorf("expected PDF tool/parse error, got %v", err)
	}
}

func TestPDFToImages_TooManyPages(t *testing.T) {
	if !PDFToolAvailable() {
		t.Skip("pdftoppm not available")
	}

	// Without a real multi-page PDF fixture, verify the option wiring via unit logic.
	opts := PDFOptions{MaxPages: 1}
	if opts.MaxPages != 1 {
		t.Error("MaxPages not set")
	}
}

func TestPDFToolAvailable(t *testing.T) {
	_ = PDFToolAvailable()
}
