package queries

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	opsports "ctf-platform/internal/module/ops/ports"
	"ctf-platform/pkg/errcode"
)

type AuditService struct {
	repo       opsports.AuditRepository
	pagination config.PaginationConfig
	logger     *zap.Logger
}

func NewAuditService(repo opsports.AuditRepository, pagination config.PaginationConfig, logger *zap.Logger) *AuditService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AuditService{
		repo:       repo,
		pagination: pagination,
		logger:     logger,
	}
}

func (s *AuditService) ListAuditLogs(ctx context.Context, query *dto.AuditLogQuery) ([]dto.AuditLogItem, int64, int, int, error) {
	userID := query.UserID
	if userID == nil {
		userID = query.ActorUserID
	}

	startTime, endTime, err := parseAuditTimeRange(query.StartTime, query.EndTime)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = s.pagination.DefaultPageSize
	}
	if pageSize > s.pagination.MaxPageSize {
		pageSize = s.pagination.MaxPageSize
	}

	records, total, err := s.repo.List(ctx, opsports.AuditLogListFilter{
		UserID:       userID,
		Action:       query.Action,
		ResourceType: query.ResourceType,
		ResourceID:   query.ResourceID,
		StartTime:    startTime,
		EndTime:      endTime,
		Offset:       (page - 1) * pageSize,
		Limit:        pageSize,
	})
	if err != nil {
		return nil, 0, 0, 0, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.AuditLogItem, 0, len(records))
	for _, record := range records {
		item := dto.AuditLogItem{
			ID:            record.ID,
			Action:        record.Action,
			ResourceType:  record.ResourceType,
			ResourceID:    record.ResourceID,
			ActorUserID:   record.UserID,
			ActorUsername: record.ActorUsername,
			UserAgent:     record.UserAgent,
			CreatedAt:     record.CreatedAt,
		}
		if strings.TrimSpace(record.IPAddress) != "" {
			ip := record.IPAddress
			item.IP = &ip
		}
		if strings.TrimSpace(record.Detail) != "" {
			var detail map[string]any
			if err := json.Unmarshal([]byte(record.Detail), &detail); err != nil {
				s.logger.Warn("audit_detail_unmarshal_failed", zap.Int64("audit_log_id", record.ID), zap.Error(err))
				item.Detail = map[string]any{"raw_detail": record.Detail}
			} else if len(detail) > 0 {
				item.Detail = detail
				if item.ActorUsername == "" {
					if username, ok := detail["username"].(string); ok {
						item.ActorUsername = username
					}
				}
			}
		}
		items = append(items, item)
	}

	return items, total, page, pageSize, nil
}

func parseAuditTimeRange(start, end string) (*time.Time, *time.Time, error) {
	var startTime *time.Time
	var endTime *time.Time

	if strings.TrimSpace(start) != "" {
		parsed, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, nil, errcode.New(errcode.ErrInvalidParams.Code, "start_time 必须为 RFC3339 格式", errcode.ErrInvalidParams.HTTPStatus)
		}
		startTime = &parsed
	}
	if strings.TrimSpace(end) != "" {
		parsed, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return nil, nil, errcode.New(errcode.ErrInvalidParams.Code, "end_time 必须为 RFC3339 格式", errcode.ErrInvalidParams.HTTPStatus)
		}
		endTime = &parsed
	}
	if startTime != nil && endTime != nil && endTime.Before(*startTime) {
		return nil, nil, errcode.New(errcode.ErrInvalidParams.Code, "end_time 不能早于 start_time", errcode.ErrInvalidParams.HTTPStatus)
	}

	return startTime, endTime, nil
}
