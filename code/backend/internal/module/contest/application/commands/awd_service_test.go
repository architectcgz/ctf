package commands_test

import (
	"bufio"
	"context"
	"errors"
	"net"
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

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/module/contest/testsupport"
	rediskeys "ctf-platform/internal/pkg/redis"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
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
	containerDetails   model.InstanceRuntimeDetails
	containerErr       error
	cleanupErr         error
}

func (f *fakeContestPreviewRuntimeProbe) CreateTopology(_ context.Context, _ *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	return nil, errors.New("unexpected CreateTopology call")
}

func (f *fakeContestPreviewRuntimeProbe) CreateContainer(_ context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
	f.createContainerCalled = true
	f.lastImageName = imageName
	f.lastEnv = env
	if f.containerErr != nil {
		return "", model.InstanceRuntimeDetails{}, f.containerErr
	}
	return f.containerAccessURL, f.containerDetails, nil
}

func (f *fakeContestPreviewRuntimeProbe) CleanupRuntimeDetails(_ context.Context, details model.InstanceRuntimeDetails) error {
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

func (f *fakeAWDPreviewRoundManager) RunRoundServiceChecks(_ context.Context, _ *model.Contest, _ *model.AWDRound, _ string) error {
	return errors.New("unexpected RunRoundServiceChecks call")
}

func (f *fakeAWDPreviewRoundManager) EnsureActiveRoundMaterialized(_ context.Context, _ *model.Contest, _ time.Time) error {
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
	return contestinfra.NewAWDPreviewRuntimeImageRepository(challengeinfra.NewImageRepository(db)),
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

func (s *awdServiceForTest) CreateRound(ctx context.Context, contestID int64, req contestcmd.CreateAWDRoundInput) (*dto.AWDRoundResp, error) {
	return s.commands.CreateRound(ctx, contestID, req)
}

func (s *awdServiceForTest) ListRounds(ctx context.Context, contestID int64) ([]contestqry.AWDRoundResult, error) {
	return s.queries.ListRounds(ctx, contestID)
}

func (s *awdServiceForTest) RunCurrentRoundChecks(ctx context.Context, contestID int64, req contestcmd.RunCurrentRoundChecksInput) (*dto.AWDCheckerRunResp, error) {
	return s.commands.RunCurrentRoundChecks(ctx, contestID, req)
}

func (s *awdServiceForTest) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error) {
	return s.commands.RunRoundChecks(ctx, contestID, roundID)
}

func awdServiceIDPtr(contestID, challengeID int64) *int64 {
	id := defaultAWDContestServiceID(contestID, challengeID)
	return &id
}

func setReflectedField(t *testing.T, target reflect.Value, field string, value any) {
	t.Helper()
	item := target.FieldByName(field)
	if !item.IsValid() {
		t.Fatalf("field %s not found", field)
	}
	if !item.CanSet() {
		t.Fatalf("field %s cannot set", field)
	}

	next := reflect.ValueOf(value)
	if !next.IsValid() {
		item.Set(reflect.Zero(item.Type()))
		return
	}
	if next.Type().AssignableTo(item.Type()) {
		item.Set(next)
		return
	}
	if next.Type().ConvertibleTo(item.Type()) {
		item.Set(next.Convert(item.Type()))
		return
	}
	t.Fatalf("field %s type mismatch: have %s want %s", field, next.Type(), item.Type())
}

func (s *awdServiceForTest) UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req contestcmd.UpsertServiceCheckInput) (*dto.AWDTeamServiceResp, error) {
	return s.commands.UpsertServiceCheck(ctx, contestID, roundID, req)
}

func (s *awdServiceForTest) CreateAttackLog(ctx context.Context, contestID, roundID int64, req contestcmd.CreateAttackLogInput) (*dto.AWDAttackLogResp, error) {
	return s.commands.CreateAttackLog(ctx, contestID, roundID, req)
}

