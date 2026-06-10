package commands_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	rediskeys "ctf-platform/internal/module/contest/infrastructure/cachekeys"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/module/contest/testsupport"
	instanceentity "ctf-platform/internal/module/instance/entity"
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
	syncAWDContestServiceFixture             = testsupport.SyncAWDContestServiceFixture
	syncAWDContestServiceReadinessFixture    = testsupport.SyncAWDContestServiceReadinessFixture
	defaultAWDContestServiceID               = testsupport.DefaultAWDContestServiceID
	assertTeamTotalScore                     = testsupport.AssertTeamTotalScore
	assertContestRedisScore                  = testsupport.AssertContestRedisScore
	assertContestRedisScoreMissing           = testsupport.AssertContestRedisScoreMissing
	assertAWDServiceStatusCache              = testsupport.AssertAWDServiceStatusCache
	assertAWDServiceStatusCacheMissing       = testsupport.AssertAWDServiceStatusCacheMissing
)

type awdServiceForTest struct {
	commands *contestcmd.AWDService
	queries  *contestqry.AWDService
}

type fakeContestPreviewRuntimeProbe struct {
	createContainerCalled bool
	cleanupCalled         bool

	lastImageName string
	lastEnv       map[string]string

	containerAccessURL string
	containerDetails   runtimecontracts.InstanceRuntimeDetails
	containerErr       error
	cleanupErr         error
}

func (f *fakeContestPreviewRuntimeProbe) CreateTopology(_ context.Context, _ *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	return nil, errors.New("unexpected CreateTopology call")
}

func (f *fakeContestPreviewRuntimeProbe) CreateContainer(_ context.Context, imageName string, env map[string]string) (string, runtimecontracts.InstanceRuntimeDetails, error) {
	f.createContainerCalled = true
	f.lastImageName = imageName
	f.lastEnv = env
	if f.containerErr != nil {
		return "", runtimecontracts.InstanceRuntimeDetails{}, f.containerErr
	}
	return f.containerAccessURL, f.containerDetails, nil
}

func (f *fakeContestPreviewRuntimeProbe) CleanupRuntimeDetails(_ context.Context, details runtimecontracts.InstanceRuntimeDetails) error {
	f.cleanupCalled = true
	if f.cleanupErr != nil {
		return f.cleanupErr
	}
	if !reflect.DeepEqual(details, f.containerDetails) {
		return errors.New("unexpected runtime details")
	}
	return nil
}

type fakeAWDPreviewRoundManager struct {
	previewCalls     int
	previewResponses []*contestports.AWDServicePreviewResult
	previewErrors    []error
	previewRequests  []contestports.AWDServicePreviewRequest
}

type fakeAWDFlagInjector struct {
	err         error
	callCount   int
	assignments []contestports.AWDFlagAssignment
}

func (f *fakeAWDFlagInjector) InjectRoundFlags(_ context.Context, _ *contestentity.Contest, _ *contestentity.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	f.callCount++
	for _, item := range assignments {
		f.assignments = append(f.assignments, item)
	}
	return f.err
}

type racingAWDRoundStateStore struct {
	contestports.AWDRoundStateStore
	replaceIfMatchFn func(context.Context, int64, int64, int64, int64, string, string, time.Duration) (bool, error)
}

func (s racingAWDRoundStateStore) ReplaceAWDRoundFlagIfMatch(
	ctx context.Context,
	contestID, roundID, teamID, serviceID int64,
	expectedFlag, nextFlag string,
	ttl time.Duration,
) (bool, error) {
	if s.replaceIfMatchFn != nil {
		return s.replaceIfMatchFn(ctx, contestID, roundID, teamID, serviceID, expectedFlag, nextFlag, ttl)
	}
	return false, nil
}

func (f *fakeAWDPreviewRoundManager) RunRoundServiceChecks(_ context.Context, _ *contestentity.Contest, _ *contestentity.AWDRound, _ string) error {
	return errors.New("unexpected RunRoundServiceChecks call")
}

func (f *fakeAWDPreviewRoundManager) EnsureActiveRoundMaterialized(_ context.Context, _ *contestentity.Contest, _ time.Time) error {
	return errors.New("unexpected EnsureActiveRoundMaterialized call")
}

func (f *fakeAWDPreviewRoundManager) PreviewServiceCheck(_ context.Context, req contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	f.previewCalls++
	f.previewRequests = append(f.previewRequests, req)

	index := f.previewCalls - 1
	if index < len(f.previewErrors) && f.previewErrors[index] != nil {
		return nil, f.previewErrors[index]
	}
	if index >= len(f.previewResponses) {
		return nil, errors.New("unexpected PreviewServiceCheck call")
	}
	return f.previewResponses[index], nil
}

func newAWDRoundUpdaterForTest(db *gorm.DB, redisClient *redis.Client, cfg config.ContestAWDConfig, flagSecret string, injector contestports.AWDFlagInjector, log *zap.Logger) *contestjobs.AWDRoundUpdater {
	updater := contestjobs.NewAWDRoundUpdater(
		contestinfra.NewAWDRepository(db),
		contestinfra.NewAWDRoundStateStore(redisClient),
		cfg,
		flagSecret,
		injector,
		log,
		contestinfra.NewScoreboardCache(db, redisClient),
	)
	updater.SetHTTPRuntime(contestinfra.NewAWDHTTPRuntimeAdapter(nil, cfg.CheckerTimeout))
	return updater
}

