package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
	ctfws "ctf-platform/pkg/websocket"
)

func (s *ParticipationService) CreateAnnouncement(ctx context.Context, contestID, actorUserID int64, req CreateAnnouncementInput) (*dto.ContestAnnouncementResp, error) {
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
	result := contestResponseMapperInst.ToContestAnnouncementRespBasePtr(item)
	broadcastContestRealtimeEvent(s.broadcaster, contestports.AnnouncementChannel(contestID), ctfws.Envelope{
		Type: "contest.announcement.created",
		Payload: map[string]any{
			"contest_id": contestID,
			"announcement": map[string]any{
				"id":         result.ID,
				"title":      result.Title,
				"content":    result.Content,
				"created_at": result.CreatedAt,
			},
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
	broadcastContestRealtimeEvent(s.broadcaster, contestports.AnnouncementChannel(contestID), ctfws.Envelope{
		Type: "contest.announcement.deleted",
		Payload: map[string]any{
			"contest_id":      contestID,
			"announcement_id": announcementID,
		},
	})
	return nil
}
