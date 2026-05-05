package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) ListAnnouncements(ctx context.Context, contestID int64) ([]*ContestAnnouncementResult, error) {
	if err := s.ensureContestExists(ctx, contestID); err != nil {
		return nil, err
	}

	announcements, err := s.repo.ListAnnouncements(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*ContestAnnouncementResult, 0, len(announcements))
	for _, item := range announcements {
		result = append(result, &ContestAnnouncementResult{
			ID:        item.ID,
			Title:     item.Title,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		})
	}
	return result, nil
}
