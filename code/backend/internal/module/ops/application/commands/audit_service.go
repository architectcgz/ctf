package commands

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
	"ctf-platform/pkg/errcode"
)

type AuditService struct {
	repo   opsports.AuditCommandRepository
	logger *zap.Logger
}

func NewAuditService(repo opsports.AuditCommandRepository, logger *zap.Logger) *AuditService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuditService) Record(ctx context.Context, entry auditlog.Entry) error {
	detail := entry.Detail
	if detail == nil {
		detail = map[string]any{}
	}

	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	logEntry := &model.AuditLog{
		UserID:       entry.UserID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Detail:       string(detailJSON),
		IPAddress:    entry.IPAddress,
		UserAgent:    entry.UserAgent,
	}

	if err := s.repo.Create(ctx, logEntry); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	return nil
}
