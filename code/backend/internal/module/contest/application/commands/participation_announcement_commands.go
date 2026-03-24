package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req *dto.CreateContestAnnouncementReq) (*dto.ContestAnnouncementResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
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
	return &dto.ContestAnnouncementResp{
		ID:        item.ID,
		Title:     item.Title,
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (s *ParticipationService) DeleteAnnouncement(ctx context.Context, contestID, announcementID int64) error {
	deleted, err := s.repo.DeleteAnnouncement(ctx, contestID, announcementID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !deleted {
		return errcode.ErrContestAnnouncementNotFound
	}
	return nil
}
