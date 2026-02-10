package utils

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

// SupportedExtensions lists the file extensions this package supports.
var SupportedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".pdf":  true,
}

// ValidateFilePath checks that a file exists, is within size limits, and has a supported extension.
func ValidateFilePath(path string, maxSize int64) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !SupportedExtensions[ext] {
		return fmt.Errorf("unsupported extension %q", ext)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	if info.Size() > maxSize {
		return fmt.Errorf("file size %d exceeds maximum %d bytes", info.Size(), maxSize)
	}

	return nil
}

// ValidateURL checks that a URL is well-formed and uses http/https.
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s (only http and https are allowed)", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("URL has no host")
	}

	// Block private/internal IPs for SSRF protection
	host := strings.ToLower(u.Hostname())
	blockedPrefixes := []string{"127.", "10.", "192.168.", "172.16.", "172.17.", "172.18.",
		"172.19.", "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31."}
	blockedHosts := []string{"localhost", "0.0.0.0", "[::1]"}

	for _, blocked := range blockedHosts {
		if host == blocked {
			return fmt.Errorf("URL points to a blocked host: %s", host)
		}
	}
	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(host, prefix) {
			return fmt.Errorf("URL points to a private network: %s", host)
		}
	}

	return nil
}

// IsURL returns true if the source looks like a URL.
func IsURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// LoadImageFromFile reads an image file and returns its bytes.
func LoadImageFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// DownloadImage fetches an image from a URL and returns its bytes.
func DownloadImage(rawURL string, maxSize int64) ([]byte, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download image: HTTP %d", resp.StatusCode)
	}

	// Limit reader to prevent downloading excessively large files
	limited := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("download image: read body: %w", err)
	}

	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("downloaded file exceeds maximum size of %d bytes", maxSize)
	}

	return data, nil
}

// EncodeBase64 encodes bytes to a base64 string.
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GetImageInfo decodes image dimensions and color mode from raw bytes.
// For PDFs it returns a placeholder since we handle them page-by-page.
func GetImageInfo(data []byte, ext string) models.ImageInfo {
	if strings.ToLower(ext) == ".pdf" {
		return models.ImageInfo{
			Width:     0,
			Height:    0,
			DPI:       nil,
			ColorMode: models.ColorModeUnknown,
		}
	}

	cfg, _, err := image.DecodeConfig(strings.NewReader(string(data)))
	if err != nil {
		return models.ImageInfo{
			Width:     0,
			Height:    0,
			DPI:       nil,
			ColorMode: models.ColorModeUnknown,
		}
	}

	colorMode := models.ColorModeUnknown
	if cfg.ColorModel != nil {
		switch cfg.ColorModel {
		case color.YCbCrModel:
			colorMode = models.ColorModeRGB
		default:
			// Try to detect via model string representation
			modelStr := fmt.Sprintf("%T", cfg.ColorModel)
			switch {
			case strings.Contains(modelStr, "RGBA") || strings.Contains(modelStr, "NRGBA"):
				colorMode = models.ColorModeRGB
			case strings.Contains(modelStr, "Gray"):
				colorMode = models.ColorModeGrayscale
			case strings.Contains(modelStr, "CMYK"):
				colorMode = models.ColorModeCMYK
			}
		}
	}

	return models.ImageInfo{
		Width:     cfg.Width,
		Height:    cfg.Height,
		DPI:       nil, // DPI not easily extractable from Go's image package
		ColorMode: colorMode,
	}
}

// FileExtension returns the lowercase extension for a source path or URL.
func FileExtension(source string) string {
	if IsURL(source) {
		u, _ := url.Parse(source)
		if u != nil {
			return strings.ToLower(filepath.Ext(u.Path))
		}
		return ""
	}
	return strings.ToLower(filepath.Ext(source))
}
