package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ChallengeService struct {
	repo          contestports.ContestChallengeRepository
	challengeRepo challengecontracts.ContestChallengeContract
	contestRepo   contestports.ContestLookupRepository
}

func NewChallengeService(repo contestports.ContestChallengeRepository, challengeRepo challengecontracts.ContestChallengeContract, contestRepo contestports.ContestLookupRepository) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
	}
}

func (s *ChallengeService) AddChallengeToContest(ctx context.Context, contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contestdomain.IsContestImmutable(contest) {
		return nil, errcode.ErrContestImmutable
	}

	challenge, err := s.challengeRepo.FindByID(req.ChallengeID)
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

func (s *ChallengeService) RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if contestdomain.IsContestImmutable(contest) {
		return errcode.ErrContestImmutable
	}

	exists, err := s.repo.Exists(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}

	hasSubmissions, err := s.repo.HasSubmissions(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if hasSubmissions {
		return errcode.ErrContestChallengeHasSubs
	}

	return s.repo.RemoveChallenge(ctx, contestID, challengeID)
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, contestID, challengeID int64, req *dto.UpdateContestChallengeReq) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if contestdomain.IsContestImmutable(contest) {
		return errcode.ErrContestImmutable
	}

	exists, err := s.repo.Exists(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}

	updates := make(map[string]any)
	if req.Points != nil {
		updates["points"] = *req.Points
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}

	return s.repo.UpdateChallenge(ctx, contestID, challengeID, updates)
}
