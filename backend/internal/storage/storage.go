package storage

import (
	"context"
	"io"
)

type StorageProvider interface {
	UploadFile(ctx context.Context, filename string, file io.Reader, size int64, contentType string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}
