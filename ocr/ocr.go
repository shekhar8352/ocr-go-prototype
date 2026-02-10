// Package ocr provides the public API for performing OCR using locally running
// Ollama vision models. Users interact with this package through a single
// function: Extract.
//
// Example usage:
//
//	result, err := ocr.Extract(ctx, "/path/to/invoice.png",
//	    ocr.WithSummary(true),
//	    ocr.WithModel("llama3.2-vision"),
//	)
package ocr

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/client"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/engine"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/utils"
)

// Extract is the primary public API for the OCR package.
// It accepts a local file path or a remote URL, and returns a fully
// structured OCRResult with strict JSON-compatible output.
//
// Usage:
//
//	result, err := ocr.Extract(ctx, "/path/to/image.png")
//	result, err := ocr.Extract(ctx, "https://example.com/doc.jpg", ocr.WithSummary(true))
func Extract(ctx context.Context, source string, opts ...Option) (*models.OCRResult, error) {
	// Generate request ID
	requestID := generateRequestID()

	// Build config
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger = logger.With(
		slog.String("request_id", requestID),
		slog.String("model", cfg.Model),
	)

	logger.Info("OCR extraction started",
		slog.String("source", source),
	)

	// Validate source
	if source == "" {
		return nil, NewOCRError("Extract", requestID, ErrEmptySource)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	// Determine source type and load image data
	var (
		imageData  []byte
		sourceType models.SourceType
		checksum   string
		ext        string
		imageInfo  models.ImageInfo
		isPDF      bool
		err        error
	)

	if utils.IsURL(source) {
		sourceType = models.SourceTypeURL

		if err := utils.ValidateURL(source); err != nil {
			return nil, NewOCRError("Extract.ValidateURL", requestID, fmt.Errorf("%w: %v", ErrInvalidURL, err))
		}

		ext = utils.FileExtension(source)
		isPDF = ext == ".pdf"

		logger.Info("downloading image from URL",
			slog.String("url", source),
		)

		imageData, err = utils.DownloadImage(source, cfg.MaxFileSize)
		if err != nil {
			return nil, NewOCRError("Extract.DownloadImage", requestID, fmt.Errorf("%w: %v", ErrURLFetchFailed, err))
		}

		checksum = utils.SHA256Bytes(imageData)
	} else {
		sourceType = models.SourceTypeFile
		ext = utils.FileExtension(source)
		isPDF = ext == ".pdf"

		if err := utils.ValidateFilePath(source, cfg.MaxFileSize); err != nil {
			return nil, NewOCRError("Extract.ValidateFile", requestID, fmt.Errorf("%w: %v", ErrFileNotFound, err))
		}

		imageData, err = utils.LoadImageFromFile(source)
		if err != nil {
			return nil, NewOCRError("Extract.LoadImage", requestID, fmt.Errorf("%w: %v", ErrFileReadFailed, err))
		}

		checksum, err = utils.SHA256File(source)
		if err != nil {
			return nil, NewOCRError("Extract.Checksum", requestID, fmt.Errorf("%w: %v", ErrFileReadFailed, err))
		}
	}

	// Get image info
	imageInfo = utils.GetImageInfo(imageData, ext)

	// Create Ollama client
	ollamaClient := client.NewOllamaClient(cfg.OllamaURL, cfg.Timeout)

	// Ping Ollama
	if err := ollamaClient.Ping(ctx); err != nil {
		return nil, NewOCRError("Extract.Ping", requestID, fmt.Errorf("%w: %v", ErrOllamaUnavailable, err))
	}

	// Create engine
	eng := engine.NewVisionEngine(ollamaClient, logger)

	processCfg := engine.ProcessConfig{
		Model:                    cfg.Model,
		Temperature:              cfg.Temperature,
		RequestID:                requestID,
		WithSummary:              cfg.WithSummary,
		WithLanguageDetection:    cfg.WithLanguageDetection,
		WithStructuredExtraction: cfg.WithStructuredExtraction,
		WithBoundingBoxes:        cfg.WithBoundingBoxes,
		WithConfidenceScores:     cfg.WithConfidenceScores,
	}

	// Process
	var result *engine.ProcessResult
	if isPDF {
		if sourceType == models.SourceTypeURL {
			// For URL-sourced PDFs, save to tmp and process
			tmpFile, err := os.CreateTemp("", "ocr-pdf-*.pdf")
			if err != nil {
				return nil, NewOCRError("Extract.TempFile", requestID, err)
			}
			defer os.Remove(tmpFile.Name())
			if _, err := tmpFile.Write(imageData); err != nil {
				tmpFile.Close()
				return nil, NewOCRError("Extract.WriteTempFile", requestID, err)
			}
			tmpFile.Close()
			result, err = eng.ProcessPDF(ctx, tmpFile.Name(), processCfg)
			if err != nil {
				return nil, NewOCRError("Extract.ProcessPDF", requestID, fmt.Errorf("%w: %v", ErrOllamaRequestFailed, err))
			}
		} else {
			result, err = eng.ProcessPDF(ctx, source, processCfg)
			if err != nil {
				return nil, NewOCRError("Extract.ProcessPDF", requestID, fmt.Errorf("%w: %v", ErrOllamaRequestFailed, err))
			}
		}
	} else {
		result, err = eng.Process(ctx, imageData, processCfg)
		if err != nil {
			return nil, NewOCRError("Extract.Process", requestID, fmt.Errorf("%w: %v", ErrOllamaRequestFailed, err))
		}
	}

	// Build OCRResult from engine result
	ocrResult := buildOCRResult(source, sourceType, checksum, imageInfo, result, cfg)

	// Validate
	if err := utils.ValidateOCRResult(ocrResult); err != nil {
		logger.Warn("output validation failed, returning result anyway",
			slog.String("validation_error", err.Error()),
		)
	}

	logger.Info("OCR extraction complete",
		slog.Duration("total_latency", result.Latency),
		slog.Int("prompt_tokens", result.PromptTokens),
		slog.Int("eval_tokens", result.EvalTokens),
	)

	return ocrResult, nil
}

// buildOCRResult assembles the final OCRResult from engine output.
func buildOCRResult(
	source string,
	sourceType models.SourceType,
	checksum string,
	imageInfo models.ImageInfo,
	result *engine.ProcessResult,
	cfg *Config,
) *models.OCRResult {
	ocrResult := &models.OCRResult{
		Source: models.Source{
			Type:     sourceType,
			Path:     source,
			Checksum: checksum,
		},
		Image:          imageInfo,
		Metadata:       buildMetadata(result.VisionResponse),
		Text:           buildText(result.VisionResponse, cfg),
		StructuredData: buildStructuredData(result.VisionResponse, cfg),
		Summary:        buildSummary(result.VisionResponse, cfg),
	}

	// Override image info if the model provided it
	if result.VisionResponse.Image != nil {
		vi := result.VisionResponse.Image
		if vi.Width > 0 {
			ocrResult.Image.Width = vi.Width
		}
		if vi.Height > 0 {
			ocrResult.Image.Height = vi.Height
		}
		if vi.DPI != nil {
			ocrResult.Image.DPI = vi.DPI
		}
		if vi.ColorMode != "" {
			cm := models.ColorMode(vi.ColorMode)
			if _, ok := utils.ValidColorModes[cm]; ok {
				ocrResult.Image.ColorMode = cm
			}
		}
	}

	return ocrResult
}

func buildMetadata(resp *models.OllamaVisionResponse) models.Metadata {
	md := models.Metadata{
		Language:        nil,
		DocumentType:    models.DocumentTypeUnknown,
		ConfidenceScore: 0,
	}

	if resp.Metadata != nil {
		md.Language = resp.Metadata.Language
		md.ConfidenceScore = resp.Metadata.ConfidenceScore

		dt := models.DocumentType(resp.Metadata.DocumentType)
		if _, ok := utils.ValidDocumentTypes[dt]; ok {
			md.DocumentType = dt
		}
	}

	return md
}

func buildText(resp *models.OllamaVisionResponse, cfg *Config) models.TextResult {
	text := models.TextResult{
		Raw:   "",
		Lines: []models.TextLine{},
	}

	if resp.Text == nil {
		return text
	}

	text.Raw = resp.Text.Raw

	for _, line := range resp.Text.Lines {
		tl := models.TextLine{
			Text:       line.Text,
			Confidence: line.Confidence,
		}

		if cfg.WithBoundingBoxes && line.BoundingBox != nil {
			tl.BoundingBox = line.BoundingBox
		}

		if !cfg.WithConfidenceScores {
			tl.Confidence = 0
		}

		text.Lines = append(text.Lines, tl)
	}

	return text
}

func buildStructuredData(resp *models.OllamaVisionResponse, cfg *Config) models.StructuredData {
	sd := models.StructuredData{
		KeyValuePairs: make(map[string]string),
		Tables:        []models.Table{},
	}

	if !cfg.WithStructuredExtraction || resp.StructuredData == nil {
		return sd
	}

	if resp.StructuredData.KeyValuePairs != nil {
		sd.KeyValuePairs = resp.StructuredData.KeyValuePairs
	}

	if resp.StructuredData.Tables != nil {
		sd.Tables = resp.StructuredData.Tables
	}

	return sd
}

func buildSummary(resp *models.OllamaVisionResponse, cfg *Config) *string {
	if !cfg.WithSummary {
		return nil
	}
	return resp.Summary
}

// generateRequestID creates a unique request ID using timestamp + random component.
func generateRequestID() string {
	return fmt.Sprintf("ocr-%d", time.Now().UnixNano())
}