func newAWDPreviewRuntimeLookupsForTest(db *gorm.DB) (challengecontracts.ImageStore, challengeports.AWDChallengeQueryRepository) {
	return challengeinfra.NewImageRepository(db),
		contestinfra.NewAWDPreviewRuntimeChallengeRepository(challengeinfra.NewRepository(db))
}

func newAWDCommandRepositoryForTest(db *gorm.DB) *contestinfra.AWDCommandRepository {
	return contestinfra.NewAWDCommandRepository(contestinfra.NewAWDRepository(db))
}

func newAWDQueryRepositoryForTest(db *gorm.DB) *contestinfra.AWDQueryRepository {
	return contestinfra.NewAWDQueryRepository(contestinfra.NewAWDRepository(db))
}

func newAWDCommandRoundManagerForTest(db *gorm.DB, redisClient *redis.Client, cfg config.ContestAWDConfig, flagSecret string, injector contestports.AWDFlagInjector, log *zap.Logger) contestports.AWDRoundManager {
	return contestinfra.NewAWDRoundManagerAdapter(newAWDRoundUpdaterForTest(db, redisClient, cfg, flagSecret, injector, log))
}

func newAWDServiceForTest(db *gorm.DB, redisClient *redis.Client, flagSecret string, cfg config.ContestAWDConfig) *awdServiceForTest {
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	return &awdServiceForTest{
		commands: contestcmd.NewAWDService(
			awdRepo,
			contestRepo,
			stateStore,
			previewTokenStore,
			flagSecret,
			cfg,
			zap.NewNop(),
			newAWDCommandRoundManagerForTest(db, redisClient, cfg, flagSecret, nil, zap.NewNop()),
			imageRepo,
			awdChallengeRepo,
			nil,
			contestinfra.NewScoreboardCache(db, redisClient),
		),
		queries: contestqry.NewAWDService(newAWDQueryRepositoryForTest(db), contestRepo),
	}
}

func newAWDCommandServiceWithStateStoreForTest(
	db *gorm.DB,
	redisClient *redis.Client,
	flagSecret string,
	cfg config.ContestAWDConfig,
	stateStore contestports.AWDRoundStateStore,
) *contestcmd.AWDService {
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	if stateStore == nil {
		stateStore = contestinfra.NewAWDRoundStateStore(redisClient)
	}
	return contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		flagSecret,
		cfg,
		zap.NewNop(),
		newAWDCommandRoundManagerForTest(db, redisClient, cfg, flagSecret, nil, zap.NewNop()),
		imageRepo,
		awdChallengeRepo,
		nil,
		contestinfra.NewScoreboardCache(db, redisClient),
	)
}

func (s *awdServiceForTest) CreateRound(ctx context.Context, contestID int64, req contestcmd.CreateAWDRoundInput) (*contestcmd.AWDRoundResp, error) {
	return s.commands.CreateRound(ctx, contestID, req)
}

func (s *awdServiceForTest) ListRounds(ctx context.Context, contestID int64) ([]contestqry.AWDRoundResult, error) {
	return s.queries.ListRounds(ctx, contestID)
}

func (s *awdServiceForTest) RunCurrentRoundChecks(ctx context.Context, contestID int64, req contestcmd.RunCurrentRoundChecksInput) (*contestcmd.AWDCheckerRunResp, error) {
	return s.commands.RunCurrentRoundChecks(ctx, contestID, req)
}

func (s *awdServiceForTest) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*contestcmd.AWDCheckerRunResp, error) {
	return s.commands.RunRoundChecks(ctx, contestID, roundID)
}

func awdServiceIDPtr(contestID, challengeID int64) *int64 {
	id := defaultAWDContestServiceID(contestID, challengeID)
	return &id
}

func (s *awdServiceForTest) UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req contestcmd.UpsertServiceCheckInput) (*contestcmd.AWDTeamServiceResp, error) {
	return s.commands.UpsertServiceCheck(ctx, contestID, roundID, req)
}

func (s *awdServiceForTest) CreateAttackLog(ctx context.Context, contestID, roundID int64, req contestcmd.CreateAttackLogInput) (*contestcmd.AWDAttackLogResp, error) {
	return s.commands.CreateAttackLog(ctx, contestID, roundID, req)
}

func (s *awdServiceForTest) SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req contestcmd.SubmitAttackInput) (*contestcmd.AWDAttackLogResp, error) {
	return s.commands.SubmitAttack(ctx, userID, contestID, serviceID, req)
}

func (s *awdServiceForTest) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*contestqry.AWDRoundSummaryResult, error) {
	return s.queries.GetRoundSummary(ctx, contestID, roundID)
}

func (s *awdServiceForTest) GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*contestqry.AWDTrafficSummaryResult, error) {
	return s.queries.GetTrafficSummary(ctx, contestID, roundID)
}

func (s *awdServiceForTest) ListTrafficEvents(ctx context.Context, contestID, roundID int64, req contestqry.ListAWDTrafficEventsInput) (*contestqry.AWDTrafficEventPageResult, error) {
	return s.queries.ListTrafficEvents(ctx, contestID, roundID, req)
}

