package queries

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *ProgressTimelineService) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error) {
	events, err := s.repo.GetUserTimeline(ctx, userID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.TimelineResp{
		Events: make([]dto.TimelineEvent, len(events)),
	}
	for i, event := range events {
		resp.Events[i] = dto.TimelineEvent{
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
