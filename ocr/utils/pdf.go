package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// PDFOptions configures PDF-to-image conversion.
type PDFOptions struct {
	MaxPages int
}

// PDFToImages converts a PDF to a slice of PNG image byte slices, one per page.
// Requires pdftoppm (poppler-utils) to be installed.
func PDFToImages(pdfPath string, opts PDFOptions) ([][]byte, error) {
	pages, err := pdfToPPM(pdfPath)
	if err != nil {
		if _, lookErr := exec.LookPath("pdftoppm"); lookErr != nil {
			return nil, fmt.Errorf("%w: install poppler-utils (e.g. brew install poppler)", ErrPDFToolUnavailable)
		}
		return nil, fmt.Errorf("%w: %v", ErrPDFParseFailed, err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("%w: PDF produced no pages", ErrPDFParseFailed)
	}

	if opts.MaxPages > 0 && len(pages) > opts.MaxPages {
		return nil, fmt.Errorf("%w: %d pages exceeds maximum %d", ErrPDFTooManyPages, len(pages), opts.MaxPages)
	}

	return pages, nil
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

// PDFToolAvailable reports whether pdftoppm is available on the system.
func PDFToolAvailable() bool {
	_, err := exec.LookPath("pdftoppm")
	return err == nil
}
