// Package engine provides the OCR orchestration logic.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/client"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/prompt"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/utils"
)

// VisionEngine orchestrates the OCR pipeline:
// load image → build prompt → call Ollama → parse/validate → return result
type VisionEngine struct {
	client *client.OllamaClient
	logger *slog.Logger
}

// NewVisionEngine creates a new VisionEngine.
func NewVisionEngine(ollamaClient *client.OllamaClient, logger *slog.Logger) *VisionEngine {
	return &VisionEngine{
		client: ollamaClient,
		logger: logger,
	}
}

// ProcessConfig holds per-request processing parameters.
type ProcessConfig struct {
	Model       string
	Temperature float64
	RequestID   string

	WithSummary              bool
	WithLanguageDetection    bool
	WithStructuredExtraction bool
	WithBoundingBoxes        bool
	WithConfidenceScores     bool
}

// ProcessResult holds the engine output.
type ProcessResult struct {
	VisionResponse *models.OllamaVisionResponse
	Model          string
	PromptTokens   int
	EvalTokens     int
	Latency        time.Duration
}

// Process runs OCR on a single image (as bytes) using the Ollama vision model.
func (e *VisionEngine) Process(ctx context.Context, imageData []byte, cfg ProcessConfig) (*ProcessResult, error) {
	startTime := time.Now()

	e.logger.Info("starting OCR processing",
		slog.String("request_id", cfg.RequestID),
		slog.String("model", cfg.Model),
		slog.Int("image_bytes", len(imageData)),
	)

	// Build prompt
	promptCfg := prompt.PromptConfig{
		WithSummary:              cfg.WithSummary,
		WithLanguageDetection:    cfg.WithLanguageDetection,
		WithStructuredExtraction: cfg.WithStructuredExtraction,
		WithBoundingBoxes:        cfg.WithBoundingBoxes,
		WithConfidenceScores:     cfg.WithConfidenceScores,
	}
	ocrPrompt := prompt.BuildOCRPrompt(promptCfg)

	// Encode image
	base64Image := utils.EncodeBase64(imageData)

	// Build Ollama request
	req := client.GenerateRequest{
		Model:  cfg.Model,
		Prompt: ocrPrompt,
		Images: []string{base64Image},
		Stream: false,
		Format: "json",
		Options: &client.ModelOptions{
			Temperature: cfg.Temperature,
			NumPredict:  4096,
		},
	}

	// Call Ollama — attempt + 1 retry on JSON parse failure
	var lastErr error
	for attempt := 0; attempt <= 1; attempt++ {
		if attempt > 0 {
			e.logger.Warn("retrying OCR request due to JSON parse failure",
				slog.String("request_id", cfg.RequestID),
				slog.Int("attempt", attempt),
			)
		}

		resp, err := e.client.Generate(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("ollama generate (attempt %d): %w", attempt, err)
		}

		e.logger.Info("ollama response received",
			slog.String("request_id", cfg.RequestID),
			slog.Int("prompt_eval_count", resp.PromptEvalCount),
			slog.Int("eval_count", resp.EvalCount),
			slog.Int64("total_duration_ns", resp.TotalDuration),
			slog.Int("response_length", len(resp.Response)),
		)

		// Parse JSON
		visionResp, err := utils.ParseAndValidateJSON(resp.Response)
		if err != nil {
			lastErr = fmt.Errorf("parse response (attempt %d): %w", attempt, err)
			e.logger.Warn("JSON parse failed",
				slog.String("request_id", cfg.RequestID),
				slog.String("error", err.Error()),
				slog.String("raw_response_preview", truncate(resp.Response, 500)),
			)
			continue
		}

		latency := time.Since(startTime)
		e.logger.Info("OCR processing complete",
			slog.String("request_id", cfg.RequestID),
			slog.Duration("latency", latency),
		)

		return &ProcessResult{
			VisionResponse: visionResp,
			Model:          resp.Model,
			PromptTokens:   resp.PromptEvalCount,
			EvalTokens:     resp.EvalCount,
			Latency:        latency,
		}, nil
	}

	return nil, fmt.Errorf("all attempts failed: %w", lastErr)
}

// ProcessPDF handles multi-page PDF processing by converting pages to images
// and processing each page, then merging results.
func (e *VisionEngine) ProcessPDF(ctx context.Context, pdfPath string, cfg ProcessConfig) (*ProcessResult, error) {
	e.logger.Info("processing PDF",
		slog.String("request_id", cfg.RequestID),
		slog.String("path", pdfPath),
	)

	pages, err := utils.PDFToImages(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("convert PDF to images: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("PDF produced no pages")
	}

	// If single page, process directly
	if len(pages) == 1 {
		return e.Process(ctx, pages[0], cfg)
	}

	// Multi-page: process each and merge
	var allResults []*ProcessResult
	for i, page := range pages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		e.logger.Info("processing PDF page",
			slog.String("request_id", cfg.RequestID),
			slog.Int("page", i+1),
			slog.Int("total_pages", len(pages)),
		)

		result, err := e.Process(ctx, page, cfg)
		if err != nil {
			return nil, fmt.Errorf("process page %d: %w", i+1, err)
		}
		allResults = append(allResults, result)
	}

	// Merge results
	return mergeResults(allResults), nil
}

// mergeResults combines multiple page results into a single result.
func mergeResults(results []*ProcessResult) *ProcessResult {
	if len(results) == 0 {
		return nil
	}
	if len(results) == 1 {
		return results[0]
	}

	merged := &ProcessResult{
		VisionResponse: &models.OllamaVisionResponse{
			Metadata: results[0].VisionResponse.Metadata,
			Text: &models.OllamaTextResult{
				Raw:   "",
				Lines: nil,
			},
			StructuredData: &models.OllamaStructuredData{
				KeyValuePairs: make(map[string]string),
				Tables:        nil,
			},
			Summary: nil,
		},
		Model: results[0].Model,
	}

	var rawParts []string
	var totalLatency time.Duration

	for i, r := range results {
		totalLatency += r.Latency
		merged.PromptTokens += r.PromptTokens
		merged.EvalTokens += r.EvalTokens

		if r.VisionResponse.Text != nil {
			pagePrefix := fmt.Sprintf("--- Page %d ---\n", i+1)
			rawParts = append(rawParts, pagePrefix+r.VisionResponse.Text.Raw)
			merged.VisionResponse.Text.Lines = append(merged.VisionResponse.Text.Lines, r.VisionResponse.Text.Lines...)
		}

		if r.VisionResponse.StructuredData != nil {
			for k, v := range r.VisionResponse.StructuredData.KeyValuePairs {
				merged.VisionResponse.StructuredData.KeyValuePairs[k] = v
			}
			merged.VisionResponse.StructuredData.Tables = append(
				merged.VisionResponse.StructuredData.Tables,
				r.VisionResponse.StructuredData.Tables...,
			)
		}

		// Use the summary from the last page if available
		if r.VisionResponse.Summary != nil {
			merged.VisionResponse.Summary = r.VisionResponse.Summary
		}
	}

	merged.VisionResponse.Text.Raw = strings.Join(rawParts, "\n")
	merged.Latency = totalLatency

	return merged
}

// IsPDF checks if a file extension indicates a PDF.
func IsPDF(source string) bool {
	ext := strings.ToLower(filepath.Ext(source))
	return ext == ".pdf"
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
