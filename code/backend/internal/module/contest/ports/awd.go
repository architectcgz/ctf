package ports

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
)

type AWDRepository interface {
	WithinTransaction(ctx context.Context, fn func(repo AWDRepository) error) error
	CreateContestAWDService(ctx context.Context, service *model.ContestAWDService) error
	UpdateContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64, updates map[string]any) error
	FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error)
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error)
	DeleteContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) error
	CreateRound(ctx context.Context, round *model.AWDRound) error
	UpsertRound(ctx context.Context, round *model.AWDRound) error
	ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error)
	FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error)
	FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error)
	FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error)
	ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error)
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]AWDServiceDefinition, error)
	ListReadinessChallengesByContest(ctx context.Context, contestID int64) ([]AWDReadinessChallengeRecord, error)
	ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error)
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error)
	ListServiceInstancesByContest(ctx context.Context, contestID int64, serviceIDs []int64) ([]AWDServiceInstance, error)
	UpsertServiceCheck(ctx context.Context, roundID, teamID, serviceID, awdChallengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*model.AWDTeamService, error)
	UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error
	ListServicesByRound(ctx context.Context, roundID int64) ([]model.AWDTeamService, error)
	CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, serviceID int64) (int64, error)
	CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, serviceID, awdChallengeID int64, scoreGained int, updatedAt time.Time) error
	ListAttackLogsByRound(ctx context.Context, roundID int64) ([]model.AWDAttackLog, error)
	ListTrafficEvents(ctx context.Context, contestID, roundID int64) ([]AWDTrafficEventRecord, error)
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
	RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error
}

type AWDFlagAssignment struct {
	ServiceID      int64
	TeamID         int64
	AWDChallengeID int64
	Flag           string
}

type AWDServiceDefinition struct {
	ServiceID      int64                `gorm:"column:service_id"`
	ServiceName    string               `gorm:"column:service_name"`
	AWDChallengeID int64                `gorm:"column:awd_challenge_id"`
	FlagPrefix     string               `gorm:"column:flag_prefix"`
	CheckerType    model.AWDCheckerType `gorm:"column:awd_checker_type"`
	CheckerConfig  string               `gorm:"column:awd_checker_config"`
	SLAScore       int                  `gorm:"column:awd_sla_score"`
	DefenseScore   int                  `gorm:"column:awd_defense_score"`
}

type AWDReadinessChallengeRecord struct {
	ServiceID         int64                           `gorm:"column:service_id"`
	AWDChallengeID    int64                           `gorm:"column:awd_challenge_id"`
	Title             string                          `gorm:"column:title"`
	CheckerType       model.AWDCheckerType            `gorm:"column:awd_checker_type"`
	CheckerConfig     string                          `gorm:"column:awd_checker_config"`
	ValidationState   model.AWDCheckerValidationState `gorm:"column:awd_checker_validation_state"`
	LastPreviewAt     *time.Time                      `gorm:"column:awd_checker_last_preview_at"`
	LastPreviewResult string                          `gorm:"column:awd_checker_last_preview_result"`
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
	AccessURL      string
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