func (s *awdServiceForTest) SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req contestcmd.SubmitAttackInput) (*dto.AWDAttackLogResp, error) {
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
	syncAWDContestServiceFixture(t, db, 1, 101, "awd-service", model.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 0, 0, now)
	syncAWDContestServiceReadinessFixture(t, db, 1, 101, model.AWDCheckerValidationStatePassed, nil, "")

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
	syncAWDContestServiceFixture(t, db, 71, 7101, "awd-service", model.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 1, 2, now)
	syncAWDContestServiceReadinessFixture(t, db, 71, 7101, model.AWDCheckerValidationStatePassed, nil, "")

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
	syncAWDContestServiceFixture(t, db, 72, 7201, "awd-service", model.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 1, 2, now)
	syncAWDContestServiceReadinessFixture(t, db, 72, 7201, model.AWDCheckerValidationStatePassed, nil, "")

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

	var resp *dto.AWDTeamServiceResp
	resp, err = service.UpsertServiceCheck(context.Background(), 2, 21, contestcmd.UpsertServiceCheckInput{
		TeamID:        211,
		ServiceID:     serviceID,
		ServiceStatus: model.AWDServiceStatusUp,
		CheckResult: map[string]any{
			"is_alive":   true,
			"latency_ms": 12,
		},
	})
	if err != nil {
		t.Fatalf("UpsertServiceCheck() error = %v", err)
	}
	if resp.DefenseScore != 40 || resp.ServiceStatus != model.AWDServiceStatusUp {
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
	assertAWDServiceStatusCache(t, redisClient, 2, 211, serviceID, model.AWDServiceStatusUp)

	resp, err = service.UpsertServiceCheck(context.Background(), 2, 21, contestcmd.UpsertServiceCheckInput{
		TeamID:        211,
		ServiceID:     serviceID,
		ServiceStatus: model.AWDServiceStatusDown,
		CheckResult: map[string]any{
			"is_alive": false,
		},
	})
	if err != nil {
		t.Fatalf("second UpsertServiceCheck() error = %v", err)
	}
	if resp.DefenseScore != 0 || resp.ServiceStatus != model.AWDServiceStatusDown {
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
	assertAWDServiceStatusCache(t, redisClient, 2, 211, serviceID, model.AWDServiceStatusDown)
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
	syncAWDContestServiceFixture(t, db, 22, 2201, "awd-service", model.AWDCheckerTypeHTTPStandard, `{"get_flag":{"path":"/health"}}`, 100, 0, 0, now)
	syncAWDContestServiceReadinessFixture(t, db, 22, 2201, model.AWDCheckerValidationStatePassed, nil, "")
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

	if err := db.Create(&model.Instance{
		ID:          8221,
		UserID:      8201,
		ChallengeID: 2201,
		ServiceID:   awdServiceIDPtr(22, 2201),
		ContainerID: "ctr-ops",
		Status:      model.InstanceStatusRunning,
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
	if resp.Services[0].ServiceStatus != model.AWDServiceStatusUp || resp.Services[0].DefenseScore != 35 {
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
	if err := db.Model(&model.Contest{}).Where("id = ?", 222).Updates(map[string]any{
		"status":   model.ContestStatusEnded,
		"end_time": now.Add(-time.Minute),
	}).Error; err != nil {
		t.Fatalf("set contest ended: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(222), "1", 0).Err(); err != nil {
		t.Fatalf("seed stale current round: %v", err)
	}

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{})

	_, err = service.RunCurrentRoundChecks(context.Background(), 222, contestcmd.RunCurrentRoundChecksInput{})
	if err != errcode.ErrContestEnded {
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
	if err != errcode.ErrInvalidParams {
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
	if err != errcode.ErrInvalidParams {
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

	if err := db.Create(&model.Instance{
		ID:          8321,
		UserID:      8301,
		ChallengeID: 2301,
		ServiceID:   awdServiceIDPtr(23, 2301),
		ContainerID: "ctr-selected-ops",
		Status:      model.InstanceStatusRunning,
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
	if resp.Services[0].ServiceStatus != model.AWDServiceStatusUp || resp.Services[0].DefenseScore != 45 {
		t.Fatalf("unexpected service status: %+v", resp.Services[0])
	}
	if checkSource := resp.Services[0].CheckResult["check_source"]; checkSource != awdCheckSourceManualSelected {
		t.Fatalf("unexpected check_source: %#v", checkSource)
	}
	if statusReason := resp.Services[0].CheckResult["status_reason"]; statusReason != "healthy" {
		t.Fatalf("unexpected status_reason: %#v", statusReason)
	}
}

func TestAWDServicePreviewCheckerRunsWithoutPersistingServices(t *testing.T) {
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

	createAWDContestFixture(t, db, 24, now)
	createAWDChallengeFixture(t, db, 2401, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/flag":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte("flag{preview}"))
				return
			}
			http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	method := reflect.ValueOf(service.commands).MethodByName("PreviewChecker")
	if !method.IsValid() {
		t.Fatalf("PreviewChecker method not implemented")
	}

	reqValue := reflect.New(method.Type().In(2))
	setReflectedField(t, reqValue.Elem(), "AWDChallengeID", int64(2401))
	setReflectedField(t, reqValue.Elem(), "CheckerType", string(model.AWDCheckerTypeHTTPStandard))
	setReflectedField(t, reqValue.Elem(), "CheckerConfig", map[string]any{
		"put_flag": map[string]any{
			"method":          "PUT",
			"path":            "/api/flag",
			"expected_status": http.StatusCreated,
			"body_template":   "{{FLAG}}",
		},
		"get_flag": map[string]any{
			"method":             "GET",
			"path":               "/api/flag",
			"expected_status":    http.StatusOK,
			"expected_substring": "{{FLAG}}",
		},
		"havoc": map[string]any{
			"method":          "GET",
			"path":            "/healthz",
			"expected_status": http.StatusOK,
		},
	})
	setReflectedField(t, reqValue.Elem(), "AccessURL", server.URL)
	setReflectedField(t, reqValue.Elem(), "PreviewFlag", "flag{preview}")

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(int64(24)),
		reqValue.Elem(),
	})

	if len(results) != 2 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	if errValue := results[1].Interface(); errValue != nil {
		t.Fatalf("PreviewChecker() error = %v", errValue)
	}

	respValue := results[0]
	if respValue.IsNil() {
		t.Fatal("expected preview response")
	}
	resp := respValue.Elem()
	if serviceStatus := resp.FieldByName("ServiceStatus").String(); serviceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", serviceStatus)
	}
	if checkerType := resp.FieldByName("CheckerType").String(); checkerType != string(model.AWDCheckerTypeHTTPStandard) {
		t.Fatalf("unexpected checker type: %s", checkerType)
	}

	checkResultField := resp.FieldByName("CheckResult")
	if !checkResultField.IsValid() || checkResultField.IsNil() {
		t.Fatal("expected check result")
	}
	checkResult, ok := checkResultField.Interface().(map[string]any)
	if !ok {
		t.Fatalf("unexpected check result type: %T", checkResultField.Interface())
	}
	if source := checkResult["check_source"]; source != "checker_preview" {
		t.Fatalf("unexpected check_source: %#v", source)
	}
	if reason := checkResult["status_reason"]; reason != "preview_quorum_passed" {
		t.Fatalf("unexpected status_reason: %#v", reason)
	}
	if summary := checkResult["preview_summary"]; summary != "3/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", summary)
	}

	previewContext := resp.FieldByName("PreviewContext")
	if !previewContext.IsValid() {
		t.Fatal("expected preview context")
	}
	if accessURL := previewContext.FieldByName("AccessURL").String(); accessURL != server.URL {
		t.Fatalf("unexpected preview access_url: %s", accessURL)
	}
	if previewFlag := previewContext.FieldByName("PreviewFlag").String(); previewFlag != "flag{preview}" {
		t.Fatalf("unexpected preview flag: %s", previewFlag)
	}
	previewToken := resp.FieldByName("PreviewToken")
	if !previewToken.IsValid() || strings.TrimSpace(previewToken.String()) == "" {
		t.Fatal("expected preview_token")
	}

	var serviceCount int64
	if err := db.Model(&model.AWDTeamService{}).Count(&serviceCount).Error; err != nil {
		t.Fatalf("count awd team services: %v", err)
	}
	if serviceCount != 0 {
		t.Fatalf("expected no persisted awd team services, got %d", serviceCount)
	}
}

func TestAWDServicePreviewCheckerTCPStandardTokenMakesReadinessPassed(t *testing.T) {
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
	const contestID int64 = 291
	const awdChallengeID int64 = 291001
	createAWDContestFixture(t, db, contestID, now)
	createAWDChallengeFixture(t, db, awdChallengeID, now)
	if err := db.Create(&model.AWDChallenge{
		ID:             awdChallengeID,
		Name:           "TCP Length Gate",
		Slug:           "awd-tcp-length-gate",
		Category:       "pwn",
		Difficulty:     model.ChallengeDifficultyMedium,
		ServiceType:    model.AWDServiceTypeBinaryTCP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusPublished,
		CheckerType:    model.AWDCheckerTypeTCPStandard,
		CheckerConfig:  `{"timeout_ms":3000,"steps":[{"send":"PING\n","expect_contains":"PONG"},{"send_template":"SET_FLAG {{FLAG}}\n","expect_contains":"OK"},{"send":"GET_FLAG\n","expect_contains":"{{FLAG}}"}]}`,
		AccessConfig:   `{"public_base_url":"tcp://preview.internal:8080","service_port":8080}`,
		RuntimeConfig:  `{"service_port":8080,"image_ref":"ctf/awd-tcp-length-gate:latest"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	accessURL, closeTCPFixture := startAWDTCPPreviewFixture(t)
	t.Cleanup(closeTCPFixture)

	checkerConfig := map[string]any{
		"timeout_ms": 3000,
		"steps": []any{
			map[string]any{"send": "PING\n", "expect_contains": "PONG"},
			map[string]any{"send_template": "SET_FLAG {{FLAG}}\n", "expect_contains": "OK"},
			map[string]any{"send": "GET_FLAG\n", "expect_contains": "{{FLAG}}"},
		},
	}
	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout: time.Second,
	})
	preview, err := service.commands.PreviewChecker(context.Background(), contestID, contestcmd.PreviewCheckerInput{
		AWDChallengeID: awdChallengeID,
		CheckerType:    string(model.AWDCheckerTypeTCPStandard),
		CheckerConfig:  checkerConfig,
		AccessURL:      accessURL,
		PreviewFlag:    "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if preview.ServiceStatus != model.AWDServiceStatusUp || preview.CheckerType != model.AWDCheckerTypeTCPStandard {
		t.Fatalf("unexpected preview response: %+v", preview)
	}
	if strings.TrimSpace(preview.PreviewToken) == "" {
		t.Fatal("expected preview token")
	}

	challengeRepo := challengeinfra.NewRepository(db)
	contestRepo := contestinfra.NewRepository(db)
	contestChallengeRepo := contestinfra.NewChallengeRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)
	contestService := contestcmd.NewContestAWDServiceService(awdRepo, contestRepo, contestChallengeRepo, challengeRepo, challengeRepo, contestinfra.NewAWDCheckerPreviewTokenStore(redisClient))
	created, err := contestService.CreateContestAWDService(context.Background(), contestID, contestcmd.CreateContestAWDServiceInput{
		AWDChallengeID:         awdChallengeID,
		Points:                 100,
		Order:                  1,
		IsVisible:              boolPtr(true),
		CheckerType:            strPtr(string(model.AWDCheckerTypeTCPStandard)),
		CheckerConfig:          checkerConfig,
		AWDCheckerPreviewToken: strPtr(preview.PreviewToken),
	})
	if err != nil {
		t.Fatalf("CreateContestAWDService() error = %v", err)
	}

	stored, err := awdRepo.FindContestAWDServiceByContestAndID(context.Background(), contestID, created.ID)
	if err != nil {
		t.Fatalf("FindContestAWDServiceByContestAndID() error = %v", err)
	}
	if stored.ValidationState != model.AWDCheckerValidationStatePassed {
		t.Fatalf("ValidationState = %s, want passed", stored.ValidationState)
	}
	readiness, err := contestqry.NewAWDService(awdRepo, contestRepo).GetReadiness(context.Background(), contestID)
	if err != nil {
		t.Fatalf("GetReadiness() error = %v", err)
	}
	if !readiness.Ready || readiness.PassedChallenges != 1 || readiness.BlockingCount != 0 {
		t.Fatalf("unexpected readiness: %+v", readiness)
	}
}

func startAWDTCPPreviewFixture(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleAWDTCPPreviewFixtureConn(conn)
		}
	}()

	closeFn := func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("tcp preview fixture did not stop")
		}
	}
	return "tcp://" + listener.Addr().String(), closeFn
}

func handleAWDTCPPreviewFixtureConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	storedFlag := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		switch {
		case line == "PING\n":
			_, _ = conn.Write([]byte("PONG\n"))
		case strings.HasPrefix(line, "SET_FLAG "):
			storedFlag = strings.TrimSpace(strings.TrimPrefix(line, "SET_FLAG "))
			_, _ = conn.Write([]byte("OK\n"))
		case line == "GET_FLAG\n":
			_, _ = conn.Write([]byte(storedFlag + "\n"))
			return
		default:
			_, _ = conn.Write([]byte("ERR\n"))
		}
	}
}

func TestAWDServicePreviewCheckerRejectsWhenRedisUnavailable(t *testing.T) {
	db := newAWDTestDB(t)
	now := time.Now()

	createAWDContestFixture(t, db, 27, now)
	createAWDChallengeFixture(t, db, 2701, now)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/flag":
			if r.Method != http.MethodGet {
				http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
				return
			}
			_, _ = w.Write([]byte("flag{preview}"))
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, nil, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	_, err := service.commands.PreviewChecker(context.Background(), 27, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2701,
		CheckerType:    string(model.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   server.URL,
		PreviewFlag: "flag{preview}",
	})
	if err != errcode.ErrAWDCheckerPreviewUnavailable {
		t.Fatalf("expected ErrAWDCheckerPreviewUnavailable, got %v", err)
	}
}

func TestAWDServicePreviewCheckerReturnsQuorumPassWhenTwoOfThreeAttemptsSucceed(t *testing.T) {
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
	createAWDContestFixture(t, db, 28, now)
	createAWDChallengeFixture(t, db, 2801, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:00Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":21}],"put_flag":{"healthy":true,"method":"PUT","path":"/api/flag"},"get_flag":{"healthy":true,"method":"GET","path":"/api/flag"}}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusDown,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"http_request_failed","checked_at":"2026-04-21T11:00:01Z","error_code":"http_request_failed","error":"connection reset by peer","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"http_request_failed","error":"connection reset by peer"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:02Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":18}],"put_flag":{"healthy":true,"method":"PUT","path":"/api/flag"},"get_flag":{"healthy":true,"method":"GET","path":"/api/flag"},"havoc":{"healthy":true,"method":"GET","path":"/healthz"}}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2801,
				},
			},
		},
	}

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	resp, err := service.PreviewChecker(context.Background(), 28, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2801,
		CheckerType:    string(model.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if roundManager.previewCalls != 3 {
		t.Fatalf("expected 3 preview attempts, got %d", roundManager.previewCalls)
	}
	if resp.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
	if resp.CheckResult["status_reason"] != "preview_quorum_passed" {
		t.Fatalf("unexpected status_reason: %#v", resp.CheckResult["status_reason"])
	}
	if resp.CheckResult["preview_pass_count"] != float64(2) {
		t.Fatalf("unexpected preview_pass_count: %#v", resp.CheckResult["preview_pass_count"])
	}
	if resp.CheckResult["preview_total_count"] != float64(3) {
		t.Fatalf("unexpected preview_total_count: %#v", resp.CheckResult["preview_total_count"])
	}
	if resp.CheckResult["preview_summary"] != "2/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", resp.CheckResult["preview_summary"])
	}
	if resp.CheckResult["check_source"] != "checker_preview" {
		t.Fatalf("unexpected check_source: %#v", resp.CheckResult["check_source"])
	}
}

func TestAWDServicePreviewCheckerBroadcastsRealtimeProgressToRequester(t *testing.T) {
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
	createAWDContestFixture(t, db, 281, now)
	createAWDChallengeFixture(t, db, 2811, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:00Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:01Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:02Z","targets":[{"access_url":"http://preview.internal","healthy":true}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2811,
				},
			},
		},
	}
	bus := platformevents.NewBus()
	received := make(chan contestcontracts.AWDPreviewProgressEvent, 8)
	bus.Subscribe(contestcontracts.EventAWDPreviewProgress, func(_ context.Context, evt platformevents.Event) error {
		payload, ok := evt.Payload.(contestcontracts.AWDPreviewProgressEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)
	service.SetEventBus(bus)

	ctx := contestcmd.WithAWDPreviewRequester(context.Background(), 9001)
	_, err = service.PreviewChecker(ctx, 281, contestcmd.PreviewCheckerInput{
		AWDChallengeID:   2811,
		CheckerType:      string(model.AWDCheckerTypeHTTPStandard),
		PreviewRequestID: "preview-progress-1",
		AccessURL:        "http://preview.internal",
		PreviewFlag:      "flag{preview}",
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}

	expectedPhases := []string{"prepare", "attempt-1", "attempt-2", "attempt-3", "summary"}
	gotEvents := make([]contestcontracts.AWDPreviewProgressEvent, 0, len(expectedPhases))
	for range expectedPhases {
		select {
		case evt := <-received:
			gotEvents = append(gotEvents, evt)
		case <-time.After(time.Second):
			t.Fatalf("expected %d realtime progress events, got %d", len(expectedPhases), len(gotEvents))
		}
	}
	for index, expectedPhase := range expectedPhases {
		evt := gotEvents[index]
		if evt.UserID != 9001 {
			t.Fatalf("unexpected broadcast user: %d", evt.UserID)
		}
		if evt.PhaseKey != expectedPhase {
			t.Fatalf("event %d unexpected phase_key: %s", index, evt.PhaseKey)
		}
		if evt.PreviewRequestID != "preview-progress-1" {
			t.Fatalf("event %d unexpected preview_request_id: %s", index, evt.PreviewRequestID)
		}
		if evt.Status != "running" {
			t.Fatalf("event %d unexpected status: %s", index, evt.Status)
		}
	}
	if gotEvents[1].Attempt != 1 {
		t.Fatalf("attempt-1 event missing attempt: %+v", gotEvents[1])
	}
	if gotEvents[3].Attempt != 3 {
		t.Fatalf("attempt-3 event missing attempt: %+v", gotEvents[3])
	}
}

func TestAWDServicePreviewCheckerReturnsQuorumFailureWhenOnlyOneAttemptSucceeds(t *testing.T) {
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
	createAWDContestFixture(t, db, 29, now)
	createAWDChallengeFixture(t, db, 2901, now)

	roundManager := &fakeAWDPreviewRoundManager{
		previewResponses: []*contestports.AWDServicePreviewResult{
			{
				ServiceStatus: model.AWDServiceStatusDown,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"http_request_failed","checked_at":"2026-04-21T11:00:00Z","error_code":"http_request_failed","error":"connection reset by peer","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"http_request_failed"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusUp,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"healthy","checked_at":"2026-04-21T11:00:01Z","targets":[{"access_url":"http://preview.internal","healthy":true,"latency_ms":20}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
			{
				ServiceStatus: model.AWDServiceStatusDown,
				CheckerType:   model.AWDCheckerTypeHTTPStandard,
				CheckResult:   `{"check_source":"checker_preview","checker_type":"http_standard","status_reason":"unexpected_http_status","checked_at":"2026-04-21T11:00:02Z","error_code":"unexpected_http_status","error":"unexpected status 502","targets":[{"access_url":"http://preview.internal","healthy":false,"error_code":"unexpected_http_status"}]}`,
				PreviewContext: contestports.AWDCheckerPreviewContext{
					AccessURL:      "http://preview.internal",
					PreviewFlag:    "flag{preview}",
					AWDChallengeID: 2901,
				},
			},
		},
	}

	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		"",
		config.ContestAWDConfig{},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	resp, err := service.PreviewChecker(context.Background(), 29, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2901,
		CheckerType:    string(model.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if roundManager.previewCalls != 3 {
		t.Fatalf("expected 3 preview attempts, got %d", roundManager.previewCalls)
	}
	if resp.ServiceStatus != model.AWDServiceStatusDown {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
	if resp.CheckResult["status_reason"] != "preview_quorum_failed" {
		t.Fatalf("unexpected status_reason: %#v", resp.CheckResult["status_reason"])
	}
	if resp.CheckResult["preview_pass_count"] != float64(1) {
		t.Fatalf("unexpected preview_pass_count: %#v", resp.CheckResult["preview_pass_count"])
	}
	if resp.CheckResult["preview_total_count"] != float64(3) {
		t.Fatalf("unexpected preview_total_count: %#v", resp.CheckResult["preview_total_count"])
	}
	if resp.CheckResult["preview_summary"] != "1/3 通过" {
		t.Fatalf("unexpected preview_summary: %#v", resp.CheckResult["preview_summary"])
	}
	if resp.CheckResult["error_code"] != "http_request_failed" {
		t.Fatalf("unexpected error_code: %#v", resp.CheckResult["error_code"])
	}
}

func TestAWDServicePreviewCheckerAcceptsServiceIDAndReturnsServiceContext(t *testing.T) {
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

	createAWDContestFixture(t, db, 25, now)
	createAWDChallengeFixture(t, db, 2501, now)
	createAWDContestChallengeFixture(t, db, 25, 2501, now)
	serviceID := defaultAWDContestServiceID(25, 2501)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	service := newAWDServiceForTest(db, redisClient, "", config.ContestAWDConfig{
		CheckerTimeout:    time.Second,
		CheckerHealthPath: "/healthz",
	})

	method := reflect.ValueOf(service.commands).MethodByName("PreviewChecker")
	if !method.IsValid() {
		t.Fatalf("PreviewChecker method not implemented")
	}

	reqValue := reflect.New(method.Type().In(2))
	setReflectedField(t, reqValue.Elem(), "ServiceID", serviceID)
	setReflectedField(t, reqValue.Elem(), "CheckerType", string(model.AWDCheckerTypeHTTPStandard))
	setReflectedField(t, reqValue.Elem(), "CheckerConfig", map[string]any{
		"get_flag": map[string]any{
			"method":             "GET",
			"path":               "/healthz",
			"expected_status":    http.StatusOK,
			"expected_substring": "",
		},
	})
	setReflectedField(t, reqValue.Elem(), "AccessURL", server.URL)

	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(int64(25)),
		reqValue.Elem(),
	})

	if len(results) != 2 {
		t.Fatalf("unexpected result count: %d", len(results))
	}
	if errValue := results[1].Interface(); errValue != nil {
		t.Fatalf("PreviewChecker() error = %v", errValue)
	}

	respValue := results[0]
	if respValue.IsNil() {
		t.Fatal("expected preview response")
	}
	resp := respValue.Elem()
	previewContext := resp.FieldByName("PreviewContext")
	if !previewContext.IsValid() {
		t.Fatal("expected preview context")
	}
	if gotServiceID := previewContext.FieldByName("ServiceID").Int(); gotServiceID != serviceID {
		t.Fatalf("unexpected preview service_id: %d", gotServiceID)
	}
	if gotChallengeID := previewContext.FieldByName("AWDChallengeID").Int(); gotChallengeID != 2501 {
		t.Fatalf("unexpected preview awd_challenge_id: %d", gotChallengeID)
	}
	previewToken := resp.FieldByName("PreviewToken")
	if !previewToken.IsValid() || strings.TrimSpace(previewToken.String()) == "" {
		t.Fatal("expected preview_token")
	}
}

func TestAWDServicePreviewCheckerStartsPreviewRuntimeWhenAccessURLMissing(t *testing.T) {
	db := newAWDTestDB(t)
	if err := db.AutoMigrate(&model.Image{}, &model.AWDChallenge{}); err != nil {
		t.Fatalf("auto migrate preview dependencies: %v", err)
	}
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
	const previewSecret = "preview-secret-12345678901234567890"
	createAWDContestFixture(t, db, 26, now)
	if err := db.Create(&model.Image{
		ID:        26001,
		Name:      "registry.example.edu/ctf/awd-preview",
		Tag:       "v1",
		Digest:    "sha256:preview-v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.AWDChallenge{
		ID:             2601,
		Name:           "Preview Target",
		Slug:           "preview-target",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyEasy,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}}`,
		RuntimeConfig:  `{"image_id":26001,"image_ref":"registry.example.edu/ctf/awd-preview:v1","checker_token_env":"CHECKER_TOKEN"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/flag" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("flag{preview}"))
	}))
	t.Cleanup(server.Close)

	runtimeProbe := &fakeContestPreviewRuntimeProbe{
		containerAccessURL: server.URL,
		containerDetails: model.InstanceRuntimeDetails{
			Containers: []model.InstanceRuntimeContainer{{ContainerID: "preview-container"}},
		},
	}
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	stateStore := contestinfra.NewAWDRoundStateStore(redisClient)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(redisClient)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		stateStore,
		previewTokenStore,
		previewSecret,
		config.ContestAWDConfig{
			CheckerTimeout:    time.Second,
			CheckerHealthPath: "/healthz",
		},
		zap.NewNop(),
		newAWDCommandRoundManagerForTest(db, redisClient, config.ContestAWDConfig{
			CheckerTimeout:    time.Second,
			CheckerHealthPath: "/healthz",
		}, previewSecret, nil, zap.NewNop()),
		imageRepo,
		awdChallengeRepo,
		runtimeProbe,
	)

	resp, err := service.PreviewChecker(context.Background(), 26, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2601,
		CheckerType:    string(model.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewChecker() error = %v", err)
	}
	if !runtimeProbe.createContainerCalled {
		t.Fatal("expected preview runtime container startup")
	}
	if !runtimeProbe.cleanupCalled {
		t.Fatal("expected preview runtime cleanup")
	}
	if runtimeProbe.lastImageName != "registry.example.edu/ctf/awd-preview@sha256:preview-v1" {
		t.Fatalf("unexpected preview image: %s", runtimeProbe.lastImageName)
	}
	if runtimeProbe.lastEnv["FLAG"] != "flag{preview}" {
		t.Fatalf("unexpected preview FLAG env: %+v", runtimeProbe.lastEnv)
	}
	if runtimeProbe.lastEnv["CHECKER_TOKEN"] != contestdomain.BuildAWDCheckerPreviewToken(26, 0, 2601, previewSecret) {
		t.Fatalf("unexpected preview CHECKER_TOKEN env: %+v", runtimeProbe.lastEnv)
	}
	if resp.PreviewContext.AccessURL != server.URL {
		t.Fatalf("unexpected preview access url: %s", resp.PreviewContext.AccessURL)
	}
	if resp.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("unexpected service status: %s", resp.ServiceStatus)
	}
}

func TestAWDServicePreviewCheckerRejectsExplicitAccessURLWhenRuntimeImageUnavailable(t *testing.T) {
	db := newAWDTestDB(t)
	if err := db.AutoMigrate(&model.AWDChallenge{}); err != nil {
		t.Fatalf("auto migrate awd challenge: %v", err)
	}
	now := time.Now()
	createAWDContestFixture(t, db, 260, now)
	if err := db.Create(&model.Image{
		ID:        26002,
		Name:      "registry.example.edu/ctf/awd-preview",
		Tag:       "pending",
		Status:    model.ImageStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.AWDChallenge{
		ID:             2602,
		Name:           "Preview Pending Image",
		Slug:           "preview-pending-image",
		Category:       "web",
		Difficulty:     model.ChallengeDifficultyEasy,
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusPublished,
		CheckerType:    model.AWDCheckerTypeHTTPStandard,
		CheckerConfig:  `{"get_flag":{"method":"GET","path":"/api/flag","expected_status":200,"expected_substring":"{{FLAG}}"}}`,
		RuntimeConfig:  `{"image_id":26002,"image_ref":"registry.example.edu/ctf/awd-preview:pending"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}

	roundManager := &fakeAWDPreviewRoundManager{}
	awdRepo := newAWDCommandRepositoryForTest(db)
	contestRepo := contestinfra.NewRepository(db)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(nil)
	imageRepo, awdChallengeRepo := newAWDPreviewRuntimeLookupsForTest(db)
	service := contestcmd.NewAWDService(
		awdRepo,
		contestRepo,
		nil,
		previewTokenStore,
		"",
		config.ContestAWDConfig{CheckerTimeout: time.Second},
		zap.NewNop(),
		roundManager,
		imageRepo,
		awdChallengeRepo,
		nil,
	)

	_, err := service.PreviewChecker(context.Background(), 260, contestcmd.PreviewCheckerInput{
		AWDChallengeID: 2602,
		CheckerType:    string(model.AWDCheckerTypeHTTPStandard),
		CheckerConfig: map[string]any{
			"get_flag": map[string]any{
				"method":             "GET",
				"path":               "/api/flag",
				"expected_status":    http.StatusOK,
				"expected_substring": "{{FLAG}}",
			},
		},
		AccessURL:   "http://preview.internal",
		PreviewFlag: "flag{preview}",
	})
	if err == nil {
		t.Fatal("expected PreviewChecker() to reject unavailable runtime image")
	}
	if len(roundManager.previewRequests) != 0 {
		t.Fatalf("preview should be blocked before checker execution, got %+v", roundManager.previewRequests)
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
		ServiceStatus: model.AWDServiceStatusUp,
		CheckResult:   map[string]any{"latency_ms": 10},
	}); err != nil {
		t.Fatalf("seed Red service check: %v", err)
	}
	if _, err := service.UpsertServiceCheck(context.Background(), 3, 31, contestcmd.UpsertServiceCheckInput{
		TeamID:        312,
		ServiceID:     serviceID,
		ServiceStatus: model.AWDServiceStatusCompromised,
		CheckResult:   map[string]any{"latency_ms": 25},
	}); err != nil {
		t.Fatalf("seed Blue service check: %v", err)
	}
	if _, err := service.UpsertServiceCheck(context.Background(), 3, 31, contestcmd.UpsertServiceCheckInput{
		TeamID:        313,
		ServiceID:     serviceID,
		ServiceStatus: model.AWDServiceStatusUp,
		CheckResult:   map[string]any{"latency_ms": 8},
	}); err != nil {
		t.Fatalf("seed Green service check: %v", err)
	}
	if err := db.Model(&model.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 311, 301).
		Updates(map[string]any{
			"sla_score":    10,
			"checker_type": model.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Red sla/checker fields: %v", err)
	}
	if err := db.Model(&model.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 312, 301).
		Updates(map[string]any{
			"sla_score":    9,
			"checker_type": model.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Blue sla/checker fields: %v", err)
	}
	if err := db.Model(&model.AWDTeamService{}).
		Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 313, 301).
		Updates(map[string]any{
			"sla_score":    8,
			"checker_type": model.AWDCheckerTypeHTTPStandard,
		}).Error; err != nil {
		t.Fatalf("seed Green sla/checker fields: %v", err)
	}

	first, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 311,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() error = %v", err)
	}
	if first.Source != model.AWDAttackSourceManual {
		t.Fatalf("expected manual attack source, got %+v", first)
	}
	if first.ScoreGained != 60 {
		t.Fatalf("expected first score gained 60, got %+v", first)
	}

	second, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 311,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() duplicate error = %v", err)
	}
	if second.ScoreGained != 0 {
		t.Fatalf("expected duplicate score gained 0, got %+v", second)
	}
	var blueService model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 31, 312, 301).First(&blueService).Error; err != nil {
		t.Fatalf("load Blue service: %v", err)
	}
	if blueService.ServiceStatus != model.AWDServiceStatusCompromised || blueService.AttackReceived != 2 || blueService.AttackScore != 60 || blueService.DefenseScore != 0 || blueService.SLAScore != 9 || blueService.CheckerType != model.AWDCheckerTypeHTTPStandard {
		t.Fatalf("unexpected Blue service impact: %+v", blueService)
	}

	if _, err := service.CreateAttackLog(context.Background(), 3, 31, contestcmd.CreateAttackLogInput{
		AttackerTeamID: 313,
		VictimTeamID:   312,
		ServiceID:      serviceID,
		AttackType:     model.AWDAttackTypeServiceExploit,
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
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  "flag{awd}",
		IsSuccess:      true,
	})
	if err != nil {
		t.Fatalf("CreateAttackLog() error = %v", err)
	}
	if resp.Source != model.AWDAttackSourceManual {
		t.Fatalf("expected manual source, got %+v", resp)
	}
	if resp.ScoreGained != 75 {
		t.Fatalf("unexpected score gained: %+v", resp)
	}
	assertAWDServiceStatusCache(t, redisClient, 6, 612, serviceID, model.AWDServiceStatusCompromised)

	var victimService model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 61, 612, 601).First(&victimService).Error; err != nil {
		t.Fatalf("load victim service: %v", err)
	}
	if victimService.ServiceStatus != model.AWDServiceStatusCompromised || victimService.AttackReceived != 1 || victimService.AttackScore != 75 || victimService.DefenseScore != 0 {
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
		ServiceStatus: model.AWDServiceStatusDown,
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

	if err := db.Model(&model.Contest{}).Where("id = ?", 17).Updates(map[string]any{
		"status":   model.ContestStatusEnded,
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
		ServiceStatus: model.AWDServiceStatusUp,
		CheckResult:   map[string]any{"reason": "postmortem-fix"},
	}); err != nil {
		t.Fatalf("ended contest UpsertServiceCheck() error = %v", err)
	}

	assertAWDServiceStatusCacheMissing(t, redisClient, 17, 1711, serviceID)
}

func TestAWDServiceSubmitAttackUsesCurrentRoundFlagAndDeduplicatesByTeam(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 4, now)
	createAWDRoundFixture(t, db, 41, 4, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 401, now)
	createAWDContestChallengeFixture(t, db, 4, 401, now)
	createAWDTeamFixture(t, db, 411, 4, "Red", now)
	createAWDTeamFixture(t, db, 412, 4, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 4, 411, 4001, now)
	createContestRegistrationForExistingTeam(t, db, 4, 411, 4002, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 401).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(4), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	serviceID := defaultAWDContestServiceID(4, 401)
	flag := contestdomain.BuildAWDRoundFlag(4, 1, 412, 401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(4, 41), map[string]any{
		rediskeys.AWDRoundFlagServiceField(412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	first, err := service.SubmitAttack(context.Background(), 4001, 4, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() first error = %v", err)
	}
	if first.Source != model.AWDAttackSourceSubmission {
		t.Fatalf("expected submission source, got %+v", first)
	}
	if !first.IsSuccess || first.ScoreGained != 80 || first.AttackerTeamID != 411 || first.VictimTeamID != 412 {
		t.Fatalf("unexpected first attack resp: %+v", first)
	}

	second, err := service.SubmitAttack(context.Background(), 4002, 4, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() second error = %v", err)
	}
	if second.Source != model.AWDAttackSourceSubmission {
		t.Fatalf("expected submission source, got %+v", second)
	}
	if !second.IsSuccess || second.ScoreGained != 0 {
		t.Fatalf("unexpected second attack resp: %+v", second)
	}

	var logs []model.AWDAttackLog
	if err := db.Order("id ASC").Find(&logs).Error; err != nil {
		t.Fatalf("query attack logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 attack logs, got %+v", logs)
	}
	if logs[0].SubmittedByUserID == nil || *logs[0].SubmittedByUserID != 4001 {
		t.Fatalf("expected first log submitted_by_user_id=4001, got %+v", logs[0])
	}
	if logs[1].SubmittedByUserID == nil || *logs[1].SubmittedByUserID != 4002 {
		t.Fatalf("expected second log submitted_by_user_id=4002, got %+v", logs[1])
	}
}

func TestAWDServiceSubmitAttackAcceptsServiceScopedRoundFlagField(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 24, now)
	createAWDRoundFixture(t, db, 241, 24, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 2401, now)
	createAWDContestChallengeFixture(t, db, 24, 2401, now)
	createAWDTeamFixture(t, db, 2411, 24, "Red", now)
	createAWDTeamFixture(t, db, 2412, 24, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 24, 2411, 24001, now)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 2401).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	serviceID := defaultAWDContestServiceID(24, 2401)
	if err := db.Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND awd_challenge_id = ?", 24, 2401).
		Updates(map[string]any{
			"display_name":   "Bank Portal",
			"order":          0,
			"is_visible":     true,
			"score_config":   `{"points":100,"awd_sla_score":1,"awd_defense_score":2}`,
			"runtime_config": `{"awd_challenge_id":2401,"checker_type":"legacy_probe","checker_config":{}}`,
			"updated_at":     now,
		}).Error; err != nil {
		t.Fatalf("update contest awd service: %v", err)
	}

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(24), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	flag := contestdomain.BuildAWDRoundFlag(24, 1, 2412, 2401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(24, 241), map[string]any{
		rediskeys.AWDRoundFlagServiceField(2412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set service scoped round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 24001, 24, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 2412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() service scoped flag error = %v", err)
	}
	if resp.Source != model.AWDAttackSourceSubmission || !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected service scoped submit resp: %+v", resp)
	}

	var logRecord model.AWDAttackLog
	if err := db.Where("round_id = ? AND attacker_team_id = ? AND victim_team_id = ?", 241, 2411, 2412).First(&logRecord).Error; err != nil {
		t.Fatalf("load service scoped attack log: %v", err)
	}
	if logRecord.ServiceID != serviceID {
		t.Fatalf("expected attack log service_id=%d, got %+v", serviceID, logRecord)
	}

	var victimService model.AWDTeamService
	if err := db.Where("round_id = ? AND team_id = ? AND awd_challenge_id = ?", 241, 2412, 2401).First(&victimService).Error; err != nil {
		t.Fatalf("load victim service after service scoped submit: %v", err)
	}
	if victimService.ServiceID != serviceID {
		t.Fatalf("expected victim service service_id=%d, got %+v", serviceID, victimService)
	}
}

func TestAWDServiceSubmitAttackPublishesAttackAcceptedEvent(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	bus := platformevents.NewBus()
	service.commands.SetEventBus(bus)

	now := time.Now()
	createAWDContestFixture(t, db, 14, now)
	createAWDRoundFixture(t, db, 141, 14, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 1401, now)
	createAWDContestChallengeFixture(t, db, 14, 1401, now)
	createAWDTeamFixture(t, db, 1411, 14, "Red", now)
	createAWDTeamFixture(t, db, 1412, 14, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 14, 1411, 14001, now)
	createContestRegistrationForExistingTeam(t, db, 14, 1411, 14002, now)
	serviceID := defaultAWDContestServiceID(14, 1401)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 1401).Updates(map[string]any{
		"flag_prefix": "awd",
		"category":    model.DimensionWeb,
	}).Error; err != nil {
		t.Fatalf("update challenge fields: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(14), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	flag := contestdomain.BuildAWDRoundFlag(14, 1, 1412, 1401, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(14, 141), map[string]any{
		rediskeys.AWDRoundFlagServiceField(1412, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	received := make(chan contestcontracts.AWDAttackAcceptedEvent, 2)
	bus.Subscribe(contestcontracts.EventAWDAttackAccepted, func(_ context.Context, evt platformevents.Event) error {
		payload, ok := evt.Payload.(contestcontracts.AWDAttackAcceptedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})

	first, err := service.SubmitAttack(context.Background(), 14001, 14, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() first error = %v", err)
	}
	if !first.IsSuccess || first.ScoreGained != 80 {
		t.Fatalf("unexpected first attack resp: %+v", first)
	}

	second, err := service.SubmitAttack(context.Background(), 14002, 14, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1412,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() second error = %v", err)
	}
	if !second.IsSuccess || second.ScoreGained != 0 {
		t.Fatalf("unexpected second attack resp: %+v", second)
	}

	select {
	case evt := <-received:
		if evt.UserID != 14001 || evt.AWDChallengeID != 1401 || evt.Dimension != model.DimensionWeb {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected contest.awd.attack_accepted event to be published")
	}

	select {
	case evt := <-received:
		t.Fatalf("expected only one accepted event, got %+v", evt)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestAWDServiceSubmitAttackAcceptsPreviousRoundFlagWithinGrace(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		PreviousRoundGrace: time.Minute,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 5, now)
	createAWDRoundFixtureWithWindow(t, db, 51, 5, 1, 80, 20, now.Add(-5*time.Minute), now.Add(-10*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 52, 5, 2, 80, 20, now.Add(-10*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 501, now)
	createAWDContestChallengeFixture(t, db, 5, 501, now)
	createAWDTeamFixture(t, db, 511, 5, "Red", now)
	createAWDTeamFixture(t, db, 512, 5, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 5, 511, 5001, now)
	serviceID := defaultAWDContestServiceID(5, 501)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 501).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(5), "2", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}
	currentFlag := contestdomain.BuildAWDRoundFlag(5, 2, 512, 501, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(5, 52), map[string]any{
		rediskeys.AWDRoundFlagServiceField(512, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set current round flag: %v", err)
	}

	previousFlag := contestdomain.BuildAWDRoundFlag(5, 1, 512, 501, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 5001, 5, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 512,
		Flag:         previousFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() previous round flag error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected previous round submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackAllowsFrozenContest(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{})
	now := time.Now()

	createAWDContestFixture(t, db, 6, now)
	createAWDRoundFixture(t, db, 61, 6, 1, 80, 20, now)
	createAWDChallengeFixture(t, db, 601, now)
	createAWDContestChallengeFixture(t, db, 6, 601, now)
	createAWDTeamFixture(t, db, 611, 6, "Red", now)
	createAWDTeamFixture(t, db, 612, 6, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 6, 611, 6001, now)
	serviceID := defaultAWDContestServiceID(6, 601)

	if err := db.Model(&model.Contest{}).Where("id = ?", 6).Update("status", model.ContestStatusFrozen).Error; err != nil {
		t.Fatalf("set contest frozen: %v", err)
	}
	if err := db.Model(&model.Challenge{}).Where("id = ?", 601).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(6), "1", 0).Err(); err != nil {
		t.Fatalf("set current round: %v", err)
	}

	flag := contestdomain.BuildAWDRoundFlag(6, 1, 612, 601, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(6, 61), map[string]any{
		rediskeys.AWDRoundFlagServiceField(612, serviceID): flag,
	}).Err(); err != nil {
		t.Fatalf("set round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 6001, 6, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 612,
		Flag:         flag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() frozen contest error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected frozen contest submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackIgnoresStaleCurrentRoundPointer(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 7, now)
	createAWDRoundFixtureWithWindow(t, db, 71, 7, 1, 80, 20, now.Add(-5*time.Minute), now.Add(-10*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 72, 7, 2, 80, 20, now.Add(-10*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 701, now)
	createAWDContestChallengeFixture(t, db, 7, 701, now)
	createAWDTeamFixture(t, db, 711, 7, "Red", now)
	createAWDTeamFixture(t, db, 712, 7, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 7, 711, 7001, now)
	serviceID := defaultAWDContestServiceID(7, 701)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 701).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(7), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(7, 2, 712, 701, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(7, 72), map[string]any{
		rediskeys.AWDRoundFlagServiceField(712, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set current round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 7001, 7, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 712,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with stale current round pointer error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 || resp.RoundID != 72 {
		t.Fatalf("unexpected stale pointer submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackUsesTimeDerivedCurrentRoundWhenRoundStatusLags(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 8, now)
	if err := db.Model(&model.Contest{}).Where("id = ?", 8).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 81, 8, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDRoundFixtureWithWindow(t, db, 82, 8, 2, 80, 20, now.Add(-30*time.Second), time.Time{})
	createAWDChallengeFixture(t, db, 801, now)
	createAWDContestChallengeFixture(t, db, 8, 801, now)
	createAWDTeamFixture(t, db, 811, 8, "Red", now)
	createAWDTeamFixture(t, db, 812, 8, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 8, 811, 8001, now)
	serviceID := defaultAWDContestServiceID(8, 801)

	if err := db.Model(&model.AWDRound{}).Where("id = ?", 81).Updates(map[string]any{
		"status":   model.AWDRoundStatusRunning,
		"ended_at": nil,
	}).Error; err != nil {
		t.Fatalf("mark stale round as running: %v", err)
	}
	if err := db.Model(&model.AWDRound{}).Where("id = ?", 82).Updates(map[string]any{
		"status": model.AWDRoundStatusPending,
	}).Error; err != nil {
		t.Fatalf("mark actual round as pending: %v", err)
	}
	if err := db.Model(&model.Challenge{}).Where("id = ?", 801).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(8), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(8, 2, 812, 801, "awd-secret", "awd")
	if err := redisClient.HSet(context.Background(), rediskeys.AWDRoundFlagsKey(8, 82), map[string]any{
		rediskeys.AWDRoundFlagServiceField(812, serviceID): currentFlag,
	}).Err(); err != nil {
		t.Fatalf("set actual round flag: %v", err)
	}

	resp, err := service.SubmitAttack(context.Background(), 8001, 8, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 812,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with lagging round status error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 || resp.RoundID != 82 {
		t.Fatalf("unexpected lagging status submit resp: %+v", resp)
	}
}

func TestAWDServiceSubmitAttackRejectsPreviousFlagAfterMaterializingMissingCurrentRound(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 9, now)
	if err := db.Model(&model.Contest{}).Where("id = ?", 9).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 91, 9, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDChallengeFixture(t, db, 901, now)
	createAWDContestChallengeFixture(t, db, 9, 901, now)
	createAWDTeamFixture(t, db, 911, 9, "Red", now)
	createAWDTeamFixture(t, db, 912, 9, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 9, 911, 9001, now)
	serviceID := defaultAWDContestServiceID(9, 901)

	if err := db.Model(&model.AWDRound{}).Where("id = ?", 91).Updates(map[string]any{
		"status":   model.AWDRoundStatusRunning,
		"ended_at": nil,
	}).Error; err != nil {
		t.Fatalf("mark stale round as running: %v", err)
	}
	if err := db.Model(&model.Challenge{}).Where("id = ?", 901).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}
	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(9), "1", 0).Err(); err != nil {
		t.Fatalf("set stale current round: %v", err)
	}

	previousFlag := contestdomain.BuildAWDRoundFlag(9, 1, 912, 901, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 9001, 9, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 912,
		Flag:         previousFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with previous flag after materializing round error = %v", err)
	}
	if resp.IsSuccess || resp.ScoreGained != 0 {
		t.Fatalf("unexpected stale flag submit resp: %+v", resp)
	}

	var round model.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 9, 2).First(&round).Error; err != nil {
		t.Fatalf("find materialized round: %v", err)
	}
	if resp.RoundID != round.ID {
		t.Fatalf("unexpected materialized round id for stale flag submit: resp=%d round=%d", resp.RoundID, round.ID)
	}
}

func TestAWDServiceSubmitAttackMaterializesMissingCurrentRound(t *testing.T) {
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

	service := newAWDServiceForTest(db, redisClient, "awd-secret", config.ContestAWDConfig{
		RoundInterval:      time.Minute,
		PreviousRoundGrace: 0,
	})
	now := time.Now()

	createAWDContestFixture(t, db, 10, now)
	if err := db.Model(&model.Contest{}).Where("id = ?", 10).Updates(map[string]any{
		"start_time": now.Add(-90 * time.Second),
		"end_time":   now.Add(90 * time.Second),
	}).Error; err != nil {
		t.Fatalf("update contest timing: %v", err)
	}
	createAWDRoundFixtureWithWindow(t, db, 101, 10, 1, 80, 20, now.Add(-90*time.Second), now.Add(-30*time.Second))
	createAWDChallengeFixture(t, db, 1001, now)
	createAWDContestChallengeFixture(t, db, 10, 1001, now)
	createAWDTeamFixture(t, db, 1011, 10, "Red", now)
	createAWDTeamFixture(t, db, 1012, 10, "Blue", now)
	createContestRegistrationForExistingTeam(t, db, 10, 1011, 10001, now)
	serviceID := defaultAWDContestServiceID(10, 1001)

	if err := db.Model(&model.Challenge{}).Where("id = ?", 1001).Update("flag_prefix", "awd").Error; err != nil {
		t.Fatalf("set flag prefix: %v", err)
	}

	currentFlag := contestdomain.BuildAWDRoundFlag(10, 2, 1012, 1001, "awd-secret", "awd")
	resp, err := service.SubmitAttack(context.Background(), 10001, 10, serviceID, contestcmd.SubmitAttackInput{
		VictimTeamID: 1012,
		Flag:         currentFlag,
	})
	if err != nil {
		t.Fatalf("SubmitAttack() with missing current round error = %v", err)
	}
	if !resp.IsSuccess || resp.ScoreGained != 80 {
		t.Fatalf("unexpected materialized round submit resp: %+v", resp)
	}

	var round model.AWDRound
	if err := db.Where("contest_id = ? AND round_number = ?", 10, 2).First(&round).Error; err != nil {
		t.Fatalf("find materialized round: %v", err)
	}
	if resp.RoundID != round.ID {
		t.Fatalf("unexpected materialized round id: resp=%d round=%d", resp.RoundID, round.ID)
	}
	if round.AttackScore != 80 || round.DefenseScore != 20 {
		t.Fatalf("unexpected materialized round score: %+v", round)
	}

	currentRound, err := redisClient.Get(context.Background(), rediskeys.AWDCurrentRoundKey(10)).Result()
	if err != nil {
		t.Fatalf("load current round key: %v", err)
	}
	if currentRound != "2" {
		t.Fatalf("unexpected current round key: %s", currentRound)
	}

	flagValue, err := redisClient.HGet(
		context.Background(),
		rediskeys.AWDRoundFlagsKey(10, round.ID),
		rediskeys.AWDRoundFlagServiceField(1012, serviceID),
	).Result()
	if err != nil {
		t.Fatalf("load materialized round flag: %v", err)
	}
	if flagValue != currentFlag {
		t.Fatalf("unexpected materialized round flag: got %q want %q", flagValue, currentFlag)
	}
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

	if err := db.Create(&model.Instance{
		ID:          9301,
		UserID:      9202,
		ContestID:   int64Ptr(90),
		TeamID:      int64Ptr(9102),
		ChallengeID: 9001,
		ServiceID:   awdServiceIDPtr(90, 9001),
		ContainerID: "ctr-blue-web",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://blue-web.local",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create blue instance: %v", err)
	}
	if err := db.Create(&model.Instance{
		ID:          9302,
		UserID:      9201,
		ContestID:   int64Ptr(90),
		TeamID:      int64Ptr(9101),
		ChallengeID: 9002,
		ServiceID:   awdServiceIDPtr(90, 9002),
		ContainerID: "ctr-red-pwn",
		Status:      model.InstanceStatusRunning,
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

	if err := db.Create(&model.Instance{
		ID:          9311,
		UserID:      9212,
		ContestID:   int64Ptr(91),
		TeamID:      int64Ptr(9112),
		ChallengeID: 91001,
		ServiceID:   awdServiceIDPtr(91, 91001),
		ContainerID: "ctr-beta-web",
		Status:      model.InstanceStatusRunning,
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

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrAWDReadinessBlocked.Code {
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

	if err := db.Create(&model.AWDTrafficEvent{
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
		Source:         model.AWDTrafficSourceRuntimeProxy,
		CreatedAt:      createdAt,
	}).Error; err != nil {
		t.Fatalf("create awd traffic event: %v", err)
	}
}
