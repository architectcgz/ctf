package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdCommandRepositorySourceStub struct {
	findRoundByContestAndIDFn             func(context.Context, int64, int64) (*model.AWDRound, error)
	findRoundByNumberFn                   func(context.Context, int64, int) (*model.AWDRound, error)
	findRunningRoundFn                    func(context.Context, int64) (*model.AWDRound, error)
	findRegistrationFn                    func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findContestTeamByMemberFn             func(context.Context, int64, int64) (*model.Team, error)
	findChallengeByIDFn                   func(context.Context, int64) (*model.Challenge, error)
	findContestAWDServiceByContestAndIDFn func(context.Context, int64, int64) (*model.ContestAWDService, error)
	withinAttackLogTransactionFn          func(context.Context, func(contestports.AWDAttackLogTxRepository) error) error
}

func (s awdCommandRepositorySourceStub) WithinServiceCheckTransaction(context.Context, func(contestports.AWDServiceCheckTxRepository) error) error {
	return nil
}

func (s awdCommandRepositorySourceStub) WithinAttackLogTransaction(ctx context.Context, fn func(contestports.AWDAttackLogTxRepository) error) error {
	if s.withinAttackLogTransactionFn != nil {
		return s.withinAttackLogTransactionFn(ctx, fn)
	}
	return fn(awdCommandAttackLogTxRepoStub{})
}

func (s awdCommandRepositorySourceStub) CreateContestAWDService(context.Context, *model.ContestAWDService) error {
	return nil
}

func (s awdCommandRepositorySourceStub) UpdateContestAWDServiceByContestAndID(context.Context, int64, int64, map[string]any) error {
	return nil
}

func (s awdCommandRepositorySourceStub) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	if s.findContestAWDServiceByContestAndIDFn != nil {
		return s.findContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID)
	}
	return &model.ContestAWDService{ID: serviceID, ContestID: contestID}, nil
}

func (s awdCommandRepositorySourceStub) ListContestAWDServicesByContest(context.Context, int64) ([]model.ContestAWDService, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) DeleteContestAWDServiceByContestAndID(context.Context, int64, int64) error {
	return nil
}

func (s awdCommandRepositorySourceStub) CreateRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdCommandRepositorySourceStub) UpsertRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdCommandRepositorySourceStub) ListRoundsByContest(context.Context, int64) ([]model.AWDRound, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	if s.findRoundByContestAndIDFn != nil {
		return s.findRoundByContestAndIDFn(ctx, contestID, roundID)
	}
	return &model.AWDRound{ID: roundID, ContestID: contestID}, nil
}

func (s awdCommandRepositorySourceStub) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	if s.findRoundByNumberFn != nil {
		return s.findRoundByNumberFn(ctx, contestID, roundNumber)
	}
	return &model.AWDRound{ID: int64(roundNumber), ContestID: contestID, RoundNumber: roundNumber}, nil
}

func (s awdCommandRepositorySourceStub) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	if s.findRunningRoundFn != nil {
		return s.findRunningRoundFn(ctx, contestID)
	}
	return &model.AWDRound{ID: 1, ContestID: contestID, RoundNumber: 1}, nil
}

func (s awdCommandRepositorySourceStub) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{}, nil
}

func (s awdCommandRepositorySourceStub) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	if s.findContestTeamByMemberFn != nil {
		return s.findContestTeamByMemberFn(ctx, contestID, userID)
	}
	return &model.Team{}, nil
}

func (s awdCommandRepositorySourceStub) ListChallengesByContest(context.Context, int64) ([]model.Challenge, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	if s.findChallengeByIDFn != nil {
		return s.findChallengeByIDFn(ctx, challengeID)
	}
	return &model.Challenge{ID: challengeID}, nil
}

func (s awdCommandRepositorySourceStub) ListReadinessChallengesByContest(context.Context, int64) ([]contestports.AWDReadinessChallengeRecord, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) UpsertServiceCheck(context.Context, int64, int64, int64, int64, string, string, int, time.Time) (*model.AWDTeamService, error) {
	return &model.AWDTeamService{}, nil
}

func (s awdCommandRepositorySourceStub) UpsertTeamServices(context.Context, []model.AWDTeamService) error {
	return nil
}

func (s awdCommandRepositorySourceStub) ListServicesByRound(context.Context, int64) ([]model.AWDTeamService, error) {
	return nil, nil
}

