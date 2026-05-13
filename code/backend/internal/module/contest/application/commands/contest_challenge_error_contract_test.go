package commands

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type contestChallengeErrorContestLookupStub struct {
	findByIDFn func(context.Context, int64) (*model.Contest, error)
}

func (s contestChallengeErrorContestLookupStub) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Contest{ID: id, Mode: model.ContestModeAWD, Status: model.ContestStatusDraft}, nil
}

func (s contestChallengeErrorContestLookupStub) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}

type contestChallengeCommandRepoStub struct{}

func (contestChallengeCommandRepoStub) AddChallenge(context.Context, *model.ContestChallenge) error {
	return errors.New("unexpected AddChallenge call")
}

func (contestChallengeCommandRepoStub) RemoveChallenge(context.Context, int64, int64) error {
	return errors.New("unexpected RemoveChallenge call")
}

func (contestChallengeCommandRepoStub) UpdateChallenge(context.Context, int64, int64, map[string]any) error {
	return errors.New("unexpected UpdateChallenge call")
}

func (contestChallengeCommandRepoStub) Exists(context.Context, int64, int64) (bool, error) {
	return false, errors.New("unexpected Exists call")
}

func (contestChallengeCommandRepoStub) HasSubmissions(context.Context, int64, int64) (bool, error) {
	return false, errors.New("unexpected HasSubmissions call")
}

type contestChallengeLookupStub struct {
	findByIDFn func(context.Context, int64) (*model.Challenge, error)
}

func (s contestChallengeLookupStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
}

func (s contestChallengeLookupStub) BatchGetSolvedStatus(context.Context, int64, []int64) (map[int64]bool, error) {
	return map[int64]bool{}, nil
}

func (s contestChallengeLookupStub) BatchGetSolvedCount(context.Context, []int64) (map[int64]int64, error) {
	return map[int64]int64{}, nil
}

type contestAWDServiceStoreStub struct {
	createContestAWDServiceFn               func(context.Context, *model.ContestAWDService) error
	updateContestAWDServiceByContestAndIDFn func(context.Context, int64, int64, map[string]any) error
	findContestAWDServiceByContestAndIDFn   func(context.Context, int64, int64) (*model.ContestAWDService, error)
	listContestAWDServicesByContestFn       func(context.Context, int64) ([]model.ContestAWDService, error)
	deleteContestAWDServiceByContestAndIDFn func(context.Context, int64, int64) error
}

func (s contestAWDServiceStoreStub) CreateContestAWDService(ctx context.Context, service *model.ContestAWDService) error {
	if s.createContestAWDServiceFn != nil {
		return s.createContestAWDServiceFn(ctx, service)
	}
	return nil
}

func (s contestAWDServiceStoreStub) UpdateContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64, updates map[string]any) error {
	if s.updateContestAWDServiceByContestAndIDFn != nil {
		return s.updateContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID, updates)
	}
	return nil
}

func (s contestAWDServiceStoreStub) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	if s.findContestAWDServiceByContestAndIDFn != nil {
		return s.findContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID)
	}
	return &model.ContestAWDService{ID: serviceID, ContestID: contestID}, nil
}

func (s contestAWDServiceStoreStub) ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error) {
	if s.listContestAWDServicesByContestFn != nil {
		return s.listContestAWDServicesByContestFn(ctx, contestID)
	}
	return nil, nil
}

func (s contestAWDServiceStoreStub) DeleteContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) error {
	if s.deleteContestAWDServiceByContestAndIDFn != nil {
		return s.deleteContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID)
	}
	return nil
}

type contestAWDChallengeLookupStub struct {
	findByIDFn func(context.Context, int64) (*model.AWDChallenge, error)
}

func (s contestAWDChallengeLookupStub) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.AWDChallenge{
		ID:          id,
		Name:        "awd challenge",
		CheckerType: model.AWDCheckerTypeHTTPStandard,
	}, nil
}

func (s contestAWDChallengeLookupStub) ListAWDChallenges(context.Context, *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return nil, 0, nil
}

type contestChallengeRelationStub struct {
	existsFn          func(context.Context, int64, int64) (bool, error)
	addChallengeFn    func(context.Context, *model.ContestChallenge) error
	removeChallengeFn func(context.Context, int64, int64) error
	updateChallengeFn func(context.Context, int64, int64, map[string]any) error
}

