package sharedfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	platformstorage "ctf-platform/internal/platform/storage"
)

type Store struct {
	root string
}

var _ platformstorage.LocalWritableStore = (*Store)(nil)

func NewStore(root string) *Store {
	return &Store{root: strings.TrimSpace(root)}
}

func (s *Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, platformstorage.ErrNotFound
		}
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, platformstorage.ErrNotFound
	}
	return file, nil
}

func (s *Store) Put(ctx context.Context, key string, reader io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if reader == nil {
		return fmt.Errorf("sharedfs put requires reader")
	}
	path, err := s.PrepareLocalWrite(ctx, key)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return err
	}
	return nil
}

func (s *Store) Stat(ctx context.Context, key string) (*platformstorage.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, platformstorage.ErrNotFound
		}
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, platformstorage.ErrNotFound
	}
	normalizedKey, err := platformstorage.NormalizeKey(key)
	if err != nil {
		return nil, err
	}
	return &platformstorage.ObjectInfo{Key: normalizedKey, Size: info.Size(), ModTime: info.ModTime()}, nil
}

func (s *Store) PrepareLocalWrite(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	path, err := s.resolvePath(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Store) resolvePath(key string) (string, error) {
	if strings.TrimSpace(s.root) == "" {
		return "", fmt.Errorf("sharedfs root is empty")
	}
	if strings.Contains(key, `\`) {
		return "", platformstorage.ErrUnsafeKey
	}
	normalizedKey, err := platformstorage.NormalizeKey(key)
	if err != nil {
		return "", err
	}
	root := filepath.Clean(strings.TrimSpace(s.root))
	fullPath := filepath.Join(root, filepath.FromSlash(normalizedKey))
	cleanPath := filepath.Clean(fullPath)
	prefix := root + string(os.PathSeparator)
	if cleanPath != root && !strings.HasPrefix(cleanPath, prefix) {
		return "", platformstorage.ErrUnsafeKey
	}
	return cleanPath, nil
}
