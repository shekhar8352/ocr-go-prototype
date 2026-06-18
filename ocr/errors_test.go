package ocr

import (
	"errors"
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

func TestErrorsIsAndAs(t *testing.T) {
	err := NewOCRError("Extract", "req-123", ErrEmptySource)

	if !errors.Is(err, ErrEmptySource) {
		t.Error("errors.Is should find sentinel through OCRError")
	}

	var ocrErr *OCRError
	if !errors.As(err, &ocrErr) {
		t.Fatal("errors.As should find OCRError")
	}
	if ocrErr.Op != "Extract" {
		t.Errorf("Op = %q, want Extract", ocrErr.Op)
	}
	if ocrErr.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want req-123", ocrErr.RequestID)
	}
}
