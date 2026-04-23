package queries

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
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

	frozen, key, err := s.resolveScoreboardKey(ctx, contest, contestID, live, time.Now())
	if err != nil {
		return nil, err
	}

	results, err := s.redis.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	results, teamIDs := filterScoreboardResults(s.logger, contestID, results)
	total := int64(len(results))

	start, stop := scoreboardPageBounds(page, pageSize)
	if start >= total {
		return &dto.ScoreboardResp{
			Contest: &dto.ScoreboardContestInfo{
				ID:        contest.ID,
				Title:     contest.Title,
				Status:    contest.Status,
				StartedAt: contest.StartTime,
				EndsAt:    contest.EndTime,
			},
			Scoreboard: &dto.ScoreboardPage{
				List:     []*dto.ScoreboardItem{},
				Total:    total,
				Page:     page,
				PageSize: pageSize,
			},
			Frozen: frozen,
		}, nil
	}
	if stop >= total {
		stop = total - 1
	}
	results = results[start : stop+1]
	teamIDs = teamIDs[start : stop+1]

	teams, err := s.repo.FindTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	statsMap, err := s.repo.FindScoreboardTeamStats(ctx, contestID, contest.Mode, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	items := buildScoreboardItems(start, results, teamIDs, teams, statsMap)

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
