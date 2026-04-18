package jobs_test

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/module/contest/testsupport"
)

const (
	awdCheckSourceScheduler      = testsupport.AWDCheckSourceScheduler
	awdCheckSourceManualCurrent  = testsupport.AWDCheckSourceManualCurrent
	awdCheckSourceManualSelected = testsupport.AWDCheckSourceManualSelected
	awdCheckSourceManualService  = testsupport.AWDCheckSourceManualService
)

var (
	newAWDTestDB                             = testsupport.SetupAWDTestDB
	createAWDContestFixture                  = testsupport.CreateAWDContestFixture
	createAWDRoundFixture                    = testsupport.CreateAWDRoundFixture
	createAWDRoundFixtureWithWindow          = testsupport.CreateAWDRoundFixtureWithWindow
	createAWDChallengeFixture                = testsupport.CreateAWDChallengeFixture
	createAWDContestChallengeFixture         = testsupport.CreateAWDContestChallengeFixture
	createAWDTeamFixture                     = testsupport.CreateAWDTeamFixture
	createAWDTeamMemberFixture               = testsupport.CreateAWDTeamMemberFixture
	createContestRegistrationForExistingTeam = testsupport.CreateContestRegistrationForExistingTeam
	defaultAWDContestServiceID               = testsupport.DefaultAWDContestServiceID
	assertTeamTotalScore                     = testsupport.AssertTeamTotalScore
	assertContestRedisScore                  = testsupport.AssertContestRedisScore
	assertContestRedisScoreMissing           = testsupport.AssertContestRedisScoreMissing
	assertAWDServiceStatusCache              = testsupport.AssertAWDServiceStatusCache
	assertAWDServiceStatusCacheMissing       = testsupport.AssertAWDServiceStatusCacheMissing
)

func buildAWDRoundFlag(contestID int64, roundNumber int, teamID, challengeID int64, secret, prefix string) string {
	return contestdomain.BuildAWDRoundFlag(contestID, roundNumber, teamID, challengeID, secret, prefix)
}

func newAWDRoundUpdaterForTest(db *gorm.DB, redisClient *redis.Client, cfg config.ContestAWDConfig, flagSecret string, injector contestports.AWDFlagInjector, log *zap.Logger) *contestjobs.AWDRoundUpdater {
	return contestjobs.NewAWDRoundUpdater(contestinfra.NewAWDRepository(db), redisClient, cfg, flagSecret, injector, log)
}
