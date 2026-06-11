package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	assessmentports "ctf-platform/internal/module/assessment/ports"
)

type ReportOutputStore struct {
	storageDir string
}

var _ assessmentports.ReportOutputStore = (*ReportOutputStore)(nil)

func NewReportOutputStore(storageDir string) *ReportOutputStore {
	return &ReportOutputStore{storageDir: strings.TrimSpace(storageDir)}
}

func (s *ReportOutputStore) PrepareReportOutput(ctx context.Context, fileName string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" || fileName != filepath.Base(fileName) {
		return "", assessmentports.ErrAssessmentReportOutputUnsafePath
	}
	storageDir := filepath.Clean(s.storageDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return "", fmt.Errorf("create report output dir: %w", err)
	}
	return filepath.Join(storageDir, fileName), nil
}

func (s *ReportOutputStore) ResolveReportDownloadPath(ctx context.Context, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	filePath, err := s.safeReportPath(path)
	if err != nil {
		return "", err
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return "", assessmentports.ErrAssessmentReportOutputNotFound
		}
		return "", statErr
	}
	return filePath, nil
}

func (s *ReportOutputStore) safeReportPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absStorage, err := filepath.Abs(s.storageDir)
	if err != nil {
		return "", err
	}
	prefix := absStorage + string(os.PathSeparator)
	if absPath != absStorage && !strings.HasPrefix(absPath, prefix) {
		return "", assessmentports.ErrAssessmentReportOutputUnsafePath
	}
	return absPath, nil
}
