package ports

import (
	"context"
	"errors"
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

var (
	ErrAWDCheckerPreviewTokenStoreUnavailable = errors.New("awd checker preview token store unavailable")
	ErrContestAWDPreviewChallengeNotFound     = errors.New("contest awd preview challenge not found")
	ErrContestAWDPreviewImageNotFound         = errors.New("contest awd preview image not found")
	ErrContestAWDRoundNotFound                = errors.New("contest awd round not found")
	ErrContestAWDChallengeNotFound            = errors.New("contest awd challenge not found")
	ErrContestAWDServiceNotFound              = errors.New("contest awd service not found")
	ErrContestAWDAttackLogTransactionNotFound = errors.New("contest awd attack log transaction not found")
)

type AWDServiceCheckTxRepository interface {
	UpsertServiceCheck(ctx context.Context, roundID, teamID, serviceID, awdChallengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*contestentity.AWDTeamService, error)
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDServiceCheckTxRunner interface {
	WithinServiceCheckTransaction(ctx context.Context, fn func(repo AWDServiceCheckTxRepository) error) error
}

type AWDAttackLogTxRepository interface {
	CreateAttackLog(ctx context.Context, logRecord *contestentity.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDAttackLogTxRunner interface {
	WithinAttackLogTransaction(ctx context.Context, fn func(repo AWDAttackLogTxRepository) error) error
}

type AWDRoundReconcileTxRepository interface {
	ListRoundsByContest(ctx context.Context, contestID int64) ([]contestentity.AWDRound, error)
	UpsertRound(ctx context.Context, round *contestentity.AWDRound) error
}

type AWDRoundReconcileTxRunner interface {
	WithinRoundReconcileTransaction(ctx context.Context, fn func(repo AWDRoundReconcileTxRepository) error) error
}

type AWDRoundServiceWritebackTxRepository interface {
	UpsertTeamServices(ctx context.Context, records []contestentity.AWDTeamService) error
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
}

type AWDRoundServiceWritebackTxRunner interface {
	WithinRoundServiceWritebackTransaction(ctx context.Context, fn func(repo AWDRoundServiceWritebackTxRepository) error) error
}

type AWDServiceStore interface {
	CreateContestAWDService(ctx context.Context, service *contestentity.ContestAWDService) error
	UpdateContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64, updates map[string]any) error
	FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*contestentity.ContestAWDService, error)
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]contestentity.ContestAWDService, error)
	DeleteContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) error
}

type AWDRoundStore interface {
	CreateRound(ctx context.Context, round *contestentity.AWDRound) error
	UpsertRound(ctx context.Context, round *contestentity.AWDRound) error
	ListRoundsByContest(ctx context.Context, contestID int64) ([]contestentity.AWDRound, error)
	FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*contestentity.AWDRound, error)
	FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*contestentity.AWDRound, error)
	FindRunningRound(ctx context.Context, contestID int64) (*contestentity.AWDRound, error)
}

type AWDContestScheduleQuery interface {
	ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]contestentity.Contest, error)
}

type AWDTeamLookup interface {
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*contestentity.Team, error)
	FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error)
	FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*contestentity.Team, error)
}

