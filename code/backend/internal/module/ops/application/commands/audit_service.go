package commands

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	opsentity "ctf-platform/internal/module/ops/entity"
	opsports "ctf-platform/internal/module/ops/ports"
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
		return apperror.ErrInternal.WithCause(err)
	}

	logEntry := &opsentity.AuditLog{
		UserID:       entry.UserID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Detail:       string(detailJSON),
		IPAddress:    entry.IPAddress,
		UserAgent:    entry.UserAgent,
	}

	if err := s.repo.Create(ctx, logEntry); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}

	return nil
}
