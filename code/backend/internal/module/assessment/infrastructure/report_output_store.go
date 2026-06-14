package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	assessmentports "ctf-platform/internal/module/assessment/ports"
	platformstorage "ctf-platform/internal/platform/storage"
)

type ReportOutputStore struct {
	store     platformstorage.LocalWritableStore
	namespace string
}

var _ assessmentports.ReportOutputStore = (*ReportOutputStore)(nil)

func NewReportOutputStore(store platformstorage.LocalWritableStore, namespace string) *ReportOutputStore {
	return &ReportOutputStore{store: store, namespace: strings.TrimSpace(namespace)}
}

func (s *ReportOutputStore) PrepareReportOutput(ctx context.Context, fileName string) (*assessmentports.ReportOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fileName = strings.TrimSpace(fileName)
	if fileName == "" || strings.Contains(fileName, "/") || strings.Contains(fileName, `\`) {
		return nil, assessmentports.ErrAssessmentReportOutputUnsafePath
	}
	storageKey := s.reportKey(fileName)
	localPath, err := s.store.PrepareLocalWrite(ctx, storageKey)
	if err != nil {
		if errors.Is(err, platformstorage.ErrUnsafeKey) {
			return nil, assessmentports.ErrAssessmentReportOutputUnsafePath
		}
		return nil, fmt.Errorf("prepare report output: %w", err)
	}
	return &assessmentports.ReportOutput{StorageKey: storageKey, LocalPath: localPath}, nil
}

func (s *ReportOutputStore) OpenReportDownload(ctx context.Context, storageKey string) (*assessmentports.ReportDownloadStream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reader, err := s.store.Open(ctx, storageKey)
	if err != nil {
		if errors.Is(err, platformstorage.ErrUnsafeKey) {
			return nil, assessmentports.ErrAssessmentReportOutputUnsafePath
		}
		if errors.Is(err, platformstorage.ErrNotFound) {
			return nil, assessmentports.ErrAssessmentReportOutputNotFound
		}
		return nil, err
	}
	info, err := s.store.Stat(ctx, storageKey)
	if err != nil {
		_ = reader.Close()
		if errors.Is(err, platformstorage.ErrUnsafeKey) {
			return nil, assessmentports.ErrAssessmentReportOutputUnsafePath
		}
		if errors.Is(err, platformstorage.ErrNotFound) {
			return nil, assessmentports.ErrAssessmentReportOutputNotFound
		}
		return nil, err
	}
	return &assessmentports.ReportDownloadStream{StorageKey: info.Key, Reader: reader, Size: info.Size}, nil
}


func (s *ReportOutputStore) reportKey(fileName string) string {
	namespace := strings.TrimSpace(s.namespace)
	if namespace == "" {
		return fileName
	}
	return strings.TrimSuffix(namespace, "/") + "/" + fileName
}