type AWDChallengeLookup interface {
	ListChallengesByContest(ctx context.Context, contestID int64) ([]contestentity.Challenge, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*contestentity.Challenge, error)
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

type AWDDefenseWorkspaceSummaryQuery interface {
	ListDefenseWorkspaceSummariesByContestTeam(ctx context.Context, contestID, teamID int64, serviceIDs []int64) ([]AWDDefenseWorkspaceSummaryRecord, error)
}

type AWDServiceOperationQuery interface {
	ListLatestServiceOperationsByContest(ctx context.Context, contestID int64) ([]runtimecontracts.AWDServiceOperation, error)
	HasSystemRecoveryOperationAt(ctx context.Context, contestID, teamID, serviceID int64, checkedAt time.Time) (bool, error)
}

type AWDTeamServiceStore interface {
	UpsertServiceCheck(ctx context.Context, roundID, teamID, serviceID, awdChallengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*contestentity.AWDTeamService, error)
	UpsertTeamServices(ctx context.Context, records []contestentity.AWDTeamService) error
	ListServicesByRound(ctx context.Context, roundID int64) ([]contestentity.AWDTeamService, error)
}

type AWDAttackLogStore interface {
	CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, serviceID int64) (int64, error)
	CreateAttackLog(ctx context.Context, logRecord *contestentity.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error
	ListAttackLogsByRound(ctx context.Context, roundID int64) ([]contestentity.AWDAttackLog, error)
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
	ServiceID        int64
	ServiceName      string
	AWDChallengeID   int64
	FlagPrefix       string
	CheckerType      contestentity.AWDCheckerType
	CheckerConfig    string
	CheckerTokenEnv  string
	CheckerToken     string
	SLAScore         int
	DefenseScore     int
	DefenseWorkspace AWDDefenseWorkspaceSummary
}

type AWDDefenseWorkspaceSummary struct {
	EntryMode         string
	WorkspaceStatus   string
	WorkspaceRevision int64
}

type AWDDefenseWorkspaceSummaryRecord struct {
	ServiceID int64
	Summary   AWDDefenseWorkspaceSummary
}

type AWDReadinessChallengeRecord struct {
	ServiceID          int64
	AWDChallengeID     int64
	Title              string
	CheckerType        contestentity.AWDCheckerType
	CheckerConfig      string
	RuntimeImageID     int64
	RuntimeImageStatus string
	ValidationState    contestentity.AWDCheckerValidationState
	LastPreviewAt      *time.Time
	LastPreviewResult  string
}

type AWDContainerFileWriter interface {
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}

type AWDFlagInjector interface {
	InjectRoundFlags(ctx context.Context, contest *contestentity.Contest, round *contestentity.AWDRound, assignments []AWDFlagAssignment) error
}

type AWDServiceInstance struct {
	InstanceID     int64
	ServiceID      int64
	TeamID         int64
	AWDChallengeID int64
	HostPort       int
	ContainerID    string
	NetworkID      string
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

type AWDCheckerPreviewTokenRecord struct {
	ContestID       int64
	ServiceID       int64
	AWDChallengeID  int64
	CheckerType     contestentity.AWDCheckerType
	CheckerConfig   string
	CheckerTokenEnv string
	Result          contestdomain.AWDCheckerPreviewResult
	CreatedAt       time.Time
}

type AWDCheckerPreviewTokenStore interface {
	StoreAWDCheckerPreviewToken(ctx context.Context, record AWDCheckerPreviewTokenRecord, ttl time.Duration) (string, error)
	LoadAWDCheckerPreviewToken(ctx context.Context, contestID int64, token string) (*AWDCheckerPreviewTokenRecord, bool, error)
	DeleteAWDCheckerPreviewToken(ctx context.Context, contestID int64, token string) error
}

type AWDServicePreviewRequest struct {
	ServiceID       int64
	AWDChallengeID  int64
	CheckerType     contestentity.AWDCheckerType
	CheckerConfig   string
	CheckerTokenEnv string
	CheckerToken    string
	AccessURL       string
	PreviewFlag     string
}

type AWDServicePreviewResult struct {
	ServiceStatus  string
	CheckerType    contestentity.AWDCheckerType
	CheckResult    string
	PreviewContext AWDCheckerPreviewContext
}

type AWDRoundManager interface {
	RunRoundServiceChecks(ctx context.Context, contest *contestentity.Contest, round *contestentity.AWDRound, source string) error
	EnsureActiveRoundMaterialized(ctx context.Context, contest *contestentity.Contest, now time.Time) error
	PreviewServiceCheck(ctx context.Context, req AWDServicePreviewRequest) (*AWDServicePreviewResult, error)
}
