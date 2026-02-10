package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSHA256File(t *testing.T) {
	// Create temp file with known content
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	hash, err := SHA256File(path)
	if err != nil {
		t.Fatalf("SHA256File: %v", err)
	}

	// SHA-256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("SHA256File = %q, want %q", hash, expected)
	}
}

func TestSHA256Bytes(t *testing.T) {
	hash := SHA256Bytes([]byte("hello world"))
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("SHA256Bytes = %q, want %q", hash, expected)
	}
}

func TestSHA256File_NotFound(t *testing.T) {
	_, err := SHA256File("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
