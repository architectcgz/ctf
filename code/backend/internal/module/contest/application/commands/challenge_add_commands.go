package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) AddChallengeToContest(ctx context.Context, contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error) {
	if _, err := s.ensureMutableContest(ctx, contestID); err != nil {
		return nil, err
	}

	challenge, err := s.challengeRepo.FindByID(ctx, req.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublished
	}

	exists, err := s.repo.Exists(ctx, contestID, req.ChallengeID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if exists {
		return nil, errcode.ErrChallengeAlreadyAdded
	}

	points := req.Points
	if points == 0 {
		points = challenge.Points
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	cc := &model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: req.ChallengeID,
		Points:      points,
		Order:       req.Order,
		IsVisible:   isVisible,
	}
	if err := s.repo.AddChallenge(ctx, cc); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contestdomain.ContestChallengeRespFromModel(cc, challenge), nil
}