func TestAWDServiceCreateRoundAndListRounds(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 1, now)
	createAWDChallengeFixture(t, db, 101, now)
	createAWDContestChallengeFixture(t, db, 1, 101, now)
	syncAWDContestServiceFixture(t, db, 1, 101, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 0, 0, now)
	syncAWDContestServiceReadinessFixture(t, db, 1, 101, contestentity.AWDCheckerValidationStatePassed, nil, "")

	round, err := service.CreateRound(context.Background(), 1, contestcmd.CreateAWDRoundInput{
		RoundNumber:  1,
		AttackScore:  intPtr(80),
		DefenseScore: intPtr(3),
	})
	if err != nil {
		t.Fatalf("CreateRound() error = %v", err)
	}
	if round.AttackScore != 80 || round.DefenseScore != 3 {
		t.Fatalf("unexpected round: %+v", round)
	}

	rounds, err := service.ListRounds(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListRounds() error = %v", err)
	}
	if len(rounds) != 1 || rounds[0].RoundNumber != 1 {
		t.Fatalf("unexpected rounds: %+v", rounds)
	}
}

func TestAWDServiceCreateRoundAppliesDefaultScoreContract(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 71, now)
	createAWDChallengeFixture(t, db, 7101, now)
	createAWDContestChallengeFixture(t, db, 71, 7101, now)
	syncAWDContestServiceFixture(t, db, 71, 7101, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 1, 2, now)
	syncAWDContestServiceReadinessFixture(t, db, 71, 7101, contestentity.AWDCheckerValidationStatePassed, nil, "")

	round, err := service.CreateRound(context.Background(), 71, contestcmd.CreateAWDRoundInput{
		RoundNumber: 1,
	})
	if err != nil {
		t.Fatalf("CreateRound() error = %v", err)
	}
	if round.AttackScore != 30 || round.DefenseScore != 3 {
		t.Fatalf("unexpected default round scores: %+v", round)
	}
}

func TestAWDServiceCreateRoundRejectsOversizedScores(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 72, now)
	createAWDChallengeFixture(t, db, 7201, now)
	createAWDContestChallengeFixture(t, db, 72, 7201, now)
	syncAWDContestServiceFixture(t, db, 72, 7201, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 1, 2, now)
	syncAWDContestServiceReadinessFixture(t, db, 72, 7201, contestentity.AWDCheckerValidationStatePassed, nil, "")

	_, err := service.CreateRound(context.Background(), 72, contestcmd.CreateAWDRoundInput{
		RoundNumber:  1,
		AttackScore:  intPtr(101),
		DefenseScore: intPtr(3),
	})
	if err == nil {
		t.Fatal("expected oversized attack score to be rejected")
	}
}

func TestAWDServiceUpsertServiceCheckAppliesDefenseScore(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 2, now)
	createAWDRoundFixture(t, db, 21, 2, 1, 70, 40, now)
	createAWDChallengeFixture(t, db, 201, now)
	createAWDContestChallengeFixture(t, db, 2, 201, now)
	createAWDTeamFixture(t, db, 211, 2, "Alpha", now)
	serviceID := defaultAWDContestServiceID(2, 201)

	var resp *contestcmd.AWDTeamServiceResp
	resp, err = service.UpsertServiceCheck(context.Background(), 2, 21, contestcmd.UpsertServiceCheckInput{
		TeamID:        211,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusUp,
		CheckResult: map[string]any{
			"is_alive":   true,
			"latency_ms": 12,
		},
	})
	if err != nil {
		t.Fatalf("UpsertServiceCheck() error = %v", err)
	}
	if resp.DefenseScore != 40 || resp.ServiceStatus != contestentity.AWDServiceStatusUp {
		t.Fatalf("unexpected service resp: %+v", resp)
	}
	if resp.SLAScore != 0 || resp.CheckerType != "" {
		t.Fatalf("unexpected sla/checker fields: %+v", resp)
	}
	if checkSource := resp.CheckResult["check_source"]; checkSource != awdCheckSourceManualService {
		t.Fatalf("unexpected check_source: %#v", checkSource)
	}
	if checkedAt, ok := resp.CheckResult["checked_at"].(string); !ok || checkedAt == "" {
		t.Fatalf("expected checked_at in manual service check result: %#v", resp.CheckResult)
	}
	assertTeamTotalScore(t, db, 211, 40)
	assertContestRedisScore(t, redisClient, 2, 211, 40)
	assertAWDServiceStatusCache(t, redisClient, 2, 211, serviceID, contestentity.AWDServiceStatusUp)

	resp, err = service.UpsertServiceCheck(context.Background(), 2, 21, contestcmd.UpsertServiceCheckInput{
		TeamID:        211,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusDown,
		CheckResult: map[string]any{
			"is_alive": false,
		},
	})
	if err != nil {
		t.Fatalf("second UpsertServiceCheck() error = %v", err)
	}
	if resp.DefenseScore != 0 || resp.ServiceStatus != contestentity.AWDServiceStatusDown {
		t.Fatalf("unexpected updated service resp: %+v", resp)
	}
	if resp.SLAScore != 0 || resp.CheckerType != "" {
		t.Fatalf("unexpected second sla/checker fields: %+v", resp)
	}
	if checkSource := resp.CheckResult["check_source"]; checkSource != awdCheckSourceManualService {
		t.Fatalf("unexpected second check_source: %#v", checkSource)
	}
	assertTeamTotalScore(t, db, 211, 0)
	assertContestRedisScoreMissing(t, redisClient, 2, 211)
	assertAWDServiceStatusCache(t, redisClient, 2, 211, serviceID, contestentity.AWDServiceStatusDown)
}

