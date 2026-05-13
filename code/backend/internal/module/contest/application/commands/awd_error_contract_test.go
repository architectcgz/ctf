package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type awdCommandRepoStub struct {
	findRoundByContestAndIDFn             func(context.Context, int64, int64) (*model.AWDRound, error)
	findRoundByNumberFn                   func(context.Context, int64, int) (*model.AWDRound, error)
	findRunningRoundFn                    func(context.Context, int64) (*model.AWDRound, error)
	findTeamsByContestFn                  func(context.Context, int64) ([]*model.Team, error)
	findRegistrationFn                    func(context.Context, int64, int64) (*model.ContestRegistration, error)
	findContestTeamByMemberFn             func(context.Context, int64, int64) (*model.Team, error)
	findChallengeByIDFn                   func(context.Context, int64) (*model.Challenge, error)
	findContestAWDServiceByContestAndIDFn func(context.Context, int64, int64) (*model.ContestAWDService, error)
	countSuccessfulAttacksFn              func(context.Context, int64, int64, int64, int64) (int64, error)
	withinAttackLogTransactionFn          func(context.Context, func(contestports.AWDAttackLogTxRepository) error) error
}

func (s awdCommandRepoStub) WithinServiceCheckTransaction(context.Context, func(contestports.AWDServiceCheckTxRepository) error) error {
	return nil
}

func (s awdCommandRepoStub) WithinAttackLogTransaction(ctx context.Context, fn func(contestports.AWDAttackLogTxRepository) error) error {
	if s.withinAttackLogTransactionFn != nil {
		return s.withinAttackLogTransactionFn(ctx, fn)
	}
	return fn(awdAttackLogTxRepoStub{})
}

func (s awdCommandRepoStub) CreateContestAWDService(context.Context, *model.ContestAWDService) error {
	return nil
}

func (s awdCommandRepoStub) UpdateContestAWDServiceByContestAndID(context.Context, int64, int64, map[string]any) error {
	return nil
}

func (s awdCommandRepoStub) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	if s.findContestAWDServiceByContestAndIDFn != nil {
		return s.findContestAWDServiceByContestAndIDFn(ctx, contestID, serviceID)
	}
	return &model.ContestAWDService{ID: serviceID, ContestID: contestID}, nil
}

func (s awdCommandRepoStub) ListContestAWDServicesByContest(context.Context, int64) ([]model.ContestAWDService, error) {
	return nil, nil
}

func (s awdCommandRepoStub) DeleteContestAWDServiceByContestAndID(context.Context, int64, int64) error {
	return nil
}

func (s awdCommandRepoStub) CreateRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdCommandRepoStub) UpsertRound(context.Context, *model.AWDRound) error {
	return nil
}

func (s awdCommandRepoStub) ListRoundsByContest(context.Context, int64) ([]model.AWDRound, error) {
	return nil, nil
}

func (s awdCommandRepoStub) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	if s.findRoundByContestAndIDFn != nil {
		return s.findRoundByContestAndIDFn(ctx, contestID, roundID)
	}
	return &model.AWDRound{ID: roundID, ContestID: contestID, RoundNumber: 1}, nil
}

func (s awdCommandRepoStub) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	if s.findRoundByNumberFn != nil {
		return s.findRoundByNumberFn(ctx, contestID, roundNumber)
	}
	return &model.AWDRound{ID: int64(roundNumber), ContestID: contestID, RoundNumber: roundNumber}, nil
}

func (s awdCommandRepoStub) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	if s.findRunningRoundFn != nil {
		return s.findRunningRoundFn(ctx, contestID)
	}
	return &model.AWDRound{ID: 1, ContestID: contestID, RoundNumber: 1}, nil
}

func (s awdCommandRepoStub) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	if s.findTeamsByContestFn != nil {
		return s.findTeamsByContestFn(ctx, contestID)
	}
	return nil, nil
}

func (s awdCommandRepoStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		Status:    model.ContestRegistrationStatusApproved,
	}, nil
}

func (s awdCommandRepoStub) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	if s.findContestTeamByMemberFn != nil {
		return s.findContestTeamByMemberFn(ctx, contestID, userID)
	}
	return &model.Team{ID: 1, ContestID: contestID}, nil
}

func (s awdCommandRepoStub) ListChallengesByContest(context.Context, int64) ([]model.Challenge, error) {
	return nil, nil
}

func (s awdCommandRepoStub) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	if s.findChallengeByIDFn != nil {
		return s.findChallengeByIDFn(ctx, challengeID)
	}
	return &model.Challenge{ID: challengeID}, nil
}

func (s awdCommandRepoStub) ListReadinessChallengesByContest(context.Context, int64) ([]contestports.AWDReadinessChallengeRecord, error) {
	return nil, nil
}

func (s awdCommandRepoStub) UpsertServiceCheck(context.Context, int64, int64, int64, int64, string, string, int, time.Time) (*model.AWDTeamService, error) {
	return &model.AWDTeamService{}, nil
}

