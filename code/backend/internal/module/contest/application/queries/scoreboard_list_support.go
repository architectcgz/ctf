package queries

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) resolveScoreboardKey(ctx context.Context, contest *model.Contest, contestID int64, live bool, now time.Time) (bool, string, error) {
	frozen := !live && contestdomain.IsFrozenContest(contest, now)
	key := rediskeys.RankContestTeamKey(contestID)
	if !frozen {
		return false, key, nil
	}

	key = rediskeys.RankContestFrozenKey(contestID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, "", errcode.ErrInternal.WithCause(err)
	}
	if exists == 0 {
		if snapshotErr := s.createSnapshotFromLive(ctx, contestID); snapshotErr != nil {
			return false, "", snapshotErr
		}
	}
	return true, key, nil
}

func scoreboardPageBounds(page, pageSize int) (int64, int64) {
	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	return start, stop
}

func filterScoreboardResults(logger *zap.Logger, contestID int64, results []redislib.Z) ([]redislib.Z, []int64) {
	filtered := make([]redislib.Z, 0, len(results))
	teamIDs := make([]int64, 0, len(results))
	for _, item := range results {
		teamID, ok := contestdomain.ParseMemberToTeamID(item.Member)
		if !ok {
			if logger != nil {
				logger.Warn("跳过非法榜单成员",
					zap.Int64("contest_id", contestID),
					zap.Any("member", item.Member))
			}
			continue
		}
		filtered = append(filtered, item)
		teamIDs = append(teamIDs, teamID)
	}
	return filtered, teamIDs
}

func buildScoreboardItems(
	logger *zap.Logger,
	contestID int64,
	start int64,
	results []redislib.Z,
	teamIDs []int64,
	teams []*model.Team,
	statsMap map[int64]contestports.ScoreboardTeamStats,
) []*ScoreboardItemResult {
	teamMap := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	items := make([]*ScoreboardItemResult, 0, len(results))
	for idx, item := range results {
		teamID := teamIDs[idx]
		team := teamMap[teamID]
		if team == nil {
			if logger != nil {
				logger.Warn("跳过缺失的排行榜队伍",
					zap.Int64("contest_id", contestID),
					zap.Int64("team_id", teamID))
			}
			continue
		}
		stats := statsMap[teamID]
		items = append(items, &ScoreboardItemResult{
			Rank:             int(start) + len(items) + 1,
			TeamID:           teamID,
			Score:            item.Score,
			TeamName:         teamName(team),
			SolvedCount:      stats.SolvedCount,
			LastSubmissionAt: stats.LastSubmissionAt,
		})
	}
	return items
}