func TestAWDServiceRunCurrentRoundChecksRefreshesServices(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()

	createAWDContestFixture(t, db, 22, now)
	createAWDRoundFixture(t, db, 221, 22, 1, 70, 35, now)
	createAWDChallengeFixture(t, db, 2201, now)
	createAWDContestChallengeFixture(t, db, 22, 2201, now)
	syncAWDContestServiceFixture(t, db, 22, 2201, "awd-service", contestentity.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 0, 0, now)
	syncAWDContestServiceReadinessFixture(t, db, 22, 2201, contestentity.AWDCheckerValidationStatePassed, nil, "")
	createAWDTeamFixture(t, db, 2211, 22, "Ops", now)
	createAWDTeamMemberFixture(t, db, 22, 2211, 8201, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&instanceentity.Instance{
		ID:          8221,
		UserID:      8201,
		ChallengeID: 2201,
		ServiceID:   awdServiceIDPtr(22, 2201),
		ContainerID: "ctr-ops",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/health",
	})

	resp, err := service.RunCurrentRoundChecks(context.Background(), 22, contestcmd.RunCurrentRoundChecksInput{})
	if err != nil {
		t.Fatalf("RunCurrentRoundChecks() error = %v", err)
	}
	if resp.Round == nil || resp.Round.ID != 221 {
		t.Fatalf("unexpected round resp: %+v", resp.Round)
	}
	if len(resp.Services) != 1 {
		t.Fatalf("unexpected services: %+v", resp.Services)
	}
	if resp.Services[0].ServiceStatus != contestentity.AWDServiceStatusUp || resp.Services[0].DefenseScore != 35 {
		t.Fatalf("unexpected service status: %+v", resp.Services[0])
	}
	if checkSource := resp.Services[0].CheckResult["check_source"]; checkSource != awdCheckSourceManualCurrent {
		t.Fatalf("unexpected check_source: %#v", checkSource)
	}
	if statusReason := resp.Services[0].CheckResult["status_reason"]; statusReason != "healthy" {
		t.Fatalf("unexpected status_reason: %#v", statusReason)
	}
	checkResult, ok := resp.Services[0].CheckResult["targets"].([]any)
	if !ok || len(checkResult) != 1 {
		t.Fatalf("unexpected targets payload: %#v", resp.Services[0].CheckResult["targets"])
	}
	if !strings.Contains(server.URL, "http") {
		t.Fatalf("unexpected server url: %s", server.URL)
	}
	assertTeamTotalScore(t, db, 2211, 35)
	assertContestRedisScore(t, redisClient, 22, 2211, 35)
}

func TestAWDServiceRunCurrentRoundChecksRejectsEndedContest(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	now := time.Now()
	createAWDContestFixture(t, db, 222, now)
	createAWDRoundFixtureWithWindow(t, db, 2221, 222, 1, 70, 35, now.Add(-10*time.Minute), now.Add(-5*time.Minute))
	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 222).Updates(map[string]any{
		"status":   contestentity.ContestStatusEnded,
		"end_time": now.Add(-time.Minute),
	}).Error; err != nil {
		t.Fatalf("set contest ended: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(222), "1", 0).Err(); err != nil {
		t.Fatalf("seed stale current round: %v", err)
	}

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})

	_, err = service.RunCurrentRoundChecks(context.Background(), 222, contestcmd.RunCurrentRoundChecksInput{})
	if err != contestcontracts.ErrContestEnded {
		t.Fatalf("expected ErrContestEnded, got %v", err)
	}
}

func TestAWDServiceCreateRoundBlocksWhenReadinessNotReady(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 110, now)

	_, err := service.CreateRound(context.Background(), 110, contestcmd.CreateAWDRoundInput{
		RoundNumber: 1,
	})
	assertAWDReadinessBlocked(t, err)
}

func TestAWDServiceCreateRoundAllowsForceOverrideWithReason(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 111, now)

	resp, err := service.CreateRound(context.Background(), 111, contestcmd.CreateAWDRoundInput{
		RoundNumber:    1,
		ForceOverride:  boolPtr(true),
		OverrideReason: strPtr("teacher drill"),
		AttackScore:    intPtr(80),
		DefenseScore:   intPtr(3),
	})
	if err != nil {
		t.Fatalf("CreateRound() error = %v", err)
	}
	if resp == nil || resp.RoundNumber != 1 {
		t.Fatalf("unexpected round response: %+v", resp)
	}
}

func TestAWDServiceCreateRoundRejectsBlankOverrideReason(t *testing.T) {
	db := newAWDTestDB(t)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 112, now)

	_, err := service.CreateRound(context.Background(), 112, contestcmd.CreateAWDRoundInput{
		RoundNumber:    1,
		ForceOverride:  boolPtr(true),
		OverrideReason: strPtr("   "),
	})
	if err != apperror.ErrInvalidParams {
		t.Fatalf("expected ErrInvalidParams, got %v", err)
	}
}

func TestAWDServiceRunCurrentRoundChecksBlocksWhenReadinessNotReady(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 240, now)
	createAWDRoundFixture(t, db, 2401, 240, 1, 70, 35, now)

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})

	_, err := service.RunCurrentRoundChecks(context.Background(), 240, contestcmd.RunCurrentRoundChecksInput{})
	assertAWDReadinessBlocked(t, err)
}

func TestAWDServiceRunCurrentRoundChecksRejectsBlankOverrideReason(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 241, now)
	createAWDRoundFixture(t, db, 2411, 241, 1, 70, 35, now)

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})

	_, err := service.RunCurrentRoundChecks(context.Background(), 241, contestcmd.RunCurrentRoundChecksInput{
		ForceOverride:  boolPtr(true),
		OverrideReason: strPtr("  "),
	})
	if err != apperror.ErrInvalidParams {
		t.Fatalf("expected ErrInvalidParams, got %v", err)
	}
}

