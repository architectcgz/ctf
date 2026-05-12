package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req CreateAnnouncementInput) (*dto.ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	item := &model.ContestAnnouncement{
		ContestID: contestID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedBy: &actorUserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateAnnouncement(ctx, item); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	result := contestResponseMapperInst.ToContestAnnouncementRespBasePtr(item)
	publishContestWeakEvent(ctx, s.eventBus, platformevents.Event{
		Name: contestcontracts.EventAnnouncementCreated,
		Payload: contestcontracts.AnnouncementCreatedEvent{
			ContestID:      contestID,
			AnnouncementID: result.ID,
			Title:          result.Title,
			Content:        result.Content,
			CreatedAt:      result.CreatedAt,
			OccurredAt:     contestEventTimestamp(result.CreatedAt),
		},
	})
	return result, nil
}

func (s *ParticipationService) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error {
	deleted, err := s.repo.DeleteAnnouncement(ctx, contestID, announcementID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !deleted {
		return errcode.ErrContestAnnouncementNotFound
	}
	publishContestWeakEvent(ctx, s.eventBus, platformevents.Event{
		Name: contestcontracts.EventAnnouncementDeleted,
		Payload: contestcontracts.AnnouncementDeletedEvent{
			ContestID:      contestID,
			AnnouncementID: announcementID,
			OccurredAt:     contestEventTimestamp(time.Now().UTC()),
		},
	})
	return nil
}
