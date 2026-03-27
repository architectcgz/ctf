package commands

import "context"

func (s *SubmissionService) syncCorrectSubmissionScoreboard(ctx context.Context, contestID *int64, teamScoreDeltas map[int64]int) error {
	if contestID == nil || s.scoreboardService == nil {
		return nil
	}

	for affectedTeamID, delta := range teamScoreDeltas {
		if delta == 0 {
			continue
		}
		if err := s.scoreboardService.UpdateScore(ctx, *contestID, affectedTeamID, float64(delta)); err != nil {
			return s.scoreboardService.RebuildScoreboard(ctx, *contestID)
		}
	}

	return nil
}
