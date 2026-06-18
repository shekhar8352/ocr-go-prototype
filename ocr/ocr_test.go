package ocr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

const validOllamaResponse = `{"metadata":{"document_type":"unknown","confidence_score":0.9},"text":{"raw":"hello","lines":[{"text":"hello","confidence":0.9}]},"structured_data":{"key_value_pairs":{},"tables":[]},"summary":null}`

func mockOllamaServer(t *testing.T, response string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/generate":
			if statusCode != http.StatusOK {
				w.WriteHeader(statusCode)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"model":             "test-model",
				"response":          response,
				"done":              true,
				"prompt_eval_count": 10,
				"eval_count":        20,
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestExtract_EmptySource(t *testing.T) {
	_, err := Extract(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty source")
	}
	if !errors.Is(err, ErrEmptySource) {
		t.Errorf("expected ErrEmptySource, got %v", err)
	}
}

func TestExtract_FileNotFound(t *testing.T) {
	_, err := Extract(context.Background(), "/nonexistent/file.png",
		WithOllamaURL("http://localhost:1"),
	)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.bmp")
	if err := os.WriteFile(path, []byte("bmp"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Extract(context.Background(), path, WithOllamaURL("http://localhost:1"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestExtract_FileTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large.png")
	if err := os.WriteFile(path, make([]byte, 2048), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Extract(context.Background(), path,
		WithMaxFileSize(1024),
		WithOllamaURL("http://localhost:1"),
	)
	if err == nil {
		t.Fatal("expected error for file too large")
	}
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestExtract_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	pngData, err := os.ReadFile(filepath.Join("utils", "testdata", "minimal.png"))
	if err != nil {
		// fallback inline 1x1 PNG
		pngData = []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
			0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
			0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, 0x33, 0x00, 0x00, 0x00,
			0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
		}
	}
	if err := os.WriteFile(path, pngData, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	server := mockOllamaServer(t, validOllamaResponse, http.StatusOK)
	defer server.Close()

	result, err := Extract(context.Background(), path,
		WithOllamaURL(server.URL),
		WithModel("test-model"),
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if result.Text.Raw != "hello" {
		t.Errorf("Text.Raw = %q, want %q", result.Text.Raw, "hello")
	}
	if result.Source.Type != models.SourceTypeFile {
		t.Errorf("Source.Type = %q, want file", result.Source.Type)
	}
	if result.Source.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestExtract_InvalidJSONResponse(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(path, minimalPNG(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	server := mockOllamaServer(t, "not json", http.StatusOK)
	defer server.Close()

	_, err := Extract(context.Background(), path,
		WithOllamaURL(server.URL),
		WithModel("test-model"),
		WithMaxRetries(0),
		WithTimeout(10*time.Second),
	)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !errors.Is(err, ErrInvalidJSONResponse) {
		t.Errorf("expected ErrInvalidJSONResponse, got %v", err)
	}
}

func TestExtract_OllamaUnavailable(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(path, minimalPNG(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Extract(context.Background(), path,
		WithOllamaURL("http://127.0.0.1:1"),
		WithTimeout(2*time.Second),
	)
	if err == nil {
		t.Fatal("expected error for unavailable Ollama")
	}
	if !errors.Is(err, ErrOllamaUnavailable) {
		t.Errorf("expected ErrOllamaUnavailable, got %v", err)
	}
}

func TestExtract_StrictValidation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(path, minimalPNG(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// confidence_score out of range will fail validation
	badResponse := `{"metadata":{"document_type":"unknown","confidence_score":2.0},"text":{"raw":"hello","lines":[{"text":"hello","confidence":0.9}]},"structured_data":{"key_value_pairs":{},"tables":[]},"summary":null}`
	server := mockOllamaServer(t, badResponse, http.StatusOK)
	defer server.Close()

	_, err := Extract(context.Background(), path,
		WithOllamaURL(server.URL),
		WithModel("test-model"),
		WithStrictValidation(true),
		WithTimeout(10*time.Second),
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed, got %v", err)
	}
}

func TestExtract_ImageTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(path, minimalPNG(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := Extract(context.Background(), path,
		WithMaxImageDimension(0),
		WithOllamaURL("http://localhost:1"),
	)
	// MaxImageDimension(0) disables the check.
	if err != nil && errors.Is(err, ErrImageDecodeFailed) {
		t.Errorf("unexpected dimension error with disabled check: %v", err)
	}
}

func minimalPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, 0x33, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