func TestAWDServiceRunRoundChecksSkipsReadinessGate(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 242, now)
	createAWDRoundFixture(t, db, 2421, 242, 1, 70, 35, now)

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})

	resp, err := service.RunRoundChecks(context.Background(), 242, 2421)
	if err != nil {
		t.Fatalf("RunRoundChecks() error = %v", err)
	}
	if resp == nil || resp.Round == nil || resp.Round.ID != 2421 {
		t.Fatalf("unexpected selected round response: %+v", resp)
	}
}

func TestAWDServiceRunRoundChecksRefreshesSelectedRound(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 23, now)
	createAWDRoundFixture(t, db, 231, 23, 1, 80, 45, now)
	createAWDChallengeFixture(t, db, 2301, now)
	createAWDContestChallengeFixture(t, db, 23, 2301, now)
	createAWDTeamFixture(t, db, 2311, 23, "Selected Ops", now)
	createAWDTeamMemberFixture(t, db, 23, 2311, 8301, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	if err := db.Create(&instanceentity.Instance{
		ID:          8321,
		UserID:      8301,
		ChallengeID: 2301,
		ServiceID:   awdServiceIDPtr(23, 2301),
		ContainerID: "ctr-selected-ops",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   server.URL,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create awd instance: %v", err)
	}

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/health",
	})

	resp, err := service.RunRoundChecks(context.Background(), 23, 231)
	if err != nil {
		t.Fatalf("RunRoundChecks() error = %v", err)
	}
	if resp.Round == nil || resp.Round.ID != 231 {
		t.Fatalf("unexpected round resp: %+v", resp.Round)
	}
	if len(resp.Services) != 1 {
		t.Fatalf("unexpected services: %+v", resp.Services)
	}
	if resp.Services[0].ServiceStatus != contestentity.AWDServiceStatusUp || resp.Services[0].DefenseScore != 45 {
		t.Fatalf("unexpected service status: %+v", resp.Services[0])
	}
	if checkSource := resp.Services[0].CheckResult["check_source"]; checkSource != awdCheckSourceManualSelected {
		t.Fatalf("unexpected check_source: %#v", checkSource)
	}
	if statusReason := resp.Services[0].CheckResult["status_reason"]; statusReason != "healthy" {
		t.Fatalf("unexpected status_reason: %#v", statusReason)
	}
}

