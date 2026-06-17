package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		return fmt.Errorf("%w: %q", ErrUnsupportedExtension, ext)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		return fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("%w: %s", ErrPathIsDirectory, path)
	}

	if info.Size() > maxSize {
		return fmt.Errorf("%w: file size %d exceeds maximum %d bytes", ErrFileTooLarge, info.Size(), maxSize)
	}

	return nil
}

// ValidateURL checks that a URL is well-formed, uses http/https, and does not target private hosts.
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: %s (only http and https are allowed)", ErrInvalidURLScheme, u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("URL has no host")
	}

	return validateHost(u.Hostname())
}

// validateHost blocks private, loopback, and link-local addresses.
func validateHost(host string) error {
	host = strings.ToLower(strings.Trim(host, "[]"))

	blockedHosts := []string{"localhost", "0.0.0.0"}
	for _, blocked := range blockedHosts {
		if host == blocked {
			return fmt.Errorf("%w: %s", ErrUnsafeURL, host)
		}
	}

	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: %s", ErrUnsafeURL, host)
		}
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		// Allow DNS failures here; the download step will fail naturally.
		return nil
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: %s resolves to blocked IP %s", ErrUnsafeURL, host, ip.String())
		}
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}

	// Block IPv6 unique local addresses (fc00::/7).
	if ip.To16() != nil && ip.To4() == nil {
		return ip[0]&0xfe == 0xfc
	}

	return false
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
func DownloadImage(ctx context.Context, rawURL string, maxSize int64) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if err := validateHost(req.URL.Hostname()); err != nil {
				return err
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download image: HTTP %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("download image: read body: %w", err)
	}

	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("%w: downloaded file exceeds maximum size of %d bytes", ErrFileTooLarge, maxSize)
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

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
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

// ValidateImageDimensions checks that image width and height do not exceed maxDimension.
func ValidateImageDimensions(info models.ImageInfo, maxDimension int) error {
	if maxDimension <= 0 {
		return nil
	}
	if info.Width > maxDimension || info.Height > maxDimension {
		return fmt.Errorf("%w: %dx%d exceeds maximum %d pixels per side", ErrImageTooLarge, info.Width, info.Height, maxDimension)
	}
	return nil
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
