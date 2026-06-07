package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ctf-platform/internal/apperror"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *ParticipationService) CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req CreateAnnouncementInput) (*ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	if s.announcementTxRunner == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("announcement transaction runner is not configured"))
	}

	now := time.Now().UTC()
	var result *ContestAnnouncementResp
	if err := s.announcementTxRunner.WithinAnnouncementTransaction(ctx, func(repo contestports.ContestParticipationAnnouncementTxRepository) error {
		item := &contestentity.ContestAnnouncement{
			ContestID: contestID,
			Title:     req.Title,
			Content:   req.Content,
			CreatedBy: &actorUserID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := repo.CreateAnnouncement(ctx, item); err != nil {
			return err
		}
		result = contestResponseMapperInst.ToContestAnnouncementRespBasePtr(item)
		return repo.EnqueueRealtimeRelay(ctx, contestcontracts.RelayAnnouncementCreated(contestcontracts.AnnouncementCreatedEvent{
			ContestID:      contestID,
			AnnouncementID: result.ID,
			Title:          result.Title,
			Content:        result.Content,
			CreatedAt:      result.CreatedAt,
			OccurredAt:     contestEventTimestamp(result.CreatedAt),
		}), fmt.Sprintf("contest:%d:announcement:%d:created", contestID, result.ID))
	}); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return result, nil
}

func (s *ParticipationService) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error {
	if s.announcementTxRunner == nil {
		return apperror.ErrInternal.WithCause(errors.New("announcement transaction runner is not configured"))
	}

	deleted := false
	occurredAt := contestEventTimestamp(time.Now().UTC())
	if err := s.announcementTxRunner.WithinAnnouncementTransaction(ctx, func(repo contestports.ContestParticipationAnnouncementTxRepository) error {
		var err error
		deleted, err = repo.DeleteAnnouncement(ctx, contestID, announcementID)
		if err != nil {
			return err
		}
		if !deleted {
			return nil
		}
		return repo.EnqueueRealtimeRelay(ctx, contestcontracts.RelayAnnouncementDeleted(contestcontracts.AnnouncementDeletedEvent{
			ContestID:      contestID,
			AnnouncementID: announcementID,
			OccurredAt:     occurredAt,
		}), fmt.Sprintf("contest:%d:announcement:%d:deleted", contestID, announcementID))
	}); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if !deleted {
		return contestcontracts.ErrContestAnnouncementNotFound
	}
	return nil
}