func TestAWDServiceCreateAttackLogDeduplicatesScoringAndBuildsSummary(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 3, now)
	createAWDRoundFixture(t, db, 31, 3, 1, 60, 25, now)
	createAWDChallengeFixture(t, db, 301, now)
	createAWDContestChallengeFixture(t, db, 3, 301, now)
	createAWDTeamFixture(t, db, 311, 3, "Red", now)
	createAWDTeamFixture(t, db, 312, 3, "Blue", now)
	createAWDTeamFixture(t, db, 313, 3, "Green", now)
	serviceID := defaultAWDContestServiceID(3, 301)

	if _, err := service.UpsertServiceCheck(context.Background(), 3, 31, contestcmd.UpsertServiceCheckInput{
		TeamID:        311,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusUp,
		CheckResult:   map[string]any{"latency_ms": 10},
	}); err != nil {
		t.Fatalf("seed Red service check: %v", err)
	}
	if _, err := service.UpsertServiceCheck(context.Background(), 3, 31, contestcmd.UpsertServiceCheckInput{
		TeamID:        312,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusCompromised,
		CheckResult:   map[string]any{"latency_ms": 25},
	}); err != nil {
		t.Fatalf("seed Blue service check: %v", err)
	}
	if _, err := service.UpsertServiceCheck(context.Background(), 3, 31, contestcmd.UpsertServiceCheckInput{
		TeamID:        313,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusUp,
		CheckResult:   map[string]any{"latency_ms": 8},
	}); err != nil {
		t.Fatalf("seed Green service check: %v", err)
	}
	if err := db.Model(&contestentity.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 311, 301).
		Updates(map[string]any{
			"sla_score":    10,
			"checker_type": contestentity.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Red sla/checker fields: %v", err)
	}
	if err := db.Model(&contestentity.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 312, 301).
		Updates(map[string]any{
			"sla_score":    9,
			"checker_type": contestentity.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Blue sla/checker fields: %v", err)
	}
	if err := db.Model(&contestentity.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 313, 301).
		Updates(map[string]any{
			"sla_score":    8,
			"checker_type": contestentity.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Green sla/checker fields: %v", err)
	}

	first, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 311,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() error = %v", err)
	}
	if first.Source != contestentity.AWDAttackSourceManual {
		t.Fatalf("expected manual attack source, got %+v", first)
	}
	if first.ScoreGained != 60 {
		t.Fatalf("expected first score gained 60, got %+v", first)
	}

	second, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 311,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() duplicate error = %v", err)
	}
	if second.ScoreGained != 0 {
		t.Fatalf("expected duplicate score gained 0, got %+v", second)
	}
	var blueService contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 312, 301).First(&blueService).Error; err != nil {
		t.Fatalf("load Blue service: %v", err)
	}
	if blueService.ServiceStatus != contestentity.AWDServiceStatusCompromised || blueService.AttackReceived != 2 || blueService.AttackScore != 60 || blueService.DefenseScore != 0 || blueService.SLAScore != 9 || blueService.CheckerType != contestentity.AWDCheckerTypeHTTPStandard {
		t.Fatalf("unexpected Blue service impact: %+v", blueService)
	}

	if _, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 313,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeServiceExploit,
		IsSuccess:      true,
	}); err != nil {
		t.Fatalf("CreateAttackLog() second attacker error = %v", err)
	}

	summary, err := service.GetRoundSummary(context.Background(), 3, 31)
	if err != nil {
		t.Fatalf("GetRoundSummary() error = %v", err)
	}
	if len(summary.Items) != 3 {
		t.Fatalf("unexpected summary size: %+v", summary.Items)
	}
	if summary.Metrics == nil {
		t.Fatalf("expected round metrics in summary")
	}
	if summary.Metrics.TotalServiceCount != 3 || summary.Metrics.ServiceUpCount != 2 || summary.Metrics.ServiceCompromisedCount != 1 {
		t.Fatalf("unexpected service metrics: %+v", summary.Metrics)
	}
	if summary.Metrics.AttackedServiceCount != 1 || summary.Metrics.DefenseSuccessCount != 0 {
		t.Fatalf("unexpected defense metrics: %+v", summary.Metrics)
	}
	if summary.Metrics.TotalAttackCount != 3 || summary.Metrics.SuccessfulAttackCount != 3 || summary.Metrics.FailedAttackCount != 0 {
		t.Fatalf("unexpected attack metrics: %+v", summary.Metrics)
	}
	if summary.Metrics.ManualServiceCheckCount != 3 || summary.Metrics.ManualAttackLogCount != 3 {
		t.Fatalf("unexpected source metrics: %+v", summary.Metrics)
	}

	red := findAWDSummaryItem(summary.Items, 311)
	if red == nil || red.AttackScore != 60 || red.DefenseScore != 25 || red.SLAScore != 10 || red.TotalScore != 95 {
		t.Fatalf("unexpected red summary: %+v", red)
	}
	blue := findAWDSummaryItem(summary.Items, 312)
	if blue == nil || blue.ServiceCompromisedCount != 1 || blue.DefenseScore != 0 || blue.SLAScore != 9 || blue.SuccessfulBreachCount != 3 || blue.UniqueAttackersAgainst != 2 || blue.TotalScore != 9 {
		t.Fatalf("unexpected blue summary: %+v", blue)
	}
	green := findAWDSummaryItem(summary.Items, 313)
	if green == nil || green.AttackScore != 60 || green.SuccessfulAttackCount != 1 || green.ServiceUpCount != 1 || green.SLAScore != 8 || green.TotalScore != 93 {
		t.Fatalf("unexpected green summary: %+v", green)
	}
	assertTeamTotalScore(t, db, 311, 35)
	assertTeamTotalScore(t, db, 312, 9)
	assertTeamTotalScore(t, db, 313, 33)
	assertContestRedisScore(t, redisClient, 3, 311, 35)
	assertContestRedisScore(t, redisClient, 3, 312, 9)
	assertContestRedisScore(t, redisClient, 3, 313, 33)

	scoreboardService := contestqry.NewScoreboardService(contestinfra.NewRepository(db), contestinfra.NewContestScoreboardStateStore(redisClient), &config.ContestConfig{}, zap.NewNop())
	scoreboard, err := scoreboardService.GetLiveScoreboard(context.Background(), 3, 1, 10)
	if err != nil {
		t.Fatalf("GetLiveScoreboard() error = %v", err)
	}
	if scoreboard.Scoreboard == nil || len(scoreboard.Scoreboard.List) != 3 {
		t.Fatalf("unexpected live scoreboard: %+v", scoreboard)
	}
	if scoreboard.Scoreboard.List[0].SolvedCount != 0 || scoreboard.Scoreboard.List[1].SolvedCount != 0 {
		t.Fatalf("expected manual attack logs excluded from official solved_count: %+v", scoreboard.Scoreboard.List)
	}
}

func TestAWDServiceCreateAttackLogCreatesVictimServiceImpactWhenMissing(t *testing.T) {
	db := newAWDTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 6, now)
	createAWDRoundFixture(t, db, 61, 6, 1, 75, 20, now)
	createAWDChallengeFixture(t, db, 601, now)
	createAWDContestChallengeFixture(t, db, 6, 601, now)
	createAWDTeamFixture(t, db, 611, 6, "Red", now)
	createAWDTeamFixture(t, db, 612, 6, "Blue", now)
	serviceID := defaultAWDContestServiceID(6, 601)

	resp, err := service.CreateAttackLog(context.Background(), 6, 61, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 611,
		VictimTeamID:   612,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() error = %v", err)
	}
	if resp.Source != contestentity.AWDAttackSourceManual {
		t.Fatalf("expected manual source, got %+v", resp)
	}
	if resp.ScoreGained != 75 {
		t.Fatalf("unexpected score gained: %+v", resp)
	}
	assertAWDServiceStatusCache(t, redisClient, 6, 612, serviceID, contestentity.AWDServiceStatusCompromised)

	var victimService contestentity.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 61, 612, 601).First(&victimService).Error; err != nil {
		t.Fatalf("load victim service: %v", err)
	}
	if victimService.ServiceStatus != contestentity.AWDServiceStatusCompromised || victimService.AttackReceived != 1 || victimService.AttackScore != 75 || victimService.DefenseScore != 0 {
		t.Fatalf("unexpected victim service: %+v", victimService)
	}
}

