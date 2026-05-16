package queries

import (
	"context"
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*ScoreboardResult, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, false)
}

func (s *ScoreboardService) GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*ScoreboardResult, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, true)
}

func (s *ScoreboardService) getScoreboard(ctx context.Context, contestID int64, page, pageSize int, live bool) (*ScoreboardResult, error) {
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

	frozen, results, err := s.resolveScoreboardMembers(ctx, contest, contestID, live, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	effectiveEnd := contestdomain.ContestEffectiveEndTime(contest)
	results, teamIDs := filterScoreboardResults(s.logger, contestID, results)

	teams, err := s.repo.FindTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	statsMap, err := s.repo.FindScoreboardTeamStats(ctx, contestID, contest.Mode, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	allItems := buildScoreboardItems(s.logger, contestID, 0, results, teamIDs, teams, statsMap)
	total := int64(len(allItems))

	start, stop := scoreboardPageBounds(page, pageSize)
	if start >= total {
		return &ScoreboardResult{
			Contest: &ScoreboardContestResult{
				ID:        contest.ID,
				Title:     contest.Title,
				Status:    contest.Status,
				StartedAt: contest.StartTime,
				EndsAt:    effectiveEnd,
			},
			Scoreboard: &ScoreboardPageResult{
				List:     []*ScoreboardItemResult{},
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
	items := allItems[start : stop+1]

	return &ScoreboardResult{
		Contest: &ScoreboardContestResult{
			ID:        contest.ID,
			Title:     contest.Title,
			Status:    contest.Status,
			StartedAt: contest.StartTime,
			EndsAt:    effectiveEnd,
		},
		Scoreboard: &ScoreboardPageResult{
			List:     items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Frozen: frozen,
	}, nil
}
