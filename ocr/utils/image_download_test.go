package utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sudhanshushekhar/ocr-go-prototype/ocr/models"
)

func TestGetImageInfo_RealPNG(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "minimal.png"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	info := GetImageInfo(data, ".png")
	if info.Width != 1 || info.Height != 1 {
		t.Errorf("expected 1x1 image, got %dx%d", info.Width, info.Height)
	}
}

func TestValidateImageDimensions(t *testing.T) {
	info := ImageInfoFromDimensions(100, 200)
	err := ValidateImageDimensions(info, 150)
	if err == nil {
		t.Fatal("expected dimension error")
	}
	if !errors.Is(err, ErrImageTooLarge) {
		t.Errorf("expected ErrImageTooLarge, got %v", err)
	}

	if err := ValidateImageDimensions(info, 200); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFilePath_ErrorsAreTyped(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.bmp")
	if err := os.WriteFile(path, []byte("bmp"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := ValidateFilePath(path, 1024)
	if !errors.Is(err, ErrUnsupportedExtension) {
		t.Errorf("expected ErrUnsupportedExtension, got %v", err)
	}
}

func TestDownloadImage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image-data"))
	}))
	defer server.Close()

	data, err := DownloadImage(context.Background(), server.URL, 1024)
	if err != nil {
		t.Fatalf("DownloadImage: %v", err)
	}
	if string(data) != "image-data" {
		t.Errorf("got %q, want image-data", data)
	}
}

func TestDownloadImage_TooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 2048))
	}))
	defer server.Close()

	_, err := DownloadImage(context.Background(), server.URL, 1024)
	if err == nil {
		t.Fatal("expected size error")
	}
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestDownloadImage_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DownloadImage(ctx, server.URL, 1024)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestDownloadImage_BlockedRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1/secret", http.StatusFound)
	}))
	defer server.Close()

	_, err := DownloadImage(context.Background(), server.URL, 1024)
	if err == nil {
		t.Fatal("expected redirect block error")
	}
}

// ImageInfoFromDimensions is a test helper.
func ImageInfoFromDimensions(width, height int) models.ImageInfo {
	return models.ImageInfo{Width: width, Height: height, ColorMode: models.ColorModeRGB}
}