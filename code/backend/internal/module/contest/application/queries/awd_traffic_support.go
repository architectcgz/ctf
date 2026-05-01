package queries

import (
	"sort"
	"strings"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

func buildAWDTrafficEvents(records []contestports.AWDTrafficEventRecord) []*AWDTrafficEventResult {
	items := make([]*AWDTrafficEventResult, 0, len(records))
	for _, record := range records {
		source := strings.TrimSpace(record.Source)
		if source == "" {
			source = "runtime_proxy"
		}
		items = append(items, &AWDTrafficEventResult{
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

func filterAWDTrafficEvents(items []*AWDTrafficEventResult, req *ListAWDTrafficEventsInput) []*AWDTrafficEventResult {
	if req == nil {
		return items
	}
	pathKeyword := strings.ToLower(strings.TrimSpace(req.PathKeyword))
	filtered := make([]*AWDTrafficEventResult, 0, len(items))
	for _, item := range items {
		if req.AttackerTeamID > 0 && item.AttackerTeamID != req.AttackerTeamID {
			continue
		}
		if req.VictimTeamID > 0 && item.VictimTeamID != req.VictimTeamID {
			continue
		}
		if req.ServiceID > 0 && item.ServiceID != req.ServiceID {
			continue
		}
		if req.AWDChallengeID > 0 && item.AWDChallengeID != req.AWDChallengeID {
			continue
		}
		if req.StatusGroup != "" && !matchAWDTrafficStatusGroup(item.StatusCode, req.StatusGroup) {
			continue
		}
		if pathKeyword != "" && !strings.Contains(strings.ToLower(item.Path), pathKeyword) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func matchAWDTrafficStatusGroup(statusCode int, group string) bool {
	return awdTrafficStatusGroup(statusCode) == group
}

func paginateAWDTrafficEvents(items []*AWDTrafficEventResult, page, size int) ([]*AWDTrafficEventResult, int64, int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	total := int64(len(items))
	start := (page - 1) * size
	if start >= len(items) {
		return []*AWDTrafficEventResult{}, total, page, size
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, page, size
}

func buildAWDTrafficSummary(round *AWDRoundResult, items []*AWDTrafficEventResult) *AWDTrafficSummaryResult {
	summary := &AWDTrafficSummaryResult{
		Round:         round,
		Trend:         []*AWDTrafficTrendBucketResult{},
		TopAttackers:  []*AWDTrafficTopTeamResult{},
		TopVictims:    []*AWDTrafficTopTeamResult{},
		TopChallenges: []*AWDTrafficTopChallengeResult{},
		TopPaths:      []*AWDTrafficTopPathResult{},
		TopErrorPaths: []*AWDTrafficTopPathResult{},
	}
	if round != nil {
		summary.ContestID = round.ContestID
		summary.RoundID = round.ID
	}
	if len(items) == 0 {
		return summary
	}
	latestEventAt := items[0].OccurredAt
	summary.LatestEventAt = &latestEventAt

	attackerSeen := make(map[int64]struct{})
	victimSeen := make(map[int64]struct{})
	paths := make(map[string]struct{})
	attackerAgg := make(map[int64]*AWDTrafficTopTeamResult)
	victimAgg := make(map[int64]*AWDTrafficTopTeamResult)
	challengeAgg := make(map[int64]*AWDTrafficTopChallengeResult)
	pathAgg := make(map[string]*AWDTrafficTopPathResult)
	trendAgg := make(map[time.Time]*AWDTrafficTrendBucketResult)

	for _, item := range items {
		summary.TotalRequests++
		if item.StatusCode >= 400 {
			summary.ErrorRequests++
		}
		if item.AttackerTeamID > 0 {
			attackerSeen[item.AttackerTeamID] = struct{}{}
			entry := attackerAgg[item.AttackerTeamID]
			if entry == nil {
				entry = &AWDTrafficTopTeamResult{TeamID: item.AttackerTeamID, TeamName: item.AttackerTeam}
				attackerAgg[item.AttackerTeamID] = entry
			}
			entry.RequestCount++
			if item.StatusCode >= 400 {
				entry.ErrorCount++
			}
		}
		if item.VictimTeamID > 0 {
			victimSeen[item.VictimTeamID] = struct{}{}
			entry := victimAgg[item.VictimTeamID]
			if entry == nil {
				entry = &AWDTrafficTopTeamResult{TeamID: item.VictimTeamID, TeamName: item.VictimTeam}
				victimAgg[item.VictimTeamID] = entry
			}
			entry.RequestCount++
			if item.StatusCode >= 400 {
				entry.ErrorCount++
			}
		}
		if strings.TrimSpace(item.Path) != "" {
			paths[item.Path] = struct{}{}
			entry := pathAgg[item.Path]
			if entry == nil {
				entry = &AWDTrafficTopPathResult{Path: item.Path}
				pathAgg[item.Path] = entry
			}
			entry.RequestCount++
			entry.LastStatusCode = item.StatusCode
			if item.StatusCode >= 400 {
				entry.ErrorCount++
			}
		}
		entry := challengeAgg[item.AWDChallengeID]
		if entry == nil {
			entry = &AWDTrafficTopChallengeResult{AWDChallengeID: item.AWDChallengeID, AWDChallengeTitle: item.AWDChallengeTitle}
			challengeAgg[item.AWDChallengeID] = entry
		}
		entry.RequestCount++
		if item.StatusCode >= 400 {
			entry.ErrorCount++
		}

		bucket := item.OccurredAt.Truncate(time.Minute)
		trendEntry := trendAgg[bucket]
		if trendEntry == nil {
			trendEntry = &AWDTrafficTrendBucketResult{BucketStart: bucket}
			trendAgg[bucket] = trendEntry
		}
		trendEntry.RequestCount++
		if item.StatusCode >= 400 {
			trendEntry.ErrorCount++
		}
	}

	summary.ActiveAttackerTeams = len(attackerSeen)
	summary.TargetedTeams = len(victimSeen)
	summary.UniquePathCount = len(paths)
	summary.TopAttackers = sortAWDTrafficTeams(attackerAgg)
	summary.TopVictims = sortAWDTrafficTeams(victimAgg)
	summary.TopChallenges = sortAWDTrafficChallenges(challengeAgg)
	summary.TopPaths = sortAWDTrafficPaths(pathAgg)
	summary.TopErrorPaths = summary.TopPaths
	summary.Trend = sortAWDTrafficTrend(trendAgg)
	return summary
}

func awdTrafficStatusGroup(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "success"
	case statusCode >= 300 && statusCode < 400:
		return "redirect"
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return ""
	}
}

func sortAWDTrafficTeams(items map[int64]*AWDTrafficTopTeamResult) []*AWDTrafficTopTeamResult {
	list := make([]*AWDTrafficTopTeamResult, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].TeamID < list[j].TeamID
	})
	return limitAWDTrafficTeams(list, 5)
}

func limitAWDTrafficTeams(items []*AWDTrafficTopTeamResult, limit int) []*AWDTrafficTopTeamResult {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func sortAWDTrafficChallenges(items map[int64]*AWDTrafficTopChallengeResult) []*AWDTrafficTopChallengeResult {
	list := make([]*AWDTrafficTopChallengeResult, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].AWDChallengeID < list[j].AWDChallengeID
	})
	if len(list) > 5 {
		return list[:5]
	}
	return list
}

func sortAWDTrafficPaths(items map[string]*AWDTrafficTopPathResult) []*AWDTrafficTopPathResult {
	list := make([]*AWDTrafficTopPathResult, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].ErrorCount != list[j].ErrorCount {
			return list[i].ErrorCount > list[j].ErrorCount
		}
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].Path < list[j].Path
	})
	if len(list) > 5 {
		return list[:5]
	}
	return list
}

func sortAWDTrafficTrend(items map[time.Time]*AWDTrafficTrendBucketResult) []*AWDTrafficTrendBucketResult {
	list := make([]*AWDTrafficTrendBucketResult, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].BucketStart.Before(list[j].BucketStart)
	})
	if len(list) > 12 {
		return list[len(list)-12:]
	}
	return list
}