func (s contestChallengeRelationStub) AddChallenge(ctx context.Context, cc *model.ContestChallenge) error {
	if s.addChallengeFn != nil {
		return s.addChallengeFn(ctx, cc)
	}
	return nil
}

func (s contestChallengeRelationStub) RemoveChallenge(ctx context.Context, contestID, challengeID int64) error {
	if s.removeChallengeFn != nil {
		return s.removeChallengeFn(ctx, contestID, challengeID)
	}
	return nil
}

func (s contestChallengeRelationStub) UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error {
	if s.updateChallengeFn != nil {
		return s.updateChallengeFn(ctx, contestID, challengeID, updates)
	}
	return nil
}

func (s contestChallengeRelationStub) Exists(ctx context.Context, contestID, challengeID int64) (bool, error) {
	if s.existsFn != nil {
		return s.existsFn(ctx, contestID, challengeID)
	}
	return false, nil
}

func (contestChallengeRelationStub) HasSubmissions(context.Context, int64, int64) (bool, error) {
	return false, nil
}

var _ challengecontracts.ContestChallengeContract = contestChallengeLookupStub{}

func TestChallengeServiceAddChallengeToContestTreatsChallengeSentinelAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		contestChallengeCommandRepoStub{},
		contestChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
				return nil, contestports.ErrContestChallengeEntityNotFound
			},
		},
		contestChallengeErrorContestLookupStub{},
		nil,
	)

	_, err := service.AddChallengeToContest(context.Background(), 10, AddContestChallengeInput{ChallengeID: 501})
	if err != errcode.ErrChallengeNotFound {
		t.Fatalf("expected ErrChallengeNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceCreateTreatsAWDChallengeSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{},
		contestChallengeErrorContestLookupStub{},
		nil,
		contestChallengeLookupStub{},
		contestAWDChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDChallengeNotFound
			},
		},
		nil,
	)

	_, err := service.CreateContestAWDService(context.Background(), 10, CreateContestAWDServiceInput{
		AWDChallengeID: 501,
		Points:         100,
	})
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceUpdateTreatsStoredServiceSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return nil, contestports.ErrContestAWDServiceNotFound
			},
		},
		contestChallengeErrorContestLookupStub{},
		nil,
		contestChallengeLookupStub{},
		contestAWDChallengeLookupStub{},
		nil,
	)

	err := service.UpdateContestAWDService(context.Background(), 10, 20, UpdateContestAWDServiceInput{})
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceUpdateTreatsNewAWDChallengeSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:            20,
					ContestID:     10,
					DisplayName:   "stored",
					Order:         1,
					IsVisible:     true,
					ScoreConfig:   `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
					RuntimeConfig: `{"checker_type":"http_standard","checker_config":{"path":"/health"}}`,
				}, nil
			},
		},
		contestChallengeErrorContestLookupStub{},
		nil,
		contestChallengeLookupStub{},
		contestAWDChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDChallengeNotFound
			},
		},
		nil,
	)

	newChallengeID := int64(30)
	err := service.UpdateContestAWDService(context.Background(), 10, 20, UpdateContestAWDServiceInput{
		AWDChallengeID: &newChallengeID,
	})
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceDeleteTreatsStoredServiceSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return nil, contestports.ErrContestAWDServiceNotFound
			},
		},
		contestChallengeErrorContestLookupStub{},
		nil,
		contestChallengeLookupStub{},
		contestAWDChallengeLookupStub{},
		nil,
	)

	err := service.DeleteContestAWDService(context.Background(), 10, 20)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceSyncContestChallengeRelationTreatsChallengeSentinelAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{},
		contestChallengeErrorContestLookupStub{},
		contestChallengeRelationStub{},
		contestChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
				return nil, contestports.ErrContestChallengeEntityNotFound
			},
		},
		contestAWDChallengeLookupStub{},
		nil,
	)

	err := service.syncContestChallengeRelation(context.Background(), &model.Contest{ID: 10}, 20, 1, true)
	if err != errcode.ErrChallengeNotFound {
		t.Fatalf("expected ErrChallengeNotFound, got %v", err)
	}
}
