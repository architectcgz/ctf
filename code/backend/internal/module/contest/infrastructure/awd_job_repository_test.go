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

type awdJobRepositorySourceStub struct {
	findRoundByNumberFn func(context.Context, int64, int) (*model.AWDRound, error)
	findRunningRoundFn  func(context.Context, int64) (*model.AWDRound, error)
}

func (s awdJobRepositorySourceStub) WithinRoundReconcileTransaction(context.Context, func(contestports.AWDRoundReconcileTxRepository) error) error {
	return nil
}

func (s awdJobRepositorySourceStub) WithinRoundServiceWritebackTransaction(context.Context, func(contestports.AWDRoundServiceWritebackTxRepository) error) error {
	return nil
}

func (s awdJobRepositorySourceStub) CreateRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdJobRepositorySourceStub) UpsertRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdJobRepositorySourceStub) ListRoundsByContest(context.Context, int64) ([]model.AWDRound, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) FindRoundByContestAndID(context.Context, int64, int64) (*model.AWDRound, error) {
	return nil, errors.New("unexpected FindRoundByContestAndID call")
}

func (s awdJobRepositorySourceStub) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	if s.findRoundByNumberFn != nil {
		return s.findRoundByNumberFn(ctx, contestID, roundNumber)
	}
	return &model.AWDRound{ID: int64(roundNumber), ContestID: contestID, RoundNumber: roundNumber}, nil
}

func (s awdJobRepositorySourceStub) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	if s.findRunningRoundFn != nil {
		return s.findRunningRoundFn(ctx, contestID)
	}
	return &model.AWDRound{ID: 1, ContestID: contestID, RoundNumber: 1}, nil
}

func (s awdJobRepositorySourceStub) ListSchedulableAWDContests(context.Context, time.Time, time.Time, int) ([]model.Contest, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) FindRegistration(context.Context, int64, int64) (*model.ContestRegistration, error) {
	return nil, errors.New("unexpected FindRegistration call")
}

func (s awdJobRepositorySourceStub) FindContestTeamByMember(context.Context, int64, int64) (*model.Team, error) {
	return nil, errors.New("unexpected FindContestTeamByMember call")
}

func (s awdJobRepositorySourceStub) ListServiceDefinitionsByContest(context.Context, int64) ([]contestports.AWDServiceDefinition, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) ListServiceInstancesByContest(context.Context, int64, []int64) ([]contestports.AWDServiceInstance, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) ListLatestServiceOperationsByContest(context.Context, int64) ([]model.AWDServiceOperation, error) {
	return nil, nil
}

func (s awdJobRepositorySourceStub) HasSystemRecoveryOperationAt(context.Context, int64, int64, int64, time.Time) (bool, error) {
	return false, nil
}

func TestAWDJobRepositoryMapsRoundLookupNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDJobRepository(awdJobRepositorySourceStub{
		findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRoundByNumber(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("FindRoundByNumber() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
	if _, err := repo.FindRunningRound(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("FindRunningRound() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
}

func TestAWDJobRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("round lookup exploded")
	repo := NewAWDJobRepository(awdJobRepositorySourceStub{
		findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
			return nil, expectedErr
		},
		findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindRoundByNumber(context.Background(), 1, 2); !errors.Is(err, expectedErr) {
		t.Fatalf("FindRoundByNumber() error = %v, want %v", err, expectedErr)
	}
	if _, err := repo.FindRunningRound(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("FindRunningRound() error = %v, want %v", err, expectedErr)
	}
}
