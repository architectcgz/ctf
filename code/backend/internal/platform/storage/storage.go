package storage

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"
)

var ErrNotFound = errors.New("shared storage object not found")
var ErrUnsafeKey = errors.New("shared storage unsafe key")

type ObjectInfo struct {
	Key         string
	Size        int64
	ModTime     time.Time
	ContentType string
}

type Store interface {
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, reader io.Reader) error
	Stat(ctx context.Context, key string) (*ObjectInfo, error)
}

type LocalWritableStore interface {
	Store
	PrepareLocalWrite(ctx context.Context, key string) (string, error)
}

func NormalizeKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", ErrUnsafeKey
	}
	if strings.Contains(trimmed, `\`) {
		return "", ErrUnsafeKey
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") || path.IsAbs(cleaned) {
		return "", ErrUnsafeKey
	}
	return strings.TrimPrefix(cleaned, "/"), nil
}
