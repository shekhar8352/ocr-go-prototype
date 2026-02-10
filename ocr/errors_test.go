package ocr

import (
	"testing"
)

func TestOCRError(t *testing.T) {
	err := NewOCRError("Extract", "req-123", ErrEmptySource)
	expected := "ocr [req-123] Extract: ocr: source path or URL is empty"
	if err.Error() != expected {
		t.Errorf("OCRError.Error() = %q, want %q", err.Error(), expected)
	}

	// Test Unwrap
	if err.Unwrap() != ErrEmptySource {
		t.Error("OCRError.Unwrap() did not return underlying error")
	}
}

func TestOCRError_NoRequestID(t *testing.T) {
	err := WrapError("LoadImage", ErrFileNotFound)
	expected := "ocr LoadImage: ocr: file not found"
	if err.Error() != expected {
		t.Errorf("OCRError.Error() = %q, want %q", err.Error(), expected)
	}
}
