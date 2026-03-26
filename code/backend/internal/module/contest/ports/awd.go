package ports

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
)

type AWDRepository interface {
	WithinTransaction(ctx context.Context, fn func(repo AWDRepository) error) error
	CreateRound(ctx context.Context, round *model.AWDRound) error
	UpsertRound(ctx context.Context, round *model.AWDRound) error
	ListRoundsByContest(ctx context.Context, contestID int64) ([]model.AWDRound, error)
	FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error)
	FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error)
	FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error)
	ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error)
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error)
	ContestHasChallenge(ctx context.Context, contestID, challengeID int64) (bool, error)
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error)
	ListServiceInstancesByContest(ctx context.Context, contestID int64, challengeIDs []int64) ([]AWDServiceInstance, error)
	UpsertServiceCheck(ctx context.Context, roundID, teamID, challengeID int64, serviceStatus, checkResult string, defenseScore int, updatedAt time.Time) (*model.AWDTeamService, error)
	UpsertTeamServices(ctx context.Context, records []model.AWDTeamService) error
	ListServicesByRound(ctx context.Context, roundID int64) ([]model.AWDTeamService, error)
	CountSuccessfulAttacks(ctx context.Context, roundID, attackerTeamID, victimTeamID, challengeID int64) (int64, error)
	CreateAttackLog(ctx context.Context, logRecord *model.AWDAttackLog) error
	ApplyAttackImpactToVictimService(ctx context.Context, roundID, victimTeamID, challengeID int64, scoreGained int, updatedAt time.Time) error
	ListAttackLogsByRound(ctx context.Context, roundID int64) ([]model.AWDAttackLog, error)
	RecalculateContestTeamScores(ctx context.Context, contestID int64) error
	RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error
}

type AWDFlagAssignment struct {
	TeamID      int64
	ChallengeID int64
	Flag        string
}

type AWDContainerFileWriter interface {
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}

type AWDFlagInjector interface {
	InjectRoundFlags(ctx context.Context, contest *model.Contest, round *model.AWDRound, assignments []AWDFlagAssignment) error
}

type AWDServiceInstance struct {
	TeamID      int64
	ChallengeID int64
	AccessURL   string
}

type AWDRoundManager interface {
	RunRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error
	EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error
}
