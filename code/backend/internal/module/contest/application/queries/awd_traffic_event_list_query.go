package queries

import (
	"context"
	"sort"
	"strings"

	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListTrafficEvents(ctx context.Context, contestID, roundID int64, req ListAWDTrafficEventsInput) (*AWDTrafficEventPageResult, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}
	records, err := s.repo.ListTrafficEvents(ctx, contestID, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	filtered := filterAWDTrafficEventResults(buildAWDTrafficEventResults(records), req)
	pageItems, total, page, size := paginateAWDTrafficEventResults(filtered, req.Page, req.Size)
	return &AWDTrafficEventPageResult{
		List:     pageItems,
		Total:    total,
		Page:     page,
		PageSize: size,
	}, nil
}

func buildAWDTrafficEventResults(records []contestports.AWDTrafficEventRecord) []AWDTrafficEventResult {
	items := make([]AWDTrafficEventResult, 0, len(records))
	for _, record := range records {
		source := strings.TrimSpace(record.Source)
		if source == "" {
			source = "runtime_proxy"
		}
		items = append(items, AWDTrafficEventResult{
			ID:                record.ID,
			ContestID:         record.ContestID,
			RoundID:           record.RoundID,
			AttackerTeamID:    record.AttackerTeamID,
			AttackerTeam:      record.AttackerTeamName,
			AttackerTeamName:  record.AttackerTeamName,
			VictimTeamID:      record.VictimTeamID,
			VictimTeam:        record.VictimTeamName,
			VictimTeamName:    record.VictimTeamName,
			ServiceID:         record.ServiceID,
			AWDChallengeID:    record.AWDChallengeID,
			AWDChallengeTitle: record.AWDChallengeTitle,
			Method:            strings.TrimSpace(record.Method),
			Path:              strings.TrimSpace(record.Path),
			StatusCode:        record.StatusCode,
			StatusGroup:       awdTrafficStatusGroup(record.StatusCode),
			IsError:           record.StatusCode >= 400,
			Source:            source,
			OccurredAt:        record.OccurredAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].OccurredAt.Equal(items[j].OccurredAt) {
			if items[i].AttackerTeamID != items[j].AttackerTeamID {
				return items[i].AttackerTeamID < items[j].AttackerTeamID
			}
			return items[i].Path < items[j].Path
		}
		return items[i].OccurredAt.After(items[j].OccurredAt)
	})
	return items
}

func filterAWDTrafficEventResults(items []AWDTrafficEventResult, input ListAWDTrafficEventsInput) []AWDTrafficEventResult {
	pathKeyword := strings.ToLower(strings.TrimSpace(input.PathKeyword))
	filtered := make([]AWDTrafficEventResult, 0, len(items))
	for _, item := range items {
		if input.AttackerTeamID > 0 && item.AttackerTeamID != input.AttackerTeamID {
			continue
		}
		if input.VictimTeamID > 0 && item.VictimTeamID != input.VictimTeamID {
			continue
		}
		if input.ServiceID > 0 && item.ServiceID != input.ServiceID {
			continue
		}
		if input.AWDChallengeID > 0 && item.AWDChallengeID != input.AWDChallengeID {
			continue
		}
		if input.StatusGroup != "" && !matchAWDTrafficStatusGroup(item.StatusCode, input.StatusGroup) {
			continue
		}
		if pathKeyword != "" && !strings.Contains(strings.ToLower(item.Path), pathKeyword) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func paginateAWDTrafficEventResults(items []AWDTrafficEventResult, page, size int) ([]AWDTrafficEventResult, int64, int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	total := int64(len(items))
	start := (page - 1) * size
	if start >= len(items) {
		return []AWDTrafficEventResult{}, total, page, size
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, page, size
}
