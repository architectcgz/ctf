package queries

import (
	"context"

	"ctf-platform/internal/apperror"
	practiceports "ctf-platform/internal/module/practice/ports"
)

func (s *ProgressTimelineService) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error) {
	events, err := s.repo.GetUserTimeline(ctx, userID, limit, offset)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	resp := &practiceports.TimelineSnapshot{
		Events: make([]practiceports.TimelineEventSnapshot, len(events)),
	}
	for i, event := range events {
		resp.Events[i] = practiceports.TimelineEventSnapshot{
			Type:        event.Type,
			ChallengeID: event.ChallengeID,
			Title:       event.Title,
			Timestamp:   event.Timestamp,
			IsCorrect:   event.IsCorrect,
			Points:      event.Points,
			Detail:      event.Detail,
		}
	}

	return resp, nil
}
