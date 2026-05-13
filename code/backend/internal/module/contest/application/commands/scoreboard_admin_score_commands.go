package commands

import (
	"context"

	"ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardAdminService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
	if err := s.stateStore.IncrementLiveTeamScore(ctx, contestID, teamID, points); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ScoreboardAdminService) RebuildScoreboard(ctx context.Context, contestID int64) error {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	entries := make([]contestports.ScoreboardTeamScoreEntry, 0, len(teams))
	for _, team := range teams {
		if team == nil || team.TotalScore <= 0 {
			continue
		}
		entries = append(entries, contestports.ScoreboardTeamScoreEntry{
			TeamID: team.ID,
			Score:  float64(team.TotalScore),
		})
	}
	if err := s.stateStore.ReplaceLiveScoreboard(ctx, contestID, entries); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ScoreboardAdminService) CalculateDynamicScoreWithBase(baseScore float64, solveCount int64) int {
	if s.cfg == nil {
		return domain.CalculateDynamicScore(baseScore, 0, 0, solveCount)
	}
	if baseScore <= 0 {
		baseScore = s.cfg.BaseScore
	}
	return domain.CalculateDynamicScore(baseScore, s.cfg.MinScore, s.cfg.Decay, solveCount)
}
