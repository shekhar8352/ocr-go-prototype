// Package utils provides utility functions for the OCR package.
package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// SHA256File computes the SHA-256 checksum of a file.
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("sha256: open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("sha256: read file: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// SHA256Bytes computes the SHA-256 checksum of a byte slice.
func SHA256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
