package queries

import (
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	contestports "ctf-platform/internal/module/contest/ports"
)

func buildAWDTrafficEvents(records []contestports.AWDTrafficEventRecord) []*dto.AWDTrafficEventResp {
	items := make([]*dto.AWDTrafficEventResp, 0, len(records))
	for _, record := range records {
		source := strings.TrimSpace(record.Source)
		if source == "" {
			source = "runtime_proxy"
		}
		items = append(items, &dto.AWDTrafficEventResp{
			ID:               record.ID,
			ContestID:        record.ContestID,
			RoundID:          record.RoundID,
			AttackerTeamID:   record.AttackerTeamID,
			AttackerTeam:     record.AttackerTeamName,
			AttackerTeamName: record.AttackerTeamName,
			VictimTeamID:     record.VictimTeamID,
			VictimTeam:       record.VictimTeamName,
			VictimTeamName:   record.VictimTeamName,
			ServiceID:        record.ServiceID,
			ChallengeID:      record.ChallengeID,
			ChallengeTitle:   record.ChallengeTitle,
			Method:           strings.TrimSpace(record.Method),
			Path:             strings.TrimSpace(record.Path),
			StatusCode:       record.StatusCode,
			StatusGroup:      awdTrafficStatusGroup(record.StatusCode),
			IsError:          record.StatusCode >= 400,
			Source:           source,
			OccurredAt:       record.OccurredAt,
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

func filterAWDTrafficEvents(items []*dto.AWDTrafficEventResp, req *dto.ListAWDTrafficEventsReq) []*dto.AWDTrafficEventResp {
	if req == nil {
		return items
	}
	pathKeyword := strings.ToLower(strings.TrimSpace(req.PathKeyword))
	filtered := make([]*dto.AWDTrafficEventResp, 0, len(items))
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
		if req.ChallengeID > 0 && item.ChallengeID != req.ChallengeID {
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

func paginateAWDTrafficEvents(items []*dto.AWDTrafficEventResp, page, size int) ([]*dto.AWDTrafficEventResp, int64, int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	total := int64(len(items))
	start := (page - 1) * size
	if start >= len(items) {
		return []*dto.AWDTrafficEventResp{}, total, page, size
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, page, size
}

func buildAWDTrafficSummary(round *dto.AWDRoundResp, items []*dto.AWDTrafficEventResp) *dto.AWDTrafficSummaryResp {
	summary := &dto.AWDTrafficSummaryResp{
		Round:         round,
		Trend:         []*dto.AWDTrafficTrendBucketResp{},
		TopAttackers:  []*dto.AWDTrafficTopTeamResp{},
		TopVictims:    []*dto.AWDTrafficTopTeamResp{},
		TopChallenges: []*dto.AWDTrafficTopChallengeResp{},
		TopPaths:      []*dto.AWDTrafficTopPathResp{},
		TopErrorPaths: []*dto.AWDTrafficTopPathResp{},
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
	attackerAgg := make(map[int64]*dto.AWDTrafficTopTeamResp)
	victimAgg := make(map[int64]*dto.AWDTrafficTopTeamResp)
	challengeAgg := make(map[int64]*dto.AWDTrafficTopChallengeResp)
	pathAgg := make(map[string]*dto.AWDTrafficTopPathResp)
	trendAgg := make(map[time.Time]*dto.AWDTrafficTrendBucketResp)

	for _, item := range items {
		summary.TotalRequests++
		if item.StatusCode >= 400 {
			summary.ErrorRequests++
		}
		if item.AttackerTeamID > 0 {
			attackerSeen[item.AttackerTeamID] = struct{}{}
			entry := attackerAgg[item.AttackerTeamID]
			if entry == nil {
				entry = &dto.AWDTrafficTopTeamResp{TeamID: item.AttackerTeamID, TeamName: item.AttackerTeam}
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
				entry = &dto.AWDTrafficTopTeamResp{TeamID: item.VictimTeamID, TeamName: item.VictimTeam}
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
				entry = &dto.AWDTrafficTopPathResp{Path: item.Path}
				pathAgg[item.Path] = entry
			}
			entry.RequestCount++
			entry.LastStatusCode = item.StatusCode
			if item.StatusCode >= 400 {
				entry.ErrorCount++
			}
		}
		entry := challengeAgg[item.ChallengeID]
		if entry == nil {
			entry = &dto.AWDTrafficTopChallengeResp{ChallengeID: item.ChallengeID, ChallengeTitle: item.ChallengeTitle}
			challengeAgg[item.ChallengeID] = entry
		}
		entry.RequestCount++
		if item.StatusCode >= 400 {
			entry.ErrorCount++
		}

		bucket := item.OccurredAt.Truncate(time.Minute)
		trendEntry := trendAgg[bucket]
		if trendEntry == nil {
			trendEntry = &dto.AWDTrafficTrendBucketResp{BucketStart: bucket}
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

func sortAWDTrafficTeams(items map[int64]*dto.AWDTrafficTopTeamResp) []*dto.AWDTrafficTopTeamResp {
	list := make([]*dto.AWDTrafficTopTeamResp, 0, len(items))
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

func limitAWDTrafficTeams(items []*dto.AWDTrafficTopTeamResp, limit int) []*dto.AWDTrafficTopTeamResp {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func sortAWDTrafficChallenges(items map[int64]*dto.AWDTrafficTopChallengeResp) []*dto.AWDTrafficTopChallengeResp {
	list := make([]*dto.AWDTrafficTopChallengeResp, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].RequestCount != list[j].RequestCount {
			return list[i].RequestCount > list[j].RequestCount
		}
		return list[i].ChallengeID < list[j].ChallengeID
	})
	if len(list) > 5 {
		return list[:5]
	}
	return list
}

func sortAWDTrafficPaths(items map[string]*dto.AWDTrafficTopPathResp) []*dto.AWDTrafficTopPathResp {
	list := make([]*dto.AWDTrafficTopPathResp, 0, len(items))
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

func sortAWDTrafficTrend(items map[time.Time]*dto.AWDTrafficTrendBucketResp) []*dto.AWDTrafficTrendBucketResp {
	list := make([]*dto.AWDTrafficTrendBucketResp, 0, len(items))
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
