package commands

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
	ctfws "ctf-platform/pkg/websocket"
)

func (s *SubmissionService) syncCorrectSubmissionScoreboard(ctx context.Context, contestID *int64, teamScoreDeltas map[int64]int) error {
	if contestID == nil || s.scoreboardService == nil {
		return nil
	}

	updated := false
	for affectedTeamID, delta := range teamScoreDeltas {
		if delta == 0 {
			continue
		}
		if err := s.scoreboardService.UpdateScore(ctx, *contestID, affectedTeamID, float64(delta)); err != nil {
			if err := s.scoreboardService.RebuildScoreboard(ctx, *contestID); err != nil {
				return err
			}
			updated = true
			break
		}
		updated = true
	}

	if updated {
		broadcastContestRealtimeEvent(s.broadcaster, contestports.ScoreboardChannel(*contestID), ctfws.Envelope{
			Type: "scoreboard.updated",
			Payload: map[string]any{
				"contest_id": *contestID,
			},
		})
	}

	return nil
}
