package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStorageProvider struct {
	staticDir string
	baseURL   string
}

func NewLocalStorageProvider(staticDir, baseURL string) *LocalStorageProvider {
	return &LocalStorageProvider{
		staticDir: staticDir,
		baseURL:   baseURL,
	}
}

func (p *LocalStorageProvider) UploadFile(ctx context.Context, filename string, file io.Reader, size int64, contentType string) (string, error) {
	if err := os.MkdirAll(p.staticDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	ext := filepath.Ext(filename)
	cleanName := strings.TrimSuffix(filename, ext)
	uniqueName := fmt.Sprintf("%s-%s%s", cleanName, uuid.New().String(), ext)
	filePath := filepath.Join(p.staticDir, uniqueName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	fileURL := fmt.Sprintf("%s/static/uploads/%s", strings.TrimSuffix(p.baseURL, "/"), uniqueName)
	return fileURL, nil
}

func (p *LocalStorageProvider) DeleteFile(ctx context.Context, fileURL string) error {
	parts := strings.Split(fileURL, "/static/uploads/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid local file URL")
	}
	filename := parts[1]
	filePath := filepath.Join(p.staticDir, filename)
	return os.Remove(filePath)
}