func TestAWDServiceHistoricalManualUpdatesDoNotOverrideLiveServiceStatusCache(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 16, now)
	createAWDRoundFixtureWithWindow(t, db, 161, 16, 1, 60, 30, now.Add(-10*time.Minute), now.Add(-5*time.Minute))
	createAWDRoundFixtureWithWindow(t, db, 162, 16, 2, 60, 30, now.Add(-5*time.Minute), time.Time{})
	createAWDChallengeFixture(t, db, 1601, now)
	createAWDContestChallengeFixture(t, db, 16, 1601, now)
	createAWDTeamFixture(t, db, 1611, 16, "Alpha", now)
	serviceID := defaultAWDContestServiceID(16, 1601)

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(16), "2", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	if _, err := service.UpsertServiceCheck(context.Background(), 16, 161, contestcmd.UpsertServiceCheckInput{
		TeamID:        1611,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusDown,
		CheckResult:   map[string]any{"reason": "historical-fix"},
	}); err != nil {
		t.Fatalf("historical UpsertServiceCheck() error = %v", err)
	}

	assertAWDServiceStatusCacheMissing(t, redisClient, 16, 1611, serviceID)
}

func TestAWDServiceEndedContestManualUpdatesDoNotRestoreLiveServiceStatusCache(t *testing.T) {
	db := newAWDTestDB(t)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 17, now)
	createAWDRoundFixture(t, db, 171, 17, 1, 60, 30, now)
	createAWDChallengeFixture(t, db, 1701, now)
	createAWDContestChallengeFixture(t, db, 17, 1701, now)
	createAWDTeamFixture(t, db, 1711, 17, "Alpha", now)
	serviceID := defaultAWDContestServiceID(17, 1701)

	if err := db.Model(&contestentity.Contest{}).Where("id = ?", 17).Updates(map[string]any{
		"status":   contestentity.ContestStatusEnded,
		"end_time": now.Add(-time.Minute),
	}).Error; err != nil {
		t.Fatalf("set contest ended: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(17), "1", 0).Err(); err != nil {
		t.Fatalf("seed stale current round: %v", err)
	}

	if _, err := service.UpsertServiceCheck(context.Background(), 17, 171, contestcmd.UpsertServiceCheckInput{
		TeamID:        1711,
		ServiceID:     serviceID,
		ServiceStatus: contestentity.AWDServiceStatusUp,
		CheckResult:   map[string]any{"reason": "postmortem-fix"},
	}); err != nil {
		t.Fatalf("ended contest UpsertServiceCheck() error = %v", err)
	}

	assertAWDServiceStatusCacheMissing(t, redisClient, 17, 1711, serviceID)
}