func (s awdCommandRepoStub) UpsertTeamServices(context.Context, []model.AWDTeamService) error {
	return nil
}

func (s awdCommandRepoStub) ListServicesByRound(context.Context, int64) ([]model.AWDTeamService, error) {
	return nil, nil
}

func (s awdCommandRepoStub) CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, serviceID int64) (int64, error) {
	if s.countSuccessfulAttacksFn != nil {
		return s.countSuccessfulAttacksFn(ctx, roundID, attackerTeamID, victimTeamID, serviceID)
	}
	return 0, nil
}

func (s awdCommandRepoStub) CreateAttackLog(context.Context, *model.AWDAttackLog) error {
	return nil
}

func (s awdCommandRepoStub) ApplyAttackImpactToVictimService(context.Context, int64, int64, int64, int64, int, time.Time) error {
	return nil
}

func (s awdCommandRepoStub) ListAttackLogsByRound(context.Context, int64) ([]model.AWDAttackLog, error) {
	return nil, nil
}

type awdAttackLogTxRepoStub struct {
	createAttackLogFn                  func(context.Context, *model.AWDAttackLog) error
	applyAttackImpactToVictimServiceFn func(context.Context, int64, int64, int64, int64, int, time.Time) error
	recalculateContestTeamScoresFn     func(context.Context, int64) error
}

func (s awdAttackLogTxRepoStub) CreateAttackLog(ctx context.Context, record *model.AWDAttackLog) error {
	if s.createAttackLogFn != nil {
		return s.createAttackLogFn(ctx, record)
	}
	return nil
}

func (s awdAttackLogTxRepoStub) ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error {
	if s.applyAttackImpactToVictimServiceFn != nil {
		return s.applyAttackImpactToVictimServiceFn(ctx, roundID, victimTeamID, serviceID, awdChallengeID, scoreGained, updatedAt)
	}
	return nil
}

func (s awdAttackLogTxRepoStub) RecalculateContestTeamScores(ctx context.Context, contestID int64) error {
	if s.recalculateContestTeamScoresFn != nil {
		return s.recalculateContestTeamScoresFn(ctx, contestID)
	}
	return nil
}

type awdContestLookupStub struct {
	findByIDFn func(context.Context, int64) (*model.Contest, error)
}

func (s awdContestLookupStub) FindByID(ctx context.Context, id int64) (*model.Contest, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Contest{ID: id, Mode: model.ContestModeAWD}, nil
}

type awdRoundStateStoreStub struct {
	loadCurrentRoundNumberFn func(context.Context, int64) (int, bool, error)
}

func (s awdRoundStateStoreStub) AcquireAWDSchedulerLock(context.Context, time.Duration) (contestports.ContestSchedulerLockLease, bool, error) {
	return nil, false, nil
}

func (s awdRoundStateStoreStub) TryAcquireAWDRoundLock(context.Context, int64, int, time.Duration) (bool, error) {
	return false, nil
}

func (s awdRoundStateStoreStub) IsAWDCurrentRound(context.Context, int64, int) (bool, error) {
	return false, nil
}

func (s awdRoundStateStoreStub) LoadAWDCurrentRoundNumber(ctx context.Context, contestID int64) (int, bool, error) {
	if s.loadCurrentRoundNumberFn != nil {
		return s.loadCurrentRoundNumberFn(ctx, contestID)
	}
	return 0, false, nil
}

func (s awdRoundStateStoreStub) LoadAWDRoundFlag(context.Context, int64, int64, int64, int64) (string, bool, error) {
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

type awdRoundManagerStub struct {
	ensureActiveRoundMaterializedFn func(context.Context, *model.Contest, time.Time) error
}

func (s awdRoundManagerStub) RunRoundServiceChecks(context.Context, *model.Contest, *model.AWDRound, string) error {
	return nil
}

func (s awdRoundManagerStub) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	if s.ensureActiveRoundMaterializedFn != nil {
		return s.ensureActiveRoundMaterializedFn(ctx, contest, now)
	}
	return nil
}

func (s awdRoundManagerStub) PreviewServiceCheck(context.Context, contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	return nil, nil
}

func TestAWDServiceEnsureAWDRoundTreatsModuleRoundSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		contestRepo: awdContestLookupStub{},
	}

	_, err := service.ensureAWDRound(context.Background(), 11, 22)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAWDServiceResolveUserTeamIDFallsBackFromRegistrationSentinelToMembershipLookup(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
			findContestTeamByMemberFn: func(context.Context, int64, int64) (*model.Team, error) {
				return &model.Team{ID: 91}, nil
			},
		},
	}

	teamID, err := service.resolveUserTeamID(context.Background(), 2001, 11)
	if err != nil {
		t.Fatalf("resolveUserTeamID() error = %v", err)
	}
	if teamID != 91 {
		t.Fatalf("unexpected team id: %d", teamID)
	}
}

