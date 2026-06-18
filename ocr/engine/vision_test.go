package engine

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/client"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

const validResponse = `{"metadata":{"document_type":"invoice","confidence_score":0.9},"text":{"raw":"page text","lines":[{"text":"line1","confidence":0.9}]},"structured_data":{"key_value_pairs":{"total":"100"},"tables":[]},"summary":null}`

func newTestEngine(t *testing.T, handler http.HandlerFunc) *VisionEngine {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	ollamaClient := client.NewOllamaClient(server.URL, 10*time.Second)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewVisionEngine(ollamaClient, logger)
}

func defaultHandler(attempts *int, response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		*attempts++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"model":    "test-model",
			"response": response,
			"done":     true,
		})
	}
}

func TestVisionEngine_Process_Success(t *testing.T) {
	attempts := 0
	eng := newTestEngine(t, defaultHandler(&attempts, validResponse))

	result, err := eng.Process(context.Background(), []byte("image"), ProcessConfig{
		Model:       "test-model",
		Temperature: 0.1,
		RequestID:   "req-1",
		MaxRetries:  1,
	})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if result.VisionResponse.Text.Raw != "page text" {
		t.Errorf("raw text = %q, want page text", result.VisionResponse.Text.Raw)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestVisionEngine_Process_RetryOnInvalidJSON(t *testing.T) {
	attempts := 0
	eng := newTestEngine(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		resp := "not json"
		if attempts > 1 {
			resp = validResponse
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"model":    "test-model",
			"response": resp,
			"done":     true,
		})
	})

	result, err := eng.Process(context.Background(), []byte("image"), ProcessConfig{
		Model:      "test-model",
		RequestID:  "req-1",
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatalf("Process after retry: %v", err)
	}
	if result.VisionResponse.Text.Raw != "page text" {
		t.Errorf("unexpected text: %q", result.VisionResponse.Text.Raw)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestVisionEngine_Process_InvalidJSONExhausted(t *testing.T) {
	attempts := 0
	eng := newTestEngine(t, defaultHandler(&attempts, "bad json"))

	_, err := eng.Process(context.Background(), []byte("image"), ProcessConfig{
		Model:      "test-model",
		RequestID:  "req-1",
		MaxRetries: 0,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidJSONResponse) {
		t.Errorf("expected ErrInvalidJSONResponse, got %v", err)
	}
}

func TestVisionEngine_Process_ContextCanceled(t *testing.T) {
	eng := newTestEngine(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := eng.Process(ctx, []byte("image"), ProcessConfig{
		Model:     "test-model",
		RequestID: "req-1",
	})
	if err == nil {
		t.Fatal("expected context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestMergeResults_MultiPage(t *testing.T) {
	results := []*ProcessResult{
		{
			VisionResponse: &models.OllamaVisionResponse{
				Text: &models.OllamaTextResult{
					Raw: "page one",
					Lines: []models.OllamaTextLine{
						{Text: "line1", Confidence: 0.9},
					},
				},
				StructuredData: &models.OllamaStructuredData{
					KeyValuePairs: map[string]string{"total": "100"},
					Tables: []models.Table{
						{Headers: []string{"A"}, Rows: [][]string{{"1"}}},
					},
				},
			},
			Model: "test-model",
		},
		{
			VisionResponse: &models.OllamaVisionResponse{
				Text: &models.OllamaTextResult{
					Raw: "page two",
					Lines: []models.OllamaTextLine{
						{Text: "line2", Confidence: 0.8},
					},
				},
				StructuredData: &models.OllamaStructuredData{
					KeyValuePairs: map[string]string{"total": "200"},
				},
				Summary: strPtr("summary"),
			},
			Model:        "test-model",
			PromptTokens: 5,
			EvalTokens:   10,
			Latency:      time.Second,
		},
	}

	merged := mergeResults(results)
	if merged == nil {
		t.Fatal("merged result is nil")
	}

	if merged.VisionResponse.Text.Lines[0].Text != "[Page 1] line1" {
		t.Errorf("line 1 prefix missing: %q", merged.VisionResponse.Text.Lines[0].Text)
	}
	if merged.VisionResponse.Text.Lines[1].Text != "[Page 2] line2" {
		t.Errorf("line 2 prefix missing: %q", merged.VisionResponse.Text.Lines[1].Text)
	}
	if merged.VisionResponse.StructuredData.KeyValuePairs["page_1_total"] != "100" {
		t.Errorf("page 1 kv missing")
	}
	if merged.VisionResponse.StructuredData.KeyValuePairs["page_2_total"] != "200" {
		t.Errorf("page 2 kv missing")
	}
	if merged.PromptTokens != 5 {
		t.Errorf("PromptTokens = %d, want 5", merged.PromptTokens)
	}
}

func strPtr(s string) *string {
	return &s
}