func TestAWDServiceGetTrafficSummaryBuildsAggregateMetrics(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})

	createAWDContestFixture(t, db, 90, now)
	createAWDRoundFixtureWithWindow(t, db, 901, 90, 3, 60, 40, now.Add(-10*time.Minute), now.Add(-5*time.Minute))
	createAWDChallengeFixture(t, db, 9001, now)
	createAWDChallengeFixture(t, db, 9002, now)
	createAWDContestChallengeFixture(t, db, 90, 9001, now)
	createAWDContestChallengeFixture(t, db, 90, 9002, now)
	createAWDTeamFixture(t, db, 9101, 90, "Red", now)
	createAWDTeamFixture(t, db, 9102, 90, "Blue", now)
	createAWDTeamMemberFixture(t, db, 90, 9101, 9201, now)
	createAWDTeamMemberFixture(t, db, 90, 9102, 9202, now)

	if err := db.Create(&instanceentity.Instance{
		ID:          9301,
		UserID:      9202,
		ContestID:   int64Ptr(90),
		TeamID:      int64Ptr(9102),
		ChallengeID: 9001,
		ServiceID:   awdServiceIDPtr(90, 9001),
		ContainerID: "ctr-blue-web",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://blue-web.local",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create blue instance: %v", err)
	}
	if err := db.Create(&instanceentity.Instance{
		ID:          9302,
		UserID:      9201,
		ContestID:   int64Ptr(90),
		TeamID:      int64Ptr(9101),
		ChallengeID: 9002,
		ServiceID:   awdServiceIDPtr(90, 9002),
		ContainerID: "ctr-red-pwn",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://red-pwn.local",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create red instance: %v", err)
	}

	mustCreateAWDTrafficEvent(t, db, 9401, 90, 901, 9101, 9102, defaultAWDContestServiceID(90, 9001), 9001, "GET", "/health", 200, now.Add(-9*time.Minute))
	mustCreateAWDTrafficEvent(t, db, 9402, 90, 901, 9101, 9102, defaultAWDContestServiceID(90, 9001), 9001, "POST", "/admin/login", 500, now.Add(-8*time.Minute))
	mustCreateAWDTrafficEvent(t, db, 9403, 90, 901, 9102, 9101, defaultAWDContestServiceID(90, 9002), 9002, "GET", "/index", 302, now.Add(-7*time.Minute))

	summary, err := service.GetTrafficSummary(context.Background(), 90, 901)
	if err != nil {
		t.Fatalf("GetTrafficSummary() error = %v", err)
	}
	if summary.TotalRequests != 3 || summary.ErrorRequests != 1 {
		t.Fatalf("unexpected traffic summary counts: %+v", summary)
	}
	if summary.ActiveAttackerTeams != 2 || summary.TargetedTeams != 2 {
		t.Fatalf("unexpected active/targeted teams: %+v", summary)
	}
	if summary.UniquePathCount != 3 {
		t.Fatalf("unexpected unique path count: %+v", summary)
	}
	if len(summary.TopAttackers) == 0 || summary.TopAttackers[0].TeamID != 9101 || summary.TopAttackers[0].RequestCount != 2 {
		t.Fatalf("unexpected top attackers: %+v", summary.TopAttackers)
	}
	if len(summary.TopVictims) == 0 || summary.TopVictims[0].TeamID != 9102 || summary.TopVictims[0].RequestCount != 2 {
		t.Fatalf("unexpected top victims: %+v", summary.TopVictims)
	}
	if len(summary.TopPaths) == 0 || summary.TopPaths[0].Path != "/admin/login" || summary.TopPaths[0].ErrorCount != 1 {
		t.Fatalf("unexpected top paths: %+v", summary.TopPaths)
	}
	if len(summary.Trend) != 3 {
		t.Fatalf("unexpected trend buckets: %+v", summary.Trend)
	}
}

func TestAWDServiceListTrafficEventsSupportsFiltersAndPagination(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{})

	createAWDContestFixture(t, db, 91, now)
	createAWDRoundFixtureWithWindow(t, db, 911, 91, 4, 60, 40, now.Add(-20*time.Minute), now.Add(-10*time.Minute))
	createAWDChallengeFixture(t, db, 91001, now)
	createAWDContestChallengeFixture(t, db, 91, 91001, now)
	createAWDTeamFixture(t, db, 9111, 91, "Alpha", now)
	createAWDTeamFixture(t, db, 9112, 91, "Beta", now)
	createAWDTeamMemberFixture(t, db, 91, 9111, 9211, now)
	createAWDTeamMemberFixture(t, db, 91, 9112, 9212, now)

	if err := db.Create(&instanceentity.Instance{
		ID:          9311,
		UserID:      9212,
		ContestID:   int64Ptr(91),
		TeamID:      int64Ptr(9112),
		ChallengeID: 91001,
		ServiceID:   awdServiceIDPtr(91, 91001),
		ContainerID: "ctr-beta-web",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://beta-web.local",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create beta instance: %v", err)
	}

	mustCreateAWDTrafficEvent(t, db, 9411, 91, 911, 9111, 9112, defaultAWDContestServiceID(91, 91001), 91001, "GET", "/api/status", 200, now.Add(-19*time.Minute))
	mustCreateAWDTrafficEvent(t, db, 9412, 91, 911, 9111, 9112, defaultAWDContestServiceID(91, 91001), 91001, "POST", "/admin/login", 401, now.Add(-18*time.Minute))
	mustCreateAWDTrafficEvent(t, db, 9413, 91, 911, 9111, 9112, defaultAWDContestServiceID(91, 91001), 91001, "POST", "/admin/login", 500, now.Add(-17*time.Minute))

	page, err := service.ListTrafficEvents(context.Background(), 91, 911, contestqry.ListAWDTrafficEventsInput{
		StatusGroup: "server_error",
		PathKeyword: "login",
		Page:        1,
		Size:        1,
	})
	if err != nil {
		t.Fatalf("ListTrafficEvents() error = %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 {
		t.Fatalf("unexpected traffic page: %+v", page)
	}
	if page.List[0].StatusCode != 500 || page.List[0].Path != "/admin/login" {
		t.Fatalf("unexpected filtered traffic event: %+v", page.List[0])
	}
	if page.List[0].ServiceID != defaultAWDContestServiceID(91, 91001) {
		t.Fatalf("expected traffic event to expose service_id, got %+v", page.List[0])
	}

	emptyPage, err := service.ListTrafficEvents(context.Background(), 91, 911, contestqry.ListAWDTrafficEventsInput{
		ServiceID: defaultAWDContestServiceID(91, 91001) + 1,
		Page:      1,
		Size:      20,
	})
	if err != nil {
		t.Fatalf("ListTrafficEvents() with service_id filter error = %v", err)
	}
	if emptyPage.Total != 0 || len(emptyPage.List) != 0 {
		t.Fatalf("expected service_id filter to exclude all traffic events, got %+v", emptyPage)
	}
}

func findAWDSummaryItem(items []*contestqry.AWDRoundSummaryItemResult, teamID int64) *contestqry.AWDRoundSummaryItemResult {
	for _, item := range items {
		if item.TeamID == teamID {
			return item
		}
	}
	return nil
}

func intPtr(v int) *int { return &v }

func int64Ptr(v int64) *int64 { return &v }

func boolPtr(v bool) *bool { return &v }

func strPtr(v string) *string { return &v }

func assertAWDReadinessBlocked(t *testing.T, err error) {
	t.Helper()

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != contestcontracts.ErrAWDReadinessBlocked.Code {
		t.Fatalf("expected ErrAWDReadinessBlocked, got %v", err)
	}
}

func mustCreateAWDTrafficEvent(
	t *testing.T,
	db *gorm.DB,
	id int64,
	contestID int64,
	roundID int64,
	attackerTeamID int64,
	victimTeamID int64,
	serviceID int64,
	challengeID int64,
	method string,
	requestPath string,
	statusCode int,
	createdAt time.Time,
) {
	t.Helper()

	if err := db.Create(&contestentity.AWDTrafficEvent{
		ID:             id,
		ContestID:      contestID,
		RoundID:        roundID,
		AttackerTeamID: attackerTeamID,
		VictimTeamID:   victimTeamID,
		ServiceID:      serviceID,
		AWDChallengeID: challengeID,
		Method:         method,
		Path:           requestPath,
		StatusCode:     statusCode,
		Source:         contestentity.AWDTrafficSourceRuntimeProxy,
		CreatedAt:      createdAt,
	}).Error; err != nil {
		t.Fatalf("create awd traffic event: %v", err)
	}
}
