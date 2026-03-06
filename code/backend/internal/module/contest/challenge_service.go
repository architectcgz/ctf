package contest

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"

	"gorm.io/gorm"
)

type ChallengeService struct {
	repo          *ChallengeRepository
	challengeRepo ChallengeRepoInterface
	contestRepo   Repository
}

type ChallengeRepoInterface interface {
	FindByID(id int64) (*model.Challenge, error)
	BatchGetSolvedStatus(userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(challengeIDs []int64) (map[int64]int64, error)
}

func NewChallengeService(repo *ChallengeRepository, challengeRepo ChallengeRepoInterface, contestRepo Repository) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
	}
}

func (s *ChallengeService) AddChallengeToContest(ctx context.Context, contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if s.isContestImmutable(contest) {
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

	return s.toResp(cc), nil
}

func (s *ChallengeService) RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if s.isContestImmutable(contest) {
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
		if errors.Is(err, ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if s.isContestImmutable(contest) {
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

func (s *ChallengeService) ListAdminChallenges(ctx context.Context, contestID int64) ([]*dto.ContestChallengeResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, false)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeResp, len(challenges))
	for i, c := range challenges {
		result[i] = s.toResp(c)
	}
	return result, nil
}

func (s *ChallengeService) GetContestChallenges(ctx context.Context, userID, contestID int64) ([]*dto.ContestChallengeInfo, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestChallengeVisible
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, true)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(challenges) == 0 {
		return []*dto.ContestChallengeInfo{}, nil
	}

	challengeIDs := make([]int64, 0, len(challenges))
	for _, item := range challenges {
		challengeIDs = append(challengeIDs, item.ChallengeID)
	}

	solvedMap, err := s.challengeRepo.BatchGetSolvedStatus(userID, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	solvedCountMap, err := s.challengeRepo.BatchGetSolvedCount(challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeInfo, 0, len(challenges))
	for _, item := range challenges {
		challenge, err := s.challengeRepo.FindByID(item.ChallengeID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		result = append(result, &dto.ContestChallengeInfo{
			ID:          item.ID,
			ChallengeID: item.ChallengeID,
			Title:       challenge.Title,
			Category:    challenge.Category,
			Difficulty:  challenge.Difficulty,
			Points:      item.Points,
			Order:       item.Order,
			SolvedCount: solvedCountMap[item.ChallengeID],
			IsSolved:    solvedMap[item.ChallengeID],
		})
	}
	return result, nil
}

func (s *ChallengeService) isContestImmutable(contest *model.Contest) bool {
	return contest.Status == model.ContestStatusRunning ||
		contest.Status == model.ContestStatusFrozen ||
		contest.Status == model.ContestStatusEnded
}

func (s *ChallengeService) toResp(cc *model.ContestChallenge) *dto.ContestChallengeResp {
	return &dto.ContestChallengeResp{
		ID:          cc.ID,
		ContestID:   cc.ContestID,
		ChallengeID: cc.ChallengeID,
		Points:      cc.Points,
		Order:       cc.Order,
		IsVisible:   cc.IsVisible,
		CreatedAt:   cc.CreatedAt,
	}
}
