// Package ocr provides typed errors for the OCR package.
package ocr

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure modes.
var (
	ErrUnsupportedFormat   = errors.New("ocr: unsupported file format")
	ErrFileTooLarge        = errors.New("ocr: file exceeds maximum allowed size")
	ErrInvalidURL          = errors.New("ocr: invalid or unsafe URL")
	ErrFileNotFound        = errors.New("ocr: file not found")
	ErrFileReadFailed      = errors.New("ocr: failed to read file")
	ErrImageDecodeFailed   = errors.New("ocr: failed to decode image")
	ErrPDFParseFailed      = errors.New("ocr: failed to parse PDF")
	ErrOllamaUnavailable   = errors.New("ocr: ollama server is unavailable")
	ErrOllamaRequestFailed = errors.New("ocr: ollama API request failed")
	ErrInvalidJSONResponse = errors.New("ocr: model returned invalid JSON")
	ErrContextCanceled     = errors.New("ocr: context canceled or deadline exceeded")
	ErrValidationFailed    = errors.New("ocr: output validation failed")
	ErrEmptySource         = errors.New("ocr: source path or URL is empty")
	ErrURLFetchFailed      = errors.New("ocr: failed to fetch image from URL")
)

// OCRError wraps errors with additional context.
type OCRError struct {
	Op        string // Operation that failed (e.g., "Extract", "LoadImage")
	RequestID string // Request ID for tracing
	Err       error  // Underlying error
}

func (e *OCRError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("ocr [%s] %s: %v", e.RequestID, e.Op, e.Err)
	}
	return fmt.Sprintf("ocr %s: %v", e.Op, e.Err)
}

func (e *OCRError) Unwrap() error {
	return e.Err
}

// NewOCRError creates a new OCRError.
func NewOCRError(op, requestID string, err error) *OCRError {
	return &OCRError{
		Op:        op,
		RequestID: requestID,
		Err:       err,
	}
}

// WrapError wraps an error with operation context.
func WrapError(op string, err error) *OCRError {
	return &OCRError{
		Op:  op,
		Err: err,
	}
}
