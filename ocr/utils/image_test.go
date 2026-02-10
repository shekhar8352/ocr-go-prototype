package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilePath_Success(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(path, []byte("fake png data"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err := ValidateFilePath(path, 1024*1024)
	if err != nil {
		t.Errorf("ValidateFilePath: unexpected error: %v", err)
	}
}

func TestValidateFilePath_UnsupportedExtension(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.bmp")
	if err := os.WriteFile(path, []byte("fake bmp data"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err := ValidateFilePath(path, 1024*1024)
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
}

func TestValidateFilePath_FileNotFound(t *testing.T) {
	err := ValidateFilePath("/nonexistent/file.png", 1024*1024)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestValidateFilePath_FileTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large.png")
	data := make([]byte, 1024)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	err := ValidateFilePath(path, 512) // Max 512 bytes
	if err == nil {
		t.Fatal("expected error for file too large")
	}
}

func TestValidateFilePath_Directory(t *testing.T) {
	// Create a directory with a supported extension-like name
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "test.png")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	err := ValidateFilePath(dirPath, 1024*1024)
	if err == nil {
		t.Fatal("expected error for directory")
	}
}

func TestValidateURL_Valid(t *testing.T) {
	tests := []string{
		"https://example.com/image.png",
		"http://example.com/doc.jpg",
		"https://cdn.example.org/files/scan.pdf",
	}

	for _, u := range tests {
		if err := ValidateURL(u); err != nil {
			t.Errorf("ValidateURL(%q): unexpected error: %v", u, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	tests := []struct {
		url  string
		desc string
	}{
		{"ftp://example.com/file.png", "unsupported scheme"},
		{"http://localhost/image.png", "localhost blocked"},
		{"http://127.0.0.1/image.png", "loopback blocked"},
		{"http://192.168.1.1/image.png", "private IP blocked"},
		{"http://10.0.0.1/image.png", "private IP blocked"},
		{"://invalid", "invalid URL"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q): expected error for %s", tt.url, tt.desc)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/image.png", true},
		{"http://example.com/image.png", true},
		{"/local/path/image.png", false},
		{"./relative/path.jpg", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsURL(tt.input); got != tt.expected {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/image.PNG", ".png"},
		{"/path/to/doc.pdf", ".pdf"},
		{"https://example.com/file.jpg", ".jpg"},
		{"/no-extension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := FileExtension(tt.input); got != tt.expected {
				t.Errorf("FileExtension(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLoadImageFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	content := []byte("fake image content")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	data, err := LoadImageFromFile(path)
	if err != nil {
		t.Fatalf("LoadImageFromFile: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("LoadImageFromFile: got %q, want %q", data, content)
	}
}

func TestLoadImageFromFile_NotFound(t *testing.T) {
	_, err := LoadImageFromFile("/nonexistent/file.png")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestEncodeBase64(t *testing.T) {
	data := []byte("hello")
	encoded := EncodeBase64(data)
	if encoded == "" {
		t.Fatal("EncodeBase64 returned empty string")
	}
	// Base64 of "hello" = "aGVsbG8="
	expected := "aGVsbG8="
	if encoded != expected {
		t.Errorf("EncodeBase64 = %q, want %q", encoded, expected)
	}
}

func TestGetImageInfo_UnknownFormat(t *testing.T) {
	info := GetImageInfo([]byte("not a real image"), ".png")
	if info.ColorMode != "Unknown" {
		t.Errorf("expected Unknown color mode, got %q", info.ColorMode)
	}
}

func TestGetImageInfo_PDF(t *testing.T) {
	info := GetImageInfo([]byte("pdf data"), ".pdf")
	if info.Width != 0 || info.Height != 0 {
		t.Errorf("expected 0x0 for PDF, got %dx%d", info.Width, info.Height)
	}
	if info.ColorMode != "Unknown" {
		t.Errorf("expected Unknown color mode for PDF, got %q", info.ColorMode)
	}
}
