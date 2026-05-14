package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdRoundUpdaterRepoStub struct {
	findRunningRoundFn  func(context.Context, int64) (*model.AWDRound, error)
	findRoundByNumberFn func(context.Context, int64, int) (*model.AWDRound, error)
}

func (s awdRoundUpdaterRepoStub) WithinRoundReconcileTransaction(context.Context, func(contestports.AWDRoundReconcileTxRepository) error) error {
	return nil
}

func (s awdRoundUpdaterRepoStub) WithinRoundServiceWritebackTransaction(context.Context, func(contestports.AWDRoundServiceWritebackTxRepository) error) error {
	return nil
}

func (s awdRoundUpdaterRepoStub) CreateRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdRoundUpdaterRepoStub) UpsertRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdRoundUpdaterRepoStub) ListRoundsByContest(context.Context, int64) ([]model.AWDRound, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) FindRoundByContestAndID(context.Context, int64, int64) (*model.AWDRound, error) {
	return nil, errors.New("unexpected FindRoundByContestAndID call")
}

func (s awdRoundUpdaterRepoStub) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	if s.findRoundByNumberFn != nil {
		return s.findRoundByNumberFn(ctx, contestID, roundNumber)
	}
	return &model.AWDRound{ID: int64(roundNumber), ContestID: contestID, RoundNumber: roundNumber}, nil
}

func (s awdRoundUpdaterRepoStub) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	if s.findRunningRoundFn != nil {
		return s.findRunningRoundFn(ctx, contestID)
	}
	return &model.AWDRound{ID: 1, ContestID: contestID, RoundNumber: 1}, nil
}

func (s awdRoundUpdaterRepoStub) ListSchedulableAWDContests(context.Context, time.Time, time.Time, int) ([]model.Contest, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) FindRegistration(context.Context, int64, int64) (*model.ContestRegistration, error) {
	return nil, errors.New("unexpected FindRegistration call")
}

func (s awdRoundUpdaterRepoStub) FindContestTeamByMember(context.Context, int64, int64) (*model.Team, error) {
	return nil, errors.New("unexpected FindContestTeamByMember call")
}

func (s awdRoundUpdaterRepoStub) ListServiceDefinitionsByContest(context.Context, int64) ([]contestports.AWDServiceDefinition, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) ListServiceInstancesByContest(context.Context, int64, []int64) ([]contestports.AWDServiceInstance, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) ListLatestServiceOperationsByContest(context.Context, int64) ([]model.AWDServiceOperation, error) {
	return nil, nil
}

func (s awdRoundUpdaterRepoStub) HasSystemRecoveryOperationAt(context.Context, int64, int64, int64, time.Time) (bool, error) {
	return false, nil
}

type awdRoundStateStoreStub struct {
	isCurrentRoundFn func(context.Context, int64, int) (bool, error)
	loadFlagFn       func(context.Context, int64, int64, int64, int64) (string, bool, error)
}

func (s awdRoundStateStoreStub) AcquireAWDSchedulerLock(context.Context, time.Duration) (contestports.ContestSchedulerLockLease, bool, error) {
	return nil, false, nil
}

func (s awdRoundStateStoreStub) TryAcquireAWDRoundLock(context.Context, int64, int, time.Duration) (bool, error) {
	return false, nil
}

func (s awdRoundStateStoreStub) IsAWDCurrentRound(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if s.isCurrentRoundFn != nil {
		return s.isCurrentRoundFn(ctx, contestID, roundNumber)
	}
	return false, nil
}

func (s awdRoundStateStoreStub) LoadAWDCurrentRoundNumber(context.Context, int64) (int, bool, error) {
	return 0, false, nil
}

func (s awdRoundStateStoreStub) LoadAWDRoundFlag(ctx context.Context, contestID, roundID, teamID, serviceID int64) (string, bool, error) {
	if s.loadFlagFn != nil {
		return s.loadFlagFn(ctx, contestID, roundID, teamID, serviceID)
	}
	return "", false, nil
}

func (s awdRoundStateStoreStub) SyncAWDCurrentRoundState(context.Context, int64, *model.AWDRound, []contestports.AWDFlagAssignment, time.Duration) error {
	return nil
}

func (s awdRoundStateStoreStub) ClearAWDCurrentRoundState(context.Context, int64) error {
	return nil
}

func (s awdRoundStateStoreStub) SetAWDServiceStatus(context.Context, int64, int64, int64, string) error {
	return nil
}

func (s awdRoundStateStoreStub) ReplaceAWDServiceStatus(context.Context, int64, []contestports.AWDServiceStatusEntry) error {
	return nil
}

func (s awdRoundStateStoreStub) ClearAWDServiceStatus(context.Context, int64) error {
	return nil
}

func TestAWDRoundUpdaterShouldSyncLiveServiceStatusCacheTreatsMissingRunningRoundAsStateFallback(t *testing.T) {
	t.Parallel()

	updater := &AWDRoundUpdater{
		repo: awdRoundUpdaterRepoStub{
			findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		stateStore: awdRoundStateStoreStub{
			isCurrentRoundFn: func(context.Context, int64, int) (bool, error) {
				return true, nil
			},
		},
	}

	ok, err := updater.shouldSyncLiveServiceStatusCache(context.Background(), 10, &model.AWDRound{ID: 20, RoundNumber: 3})
	if err != nil {
		t.Fatalf("shouldSyncLiveServiceStatusCache() error = %v", err)
	}
	if !ok {
		t.Fatal("expected fallback to state store current round marker")
	}
}

func TestAWDRoundUpdaterResolveAcceptedRoundFlagsTreatsPreviousRoundSentinelAsCurrentOnly(t *testing.T) {
	t.Parallel()

	startedAt := time.Now().UTC().Add(-30 * time.Second)
	updater := &AWDRoundUpdater{
		repo: awdRoundUpdaterRepoStub{
			findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		cfg:        config.ContestAWDConfig{PreviousRoundGrace: time.Minute},
		flagSecret: "slice39-secret",
	}

	round := &model.AWDRound{ID: 11, ContestID: 10, RoundNumber: 2, StartedAt: &startedAt}
	definition := contestports.AWDServiceDefinition{
		ServiceID:      21,
		AWDChallengeID: 31,
		FlagPrefix:     "awd",
	}

	flags, err := updater.resolveAcceptedRoundFlags(context.Background(), 10, round, 41, definition, startedAt.Add(30*time.Second))
	if err != nil {
		t.Fatalf("resolveAcceptedRoundFlags() error = %v", err)
	}
	if len(flags) != 1 {
		t.Fatalf("expected only current round flag, got %d: %+v", len(flags), flags)
	}
	expected := contestdomain.BuildAWDRoundFlag(10, 2, 41, 31, "slice39-secret", "awd")
	if flags[0] != expected {
		t.Fatalf("unexpected current round flag: got %q want %q", flags[0], expected)
	}
}

func TestAWDRoundUpdaterEnsureActiveRoundMaterializedTreatsNoActiveRoundAsSentinel(t *testing.T) {
	t.Parallel()

	updater := &AWDRoundUpdater{
		cfg: config.ContestAWDConfig{RoundInterval: time.Minute},
	}

	contest := &model.Contest{
		ID:        10,
		StartTime: time.Now().UTC().Add(time.Hour),
		EndTime:   time.Now().UTC().Add(2 * time.Hour),
	}

	err := updater.EnsureActiveRoundMaterialized(context.Background(), contest, time.Now().UTC())
	if !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		t.Fatalf("EnsureActiveRoundMaterialized() error = %v, want %v", err, contestports.ErrContestAWDRoundNotFound)
	}
}
