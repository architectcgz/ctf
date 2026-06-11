package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const defaultBootIDPath = "/proc/sys/kernel/random/boot_id"

type HostBootIDReader struct {
	path string
}

func NewHostBootIDReader(path string) *HostBootIDReader {
	return &HostBootIDReader{path: strings.TrimSpace(path)}
}

func (r *HostBootIDReader) ReadBootID(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	content, err := os.ReadFile(r.pathOrDefault())
	if err != nil {
		return "", err
	}
	bootID := strings.TrimSpace(string(content))
	if bootID == "" {
		return "", fmt.Errorf("boot id is empty")
	}
	return bootID, nil
}

func (r *HostBootIDReader) pathOrDefault() string {
	if r != nil && strings.TrimSpace(r.path) != "" {
		return strings.TrimSpace(r.path)
	}
	return defaultBootIDPath
}
