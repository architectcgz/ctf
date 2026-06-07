package queries

import (
	"context"

	"ctf-platform/internal/apperror"
)

func (s *ParticipationService) ListAnnouncements(ctx context.Context, contestID int64) ([]*ContestAnnouncementResult, error) {
	if err := s.ensureContestExists(ctx, contestID); err != nil {
		return nil, err
	}

	announcements, err := s.repo.ListAnnouncements(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
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

func (s *ParticipationService) SyncAnnouncements(ctx context.Context, contestID int64, afterID *int64) (*ContestAnnouncementSyncResult, error) {
	if err := s.ensureContestExists(ctx, contestID); err != nil {
		return nil, err
	}

	if afterID == nil {
		cursor, err := s.repo.LatestAnnouncementSyncCursor(ctx, contestID)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		return &ContestAnnouncementSyncResult{
			Events:     []*ContestAnnouncementSyncEventResult{},
			NextCursor: cursor,
			HasMore:    false,
		}, nil
	}

	const pageSize = 100
	rows, err := s.repo.ListAnnouncementSyncEvents(ctx, contestID, *afterID, pageSize+1)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	events := make([]*ContestAnnouncementSyncEventResult, 0, len(rows))
	nextCursor := *afterID
	for _, row := range rows {
		if row == nil {
			continue
		}

		nextCursor = row.Cursor
		event := &ContestAnnouncementSyncEventResult{
			Cursor:         row.Cursor,
			Type:           row.MessageType,
			AnnouncementID: row.AnnouncementID,
			OccurredAt:     row.OccurredAt,
		}
		if row.Announcement != nil {
			event.Announcement = &ContestAnnouncementResult{
				ID:        row.Announcement.ID,
				Title:     row.Announcement.Title,
				Content:   row.Announcement.Content,
				CreatedAt: row.Announcement.CreatedAt,
			}
		}
		events = append(events, event)
	}

	return &ContestAnnouncementSyncResult{
		Events:     events,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
