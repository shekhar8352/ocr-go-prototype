package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaClient_Generate(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Model == "" {
			t.Error("model is empty")
		}
		if req.Stream {
			t.Error("stream should be false")
		}

		resp := GenerateResponse{
			Model:     req.Model,
			Response:  `{"metadata":{"document_type":"unknown","confidence_score":0.5},"text":{"raw":"test","lines":[{"text":"test","confidence":0.5}]},"structured_data":{"key_value_pairs":{},"tables":[]},"summary":null}`,
			Done:      true,
			EvalCount: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 10*time.Second)

	req := GenerateRequest{
		Model:  "test-model",
		Prompt: "Extract text",
		Stream: false,
	}

	resp, err := client.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if resp.Model != "test-model" {
		t.Errorf("Model = %q, want %q", resp.Model, "test-model")
	}
	if resp.Response == "" {
		t.Error("Response is empty")
	}
	if resp.EvalCount != 100 {
		t.Errorf("EvalCount = %d, want %d", resp.EvalCount, 100)
	}
}

func TestOllamaClient_Generate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 10*time.Second)

	_, err := client.Generate(context.Background(), GenerateRequest{
		Model:  "test",
		Prompt: "test",
		Stream: false,
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestOllamaClient_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 10*time.Second)

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOllamaClient_Ping_Unavailable(t *testing.T) {
	client := NewOllamaClient("http://localhost:99999", 2*time.Second)

	err := client.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for unavailable server")
	}
}

func TestOllamaClient_Generate_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Generate(ctx, GenerateRequest{
		Model:  "test",
		Prompt: "test",
		Stream: false,
	})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
