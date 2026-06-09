package commands

import (
	"context"
	"errors"
	"testing"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type contestChallengeErrorContestLookupStub struct {
	findByIDFn func(context.Context, int64) (*contestentity.Contest, error)
}

func (s contestChallengeErrorContestLookupStub) FindByID(ctx context.Context, id int64) (*contestentity.Contest, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &contestentity.Contest{ID: id, Mode: contestentity.ContestModeAWD, Status: contestentity.ContestStatusDraft}, nil
}

func (s contestChallengeErrorContestLookupStub) List(context.Context, *string, int, int) ([]*contestentity.Contest, int64, error) {
	return nil, 0, nil
}

type contestChallengeCommandRepoStub struct{}

func (contestChallengeCommandRepoStub) AddChallenge(context.Context, *contestentity.ContestChallenge) error {
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
	findByIDFn func(context.Context, int64) (*contestentity.Challenge, error)
}

func (s contestChallengeLookupStub) FindByID(ctx context.Context, id int64) (*contestentity.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &contestentity.Challenge{ID: id, Status: contestentity.ChallengeStatusPublished}, nil
}

func (s contestChallengeLookupStub) BatchGetSolvedStatus(context.Context, int64, []int64) (map[int64]bool, error) {
	return map[int64]bool{}, nil
}

func (s contestChallengeLookupStub) BatchGetSolvedCount(context.Context, []int64) (map[int64]int64, error) {
	return map[int64]int64{}, nil
}

type contestAWDServiceStoreStub struct {
	createContestAWDServiceFn               func(context.Context, *contestentity.ContestAWDService) error
	updateContestAWDServiceByContestAndIDFn func(context.Context, int64, int64, map[string]any) error
	findContestAWDServiceByContestAndIDFn   func(context.Context, int64, int64) (*contestentity.ContestAWDService, error)
	listContestAWDServicesByContestFn       func(context.Context, int64) ([]contestentity.ContestAWDService, error)
	deleteContestAWDServiceByContestAndIDFn func(context.Context, int64, int64) error
}

func (s contestAWDServiceStoreStub) CreateContestAWDService(ctx context.Context, service *contestentity.ContestAWDService) error {
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

func (s contestAWDServiceStoreStub) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*contestentity.ContestAWDService, error) {
	if s.findContestAWDServiceByContestAndIDFn != nil {
		return s.findContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID)
	}
	return &contestentity.ContestAWDService{ID: serviceID, ContestID: contestID}, nil
}

func (s contestAWDServiceStoreStub) ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]contestentity.ContestAWDService, error) {
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
	findByIDFn func(context.Context, int64) (*challengecontracts.AWDChallenge, error)
}

func (s contestAWDChallengeLookupStub) FindAWDChallengeByID(ctx context.Context, id int64) (*challengecontracts.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &challengecontracts.AWDChallenge{
		ID:          id,
		Name:        "awd challenge",
		CheckerType: challengecontracts.AWDCheckerTypeHTTPStandard,
	}, nil
}

func (s contestAWDChallengeLookupStub) ListAWDChallenges(context.Context, *challengecontracts.AWDChallengeQuery) ([]*challengecontracts.AWDChallenge, int64, error) {
	return nil, 0, nil
}

type contestChallengeRelationStub struct {
	existsFn          func(context.Context, int64, int64) (bool, error)
	addChallengeFn    func(context.Context, *contestentity.ContestChallenge) error
	removeChallengeFn func(context.Context, int64, int64) error
	updateChallengeFn func(context.Context, int64, int64, map[string]any) error
}

func (s contestChallengeRelationStub) AddChallenge(ctx context.Context, cc *contestentity.ContestChallenge) error {
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

var _ contestports.ContestChallengeCatalog = contestChallengeLookupStub{}

func TestChallengeServiceAddChallengeToContestTreatsChallengeSentinelAsErrChallengeNotFound(t *testing.T) {
	t.Parallel()

	service := NewChallengeService(
		contestChallengeCommandRepoStub{},
		contestChallengeLookupStub{
			findByIDFn: func(context.Context, int64) (*contestentity.Challenge, error) {
				return nil, contestports.ErrContestChallengeEntityNotFound
			},
		},
		contestChallengeErrorContestLookupStub{},
		nil,
	)

	_, err := service.AddChallengeToContest(context.Background(), 10, AddContestChallengeInput{ChallengeID: 501})
	if err != challengecontracts.ErrChallengeNotFound {
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
			findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDChallengeNotFound
			},
		},
		nil,
	)

	_, err := service.CreateContestAWDService(context.Background(), 10, CreateContestAWDServiceInput{
		AWDChallengeID: 501,
		Points:         100,
	})
	if err != apperror.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceUpdateTreatsStoredServiceSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*contestentity.ContestAWDService, error) {
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
	if err != apperror.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceUpdateTreatsNewAWDChallengeSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*contestentity.ContestAWDService, error) {
				return &contestentity.ContestAWDService{
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
			findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
				return nil, contestports.ErrContestAWDChallengeNotFound
			},
		},
		nil,
	)

	newChallengeID := int64(30)
	err := service.UpdateContestAWDService(context.Background(), 10, 20, UpdateContestAWDServiceInput{
		AWDChallengeID: &newChallengeID,
	})
	if err != apperror.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContestAWDServiceServiceDeleteTreatsStoredServiceSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestAWDServiceService(
		contestAWDServiceStoreStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*contestentity.ContestAWDService, error) {
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
	if err != apperror.ErrNotFound {
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
			findByIDFn: func(context.Context, int64) (*contestentity.Challenge, error) {
				return nil, contestports.ErrContestChallengeEntityNotFound
			},
		},
		contestAWDChallengeLookupStub{},
		nil,
	)

	err := service.syncContestChallengeRelation(context.Background(), &contestentity.Contest{ID: 10}, 20, 1, true)
	if err != challengecontracts.ErrChallengeNotFound {
		t.Fatalf("expected ErrChallengeNotFound, got %v", err)
	}
}
