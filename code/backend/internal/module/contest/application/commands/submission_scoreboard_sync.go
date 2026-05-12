package commands

import (
	"context"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
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
		publishContestWeakEvent(ctx, s.eventBus, platformevents.Event{
			Name: contestcontracts.EventScoreboardUpdated,
			Payload: contestcontracts.ScoreboardUpdatedEvent{
				ContestID:  *contestID,
				OccurredAt: contestEventTimestamp(time.Now().UTC()),
			},
		})
	}

	return nil
}
