package queries

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, false)
}

func (s *ScoreboardService) GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, true)
}

func (s *ScoreboardService) getScoreboard(ctx context.Context, contestID int64, page, pageSize int, live bool) (*dto.ScoreboardResp, error) {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == contestdomain.ErrContestNotFound {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	frozen := !live && contestdomain.IsFrozenContest(contest, time.Now())
	key := rediskeys.RankContestTeamKey(contestID)
	if frozen {
		key = rediskeys.RankContestFrozenKey(contestID)
		exists, existsErr := s.redis.Exists(ctx, key).Result()
		if existsErr != nil {
			return nil, errcode.ErrInternal.WithCause(existsErr)
		}
		if exists == 0 {
			if snapshotErr := s.createSnapshotFromLive(ctx, contestID); snapshotErr != nil {
				return nil, snapshotErr
			}
		}
	}

	total, err := s.redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	results, err := s.redis.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, 0, len(results))
	for _, item := range results {
		teamIDs = append(teamIDs, contestdomain.MemberToTeamID(item.Member))
	}

	teams, err := s.repo.FindTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	statsMap, err := s.repo.FindScoreboardTeamStats(ctx, contestID, contest.Mode, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	teamMap := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	items := make([]*dto.ScoreboardItem, 0, len(results))
	for idx, item := range results {
		teamID := teamIDs[idx]
		team := teamMap[teamID]
		stats := statsMap[teamID]
		items = append(items, &dto.ScoreboardItem{
			Rank:             int(start) + idx + 1,
			TeamID:           teamID,
			Score:            item.Score,
			TeamName:         teamName(team),
			SolvedCount:      stats.SolvedCount,
			LastSubmissionAt: stats.LastSubmissionAt,
		})
	}

	return &dto.ScoreboardResp{
		Contest: &dto.ScoreboardContestInfo{
			ID:        contest.ID,
			Title:     contest.Title,
			Status:    contest.Status,
			StartedAt: contest.StartTime,
			EndsAt:    contest.EndTime,
		},
		Scoreboard: &dto.ScoreboardPage{
			List:     items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Frozen: frozen,
	}, nil
}

func (s *ScoreboardService) GetTeamRank(ctx context.Context, contestID, teamID int64) (*dto.TeamRankResp, error) {
	key := rediskeys.RankContestTeamKey(contestID)
	score, err := s.redis.ZScore(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		if err == redislib.Nil {
			return &dto.TeamRankResp{TeamID: teamID, Rank: 0, Score: 0}, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rank, err := s.redis.ZRevRank(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TeamRankResp{
		TeamID: teamID,
		Rank:   int(rank) + 1,
		Score:  score,
	}, nil
}
