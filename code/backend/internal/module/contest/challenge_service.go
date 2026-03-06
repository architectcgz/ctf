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

	if s.isContestRunning(contest) {
		return nil, errcode.ErrContestRunning
	}

	_, err = s.challengeRepo.FindByID(req.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}

	exists, err := s.repo.Exists(contestID, req.ChallengeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.ErrChallengeAlreadyAdded
	}

	cc := &model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: req.ChallengeID,
		Points:      req.Points,
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

	if s.isContestRunning(contest) {
		return errcode.ErrContestRunning
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

	if s.isContestRunning(contest) {
		return errcode.ErrContestRunning
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

func (s *ChallengeService) isContestRunning(contest *model.Contest) bool {
	return contest.Status == "running"
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
