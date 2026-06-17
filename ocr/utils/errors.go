package utils

import "errors"

// Sentinel errors for input validation and processing.
var (
	ErrUnsupportedExtension = errors.New("unsupported file extension")
	ErrFileNotFound         = errors.New("file not found")
	ErrFileTooLarge         = errors.New("file exceeds maximum allowed size")
	ErrPathIsDirectory      = errors.New("path is a directory, not a file")
	ErrImageTooLarge        = errors.New("image dimensions exceed maximum allowed size")
	ErrImageDecodeFailed    = errors.New("failed to decode image")
	ErrPDFToolUnavailable   = errors.New("pdftoppm is not available; install poppler-utils for PDF support")
	ErrPDFParseFailed       = errors.New("failed to convert PDF to images")
	ErrPDFTooManyPages      = errors.New("PDF exceeds maximum allowed page count")
	ErrUnsafeURL            = errors.New("URL points to a blocked or private host")
	ErrInvalidURLScheme     = errors.New("unsupported URL scheme")
)