func TestAWDServiceResolveUserTeamIDTreatsMembershipSentinelAsNotRegistered(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
				return nil, contestports.ErrContestParticipationRegistrationNotFound
			},
			findContestTeamByMemberFn: func(context.Context, int64, int64) (*model.Team, error) {
				return nil, contestports.ErrContestUserTeamNotFound
			},
		},
	}

	_, err := service.resolveUserTeamID(context.Background(), 2001, 11)
	if err != errcode.ErrNotRegistered {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
}

func TestAWDServiceResolveCurrentRoundFromFallbacksTreatsMissingRoundSentinelAsNotActive(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRunningRoundFn: func(context.Context, int64) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
			findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		stateStore: awdRoundStateStoreStub{
			loadCurrentRoundNumberFn: func(context.Context, int64) (int, bool, error) {
				return 3, true, nil
			},
		},
	}

	_, err := service.resolveCurrentRoundFromFallbacks(context.Background(), 11)
	if err != errcode.ErrAWDRoundNotActive {
		t.Fatalf("expected ErrAWDRoundNotActive, got %v", err)
	}
}

func TestAWDServiceResolveMaterializedActiveRoundTreatsRoundManagerSentinelAsNotActive(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRoundByNumberFn: func(context.Context, int64, int) (*model.AWDRound, error) {
				return nil, contestports.ErrContestAWDRoundNotFound
			},
		},
		roundManager: awdRoundManagerStub{
			ensureActiveRoundMaterializedFn: func(context.Context, *model.Contest, time.Time) error {
				return contestports.ErrContestAWDRoundNotFound
			},
		},
	}

	contest := &model.Contest{ID: 11, Mode: model.ContestModeAWD}
	_, err := service.resolveMaterializedActiveRound(context.Background(), contest, 3, time.Now().UTC())
	if err != errcode.ErrAWDRoundNotActive {
		t.Fatalf("expected ErrAWDRoundNotActive, got %v", err)
	}
}

func TestAWDServiceLoadChallengeTreatsChallengeSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findChallengeByIDFn: func(context.Context, int64) (*model.Challenge, error) {
				return nil, contestports.ErrContestAWDChallengeNotFound
			},
		},
	}

	_, err := service.loadChallenge(context.Background(), 501)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAWDServiceResolveContestRuntimeServiceTreatsServiceSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return nil, contestports.ErrContestAWDServiceNotFound
			},
		},
	}

	_, err := service.resolveContestRuntimeService(context.Background(), 11, 33)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAWDServiceCreateAttackLogTreatsTransactionSentinelAsErrNotFound(t *testing.T) {
	t.Parallel()

	service := &AWDService{
		repo: awdCommandRepoStub{
			findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
				return &model.AWDRound{ID: 31, ContestID: 11, RoundNumber: 1}, nil
			},
			findTeamsByContestFn: func(context.Context, int64) ([]*model.Team, error) {
				return []*model.Team{{ID: 101}, {ID: 102}}, nil
			},
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{ID: 201, ContestID: 11, AWDChallengeID: 301}, nil
			},
			withinAttackLogTransactionFn: func(context.Context, func(contestports.AWDAttackLogTxRepository) error) error {
				return contestports.ErrContestAWDAttackLogTransactionNotFound
			},
		},
		contestRepo: awdContestLookupStub{},
	}

	_, err := service.CreateAttackLog(context.Background(), 11, 31, CreateAttackLogInput{
		AttackerTeamID: 101,
		VictimTeamID:   102,
		ServiceID:      201,
		AttackType:     "exploit",
	})
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAWDServiceCreateAttackLogDoesNotSwallowUnexpectedTransactionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("attack log write exploded")
	service := &AWDService{
		repo: awdCommandRepoStub{
			findRoundByContestAndIDFn: func(context.Context, int64, int64) (*model.AWDRound, error) {
				return &model.AWDRound{ID: 31, ContestID: 11, RoundNumber: 1}, nil
			},
			findTeamsByContestFn: func(context.Context, int64) ([]*model.Team, error) {
				return []*model.Team{{ID: 101}, {ID: 102}}, nil
			},
			findContestAWDServiceByContestAndIDFn: func(context.Context, int64, int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{ID: 201, ContestID: 11, AWDChallengeID: 301}, nil
			},
			withinAttackLogTransactionFn: func(context.Context, func(contestports.AWDAttackLogTxRepository) error) error {
				return expectedErr
			},
		},
		contestRepo: awdContestLookupStub{},
	}

	_, err := service.CreateAttackLog(context.Background(), 11, 31, CreateAttackLogInput{
		AttackerTeamID: 101,
		VictimTeamID:   102,
		ServiceID:      201,
		AttackType:     "exploit",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected original transaction error, got %v", err)
	}
	if errors.Is(err, errcode.ErrNotFound) {
		t.Fatalf("expected non-not-found transaction error to avoid ErrNotFound mapping, got %v", err)
	}
}