func (s awdCommandRepositorySourceStub) CountSuccessfulAttacks(context.Context, int64, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (s awdCommandRepositorySourceStub) CreateAttackLog(context.Context, *model.AWDAttackLog) error {
	return nil
}

func (s awdCommandRepositorySourceStub) ApplyAttackImpactToVictimService(context.Context, int64, int64, int64, int64, int, time.Time) error {
	return nil
}

func (s awdCommandRepositorySourceStub) ListAttackLogsByRound(context.Context, int64) ([]model.AWDAttackLog, error) {
	return nil, nil
}

type awdCommandAttackLogTxRepoStub struct {
	createAttackLogFn func(context.Context, *model.AWDAttackLog) error
}

func (s awdCommandAttackLogTxRepoStub) CreateAttackLog(ctx context.Context, record *model.AWDAttackLog) error {
	if s.createAttackLogFn != nil {
		return s.createAttackLogFn(ctx, record)
	}
	return nil
}

func (s awdCommandAttackLogTxRepoStub) ApplyAttackImpactToVictimService(context.Context, int64, int64, int64, int64, int, time.Time) error {
	return nil
}

func (s awdCommandAttackLogTxRepoStub) RecalculateContestTeamScores(context.Context, int64) error {
	return nil
}

type awdRoundManagerSourceStub struct {
	ensureActiveRoundMaterializedFn func(context.Context, *model.Contest, time.Time) error
}

func (s awdRoundManagerSourceStub) RunRoundServiceChecks(context.Context, *model.Contest, *model.AWDRound, string) error {
	return nil
}

func (s awdRoundManagerSourceStub) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	if s.ensureActiveRoundMaterializedFn != nil {
		return s.ensureActiveRoundMaterializedFn(ctx, contest, now)
	}
	return nil
}

func (s awdRoundManagerSourceStub) PreviewServiceCheck(context.Context, contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	return nil, nil
}

func TestAWDCommandRepositoryMapsLookupNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDCommandRepository(awdCommandRepositorySourceStub{
		findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestTeamByMemberFn: func(context.Context, int64, int64) (*model.Team, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findChallengeByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRoundByContestAndID(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("FindRoundByContestAndID() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
	if _, err := repo.FindRoundByNumber(context.Background(), 1, 3); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("FindRoundByNumber() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
	if _, err := repo.FindRunningRound(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("FindRunningRound() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
	if _, err := repo.FindRegistration(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("FindRegistration() error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
	if _, err := repo.FindContestTeamByMember(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		t.Fatalf("FindContestTeamByMember() error = %v, want %v", err, contestports.ErrContestUserTeamNotFound)
	}
	if _, err := repo.FindChallengeByID(context.Background(), 9); !errors.Is(err, contestports.ErrContestAWDChallengeNotFound) {
		t.Fatalf("FindChallengeByID() error = %v, want %v", err, contestports.ErrContestAWDChallengeNotFound)
	}
	if _, err := repo.FindContestAWDServiceByContestAndID(context.Background(), 1, 8); !errors.Is(err, contestports.ErrContestAWDServiceNotFound) {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v, want %v", err, contestports.ErrContestAWDServiceNotFound)
	}
}

func TestAWDCommandRepositoryMapsAttackLogTransactionNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDCommandRepository(awdCommandRepositorySourceStub{
		withinAttackLogTransactionFn: func(ctx context.Context, fn func(contestports.AWDAttackLogTxRepository) error) error {
			return fn(awdCommandAttackLogTxRepoStub{
				createAttackLogFn: func(context.Context, *model.AWDAttackLog) error {
					return gorm.ErrRecordNotFound
				},
			})
		},
	})

	err := repo.WithinAttackLogTransaction(context.Background(), func(txRepo contestports.AWDAttackLogTxRepository) error {
		return txRepo.CreateAttackLog(context.Background(), &model.AWDAttackLog{})
	})
	if !errors.Is(err, contestports.ErrContestAWDAttackLogTransactionNotFound) {
		t.Fatalf("WithinAttackLogTransaction() error = %v, want %v", err, contestports.ErrContestAWDAttackLogTransactionNotFound)
	}
}

func TestAWDCommandRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewAWDCommandRepository(awdCommandRepositorySourceStub{
		findChallengeByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
		withinAttackLogTransactionFn: func(context.Context, func(contestports.AWDAttackLogTxRepository) error) error {
			return expectedErr
		},
	})

	if _, err := repo.FindChallengeByID(context.Background(), 9); !errors.Is(err, expectedErr) {
		t.Fatalf("FindChallengeByID() error = %v, want %v", err, expectedErr)
	}
	if err := repo.WithinAttackLogTransaction(context.Background(), func(contestports.AWDAttackLogTxRepository) error {
		return nil
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("WithinAttackLogTransaction() error = %v, want %v", err, expectedErr)
	}
}

func TestAWDRoundManagerAdapterMapsEnsureActiveRoundMaterializedNotFound(t *testing.T) {
	t.Parallel()

	adapter := NewAWDRoundManagerAdapter(awdRoundManagerSourceStub{
		ensureActiveRoundMaterializedFn: func(context.Context, *model.Contest, time.Time) error {
			return gorm.ErrRecordNotFound
		},
	})

	err := adapter.EnsureActiveRoundMaterialized(context.Background(), &model.Contest{ID: 1}, time.Now().UTC())
	if !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("EnsureActiveRoundMaterialized() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
}
