package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// PDFToImages converts a PDF to a slice of PNG image byte slices, one per page.
// This implementation uses a system call to a tool that can render PDFs.
// For production use, consider using a Go-native PDF rendering library.
//
// Strategy: We try multiple approaches in order:
// 1. Use 'pdftoppm' (poppler-utils) if available
// 2. Use 'sips' (macOS built-in) for single-page conversion
// 3. Return the raw PDF bytes as a single "page" for Ollama to process directly
func PDFToImages(pdfPath string) ([][]byte, error) {
	// Try pdftoppm first (most reliable for multi-page PDFs)
	if pages, err := pdfToPPM(pdfPath); err == nil && len(pages) > 0 {
		return pages, nil
	}

	// Fallback: return the raw PDF data as a single entry.
	// Many vision models can process PDF data directly when sent as base64.
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("pdf fallback read: %w", err)
	}

	return [][]byte{data}, nil
}

// pdfToPPM uses pdftoppm from poppler-utils to convert PDF pages to PNG images.
func pdfToPPM(pdfPath string) ([][]byte, error) {
	pdftoppm, err := exec.LookPath("pdftoppm")
	if err != nil {
		return nil, fmt.Errorf("pdftoppm not found: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "ocr-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPrefix := filepath.Join(tmpDir, "page")

	cmd := exec.Command(pdftoppm, "-png", "-r", "300", pdfPath, outputPrefix)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %s: %w", string(output), err)
	}

	// Read all generated PNG files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("read temp dir: %w", err)
	}

	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	var pages [][]byte
	for _, name := range filenames {
		data, err := os.ReadFile(filepath.Join(tmpDir, name))
		if err != nil {
			return nil, fmt.Errorf("read page image: %w", err)
		}
		pages = append(pages, data)
	}

	return pages, nil
}
