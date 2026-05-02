package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type AWDServiceCheckTxRepository interface {
	UpsertServiceCheck(ctx context.Context, roundID, teamID, serviceID, awdChallengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*model.AWDTeamService, error)
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDServiceCheckTxRunner interface {
	WithinServiceCheckTransaction(ctx context.Context, fn func(repo AWDServiceCheckTxRepository) error) error
}

type AWDAttackLogTxRepository interface {
	CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDAttackLogTxRunner interface {
	WithinAttackLogTransaction(ctx context.Context, fn func(repo AWDAttackLogTxRepository) error) error
}

type AWDRoundReconcileTxRepository interface {
	ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error)
	UpsertRound(ctx context.Context, round *model.AWDRound) error
}

type AWDRoundReconcileTxRunner interface {
	WithinRoundReconcileTransaction(ctx context.Context, fn func(repo AWDRoundReconcileTxRepository) error) error
}

type AWDRoundServiceWritebackTxRepository interface {
	UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDRoundServiceWritebackTxRunner interface {
	WithinRoundServiceWritebackTransaction(ctx context.Context, fn func(repo AWDRoundServiceWritebackTxRepository) error) error
}

type AWDServiceStore interface {
	CreateContestAWDService(ctx context.Context, service *model.ContestAWDService) error
	UpdateContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64, updates map[string]any) error
	FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error)
	DeleteContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) error
}

type AWDRoundStore interface {
	CreateRound(ctx context.Context, round *model.AWDRound) error
	UpsertRound(ctx context.Context, round *model.AWDRound) error
	ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error)
	FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error)
	FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error)
	FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error)
}

type AWDContestScheduleQuery interface {
	ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error)
}

type AWDTeamLookup interface {
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error)
}

type AWDChallengeLookup interface {
	ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error)
}

type AWDServiceDefinitionQuery interface {
	ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]AWDServiceDefinition, error)
}

type AWDReadinessQuery interface {
	ListReadinessChallengesByContest(ctx context.Context, contestID int64) ([]AWDReadinessChallengeRecord, error)
}

type AWDServiceInstanceQuery interface {
	ListServiceInstancesByContest(ctx context.Context, contestID int64, serviceIDs []int64) ([]AWDServiceInstance, error)
}

type AWDServiceOperationQuery interface {
	ListLatestServiceOperationsByContest(ctx context.Context, contestID int64) ([]model.AWDServiceOperation, error)
	HasSystemRecoveryOperationAt(ctx context.Context, contestID, teamID, serviceID int64, checkedAt time.Time) (bool, error)
}

type AWDTeamServiceStore interface {
	UpsertServiceCheck(ctx context.Context, roundID, teamID, serviceID, awdChallengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*model.AWDTeamService, error)
	UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error
	ListServicesByRound(ctx context.Context, roundID int64) ([]model.AWDTeamService, error)
}

type AWDAttackLogStore interface {
	CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, serviceID int64) (int64, error)
	CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error
	ListAttackLogsByRound(ctx context.Context, roundID int64) ([]model.AWDAttackLog, error)
}

type AWDTrafficEventQuery interface {
	ListTrafficEvents(ctx context.Context, contestID, roundID int64) ([]AWDTrafficEventRecord, error)
}

type AWDScoreStore interface {
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type ScoreboardCacheWriter interface {
	RebuildContestScoreboard(ctx context.Context, contestID int64) error
}

type AWDFlagAssignment struct {
	ServiceID      int64
	TeamID         int64
	AWDChallengeID int64
	Flag           string
}

type AWDServiceDefinition struct {
	ServiceID      int64
	ServiceName    string
	AWDChallengeID int64
	FlagPrefix     string
	CheckerType    model.AWDCheckerType
	CheckerConfig  string
	SLAScore       int
	DefenseScore   int
}

type AWDReadinessChallengeRecord struct {
	ServiceID         int64
	AWDChallengeID    int64
	Title             string
	CheckerType       model.AWDCheckerType
	CheckerConfig     string
	ValidationState   model.AWDCheckerValidationState
	LastPreviewAt     *time.Time
	LastPreviewResult string
}

type AWDContainerFileWriter interface {
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}

type AWDFlagInjector interface {
	InjectRoundFlags(ctx context.Context, contest *model.Contest, round *model.AWDRound, assignments []AWDFlagAssignment) error
}

type AWDServiceInstance struct {
	InstanceID     int64
	ServiceID      int64
	TeamID         int64
	AWDChallengeID int64
	Status         string
	AccessURL      string
	RuntimeDetails string
}

type AWDTrafficEventRecord struct {
	ID                int64
	ContestID         int64
	RoundID           int64
	AttackerTeamID    int64
	AttackerTeamName  string
	VictimTeamID      int64
	VictimTeamName    string
	ServiceID         int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	Method            string
	Path              string
	StatusCode        int
	Source            string
	OccurredAt        time.Time
}

type AWDCheckerPreviewContext struct {
	ServiceID      int64
	AccessURL      string
	PreviewFlag    string
	RoundNumber    int
	TeamID         int64
	AWDChallengeID int64
}

type AWDServicePreviewRequest struct {
	ServiceID      int64
	AWDChallengeID int64
	CheckerType    model.AWDCheckerType
	CheckerConfig  string
	AccessURL      string
	PreviewFlag    string
}

type AWDServicePreviewResult struct {
	ServiceStatus  string
	CheckerType    model.AWDCheckerType
	CheckResult    string
	PreviewContext AWDCheckerPreviewContext
}

type AWDRoundManager interface {
	RunRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error
	EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error
	PreviewServiceCheck(ctx context.Context, req AWDServicePreviewRequest) (*AWDServicePreviewResult, error)
}
