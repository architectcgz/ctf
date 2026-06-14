package config

import (
	"path/filepath"
	"strings"
)

func (c *Config) SharedStoragePath(relative string) string {
	if c == nil {
		return filepath.Clean(relative)
	}
	trimmed := strings.TrimSpace(relative)
	if trimmed == "" {
		return filepath.Clean(trimmed)
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	root := strings.TrimSpace(c.SharedStorage.SharedFS.Root)
	if root == "" {
		return filepath.Clean(trimmed)
	}
	return filepath.Join(filepath.Clean(root), filepath.Clean(trimmed))
}
