package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"errors"

	"gorm.io/gorm"
)

type ChallengeService struct {
	repo          *ChallengeRepository
	challengeRepo ChallengeRepoInterface
	contestRepo   ContestRepoInterface
}

type ChallengeRepoInterface interface {
	FindByID(id int64) (*model.Challenge, error)
}

type ContestRepoInterface interface {
	FindByID(id int64) (*model.Contest, error)
}

func NewChallengeService(repo *ChallengeRepository, challengeRepo ChallengeRepoInterface, contestRepo ContestRepoInterface) *ChallengeService {
	return &ChallengeService{
		repo:          repo,
		challengeRepo: challengeRepo,
		contestRepo:   contestRepo,
	}
}

func (s *ChallengeService) AddChallengeToContest(contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error) {
	contest, err := s.contestRepo.FindByID(contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, err
	}

	if s.isContestModifiable(contest) {
		return nil, errcode.ErrContestRunning
	}

	challenge, err := s.challengeRepo.FindByID(req.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublished
	}

	exists, err := s.repo.Exists(contestID, req.ChallengeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.ErrChallengeAlreadyAdded
	}

	points := req.Points
	if points == 0 {
		points = challenge.Points
	}

	cc := &model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: req.ChallengeID,
		Points:      points,
		Order:       req.Order,
	}

	if err := s.repo.AddChallenge(cc); err != nil {
		return nil, err
	}

	return s.toResp(cc), nil
}

func (s *ChallengeService) RemoveChallengeFromContest(contestID, challengeID int64) error {
	contest, err := s.contestRepo.FindByID(contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrContestNotFound
		}
		return err
	}

	if s.isContestModifiable(contest) {
		return errcode.ErrContestRunning
	}

	exists, err := s.repo.Exists(contestID, challengeID)
	if err != nil {
		return err
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}

	return s.repo.RemoveChallenge(contestID, challengeID)
}

func (s *ChallengeService) UpdateChallengePoints(contestID, challengeID int64, req *dto.UpdateContestChallengeReq) error {
	contest, err := s.contestRepo.FindByID(contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrContestNotFound
		}
		return err
	}

	if s.isContestModifiable(contest) {
		return errcode.ErrContestRunning
	}

	exists, err := s.repo.Exists(contestID, challengeID)
	if err != nil {
		return err
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}

	return s.repo.UpdatePoints(contestID, challengeID, req.Points)
}

func (s *ChallengeService) GetContestChallenges(contestID int64) ([]*dto.ContestChallengeResp, error) {
	challenges, err := s.repo.ListChallenges(contestID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.ContestChallengeResp, len(challenges))
	for i, c := range challenges {
		result[i] = s.toResp(c)
	}
	return result, nil
}

func (s *ChallengeService) isContestModifiable(contest *model.Contest) bool {
	return contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusEnded
}

func (s *ChallengeService) toResp(cc *model.ContestChallenge) *dto.ContestChallengeResp {
	return &dto.ContestChallengeResp{
		ID:          cc.ID,
		ContestID:   cc.ContestID,
		ChallengeID: cc.ChallengeID,
		Points:      cc.Points,
		Order:       cc.Order,
		CreatedAt:   cc.CreatedAt,
	}
}
