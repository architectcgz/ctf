package commands_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
)

func TestServiceStartContestChallengeRejectsAWDContest(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1001, 2001, now)
	seedContestInstanceAWDContest(t, db, 3001, 2001, now)
	seedContestInstanceTeam(t, db, 3001, 4001, 5001, now)
	seedContestInstanceRegistration(t, db, 3001, 5001, 4001, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceRegistration(t, db, 3001, 5002, 4001, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3001, 4001, 5001, now)
	seedContestInstanceTeamMember(t, db, 3001, 4001, 5002, now)

	service := newContestInstanceTestService(t, db)

	resp, err := service.StartContestChallenge(context.Background(), 5001, 3001, 2001)
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected awd contest challenge entry rejected, resp=%+v err=%v", resp, err)
	}
}

func TestServiceStartContestChallengeAWDDoesNotReuseExistingTeamInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1002, 2002, now)
	seedContestInstanceAWDContest(t, db, 3002, 2002, now)
	seedContestInstanceTeam(t, db, 3002, 4002, 5003, now)
	seedContestInstanceRegistration(t, db, 3002, 5003, 4002, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceRegistration(t, db, 3002, 5004, 4002, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3002, 4002, 5003, now)
	seedContestInstanceTeamMember(t, db, 3002, 4002, 5004, now)

	contestID := int64(3002)
	teamID := int64(4002)
	if err := db.Create(&model.Instance{
		ID:          9002,
		UserID:      5003,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2002,
		ShareScope:  model.InstanceSharingPerTeam,
		ContainerID: "existing-team-instance",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30001",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("seed shared instance: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartContestChallenge(context.Background(), 5004, 3002, 2002)
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected awd contest challenge entry rejected even with existing instance, resp=%+v err=%v", resp, err)
	}
}

func TestServiceStartContestAWDServiceResolvesServiceIDAndReusesTeamInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1003, 2003, now)
	seedContestInstanceAWDContest(t, db, 3003, 2003, now)
	seedContestInstanceAWDService(t, db, 7003003, 3003, 2003, now)
	seedContestInstanceTeam(t, db, 3003, 4003, 5005, now)
	seedContestInstanceRegistration(t, db, 3003, 5005, 4003, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceRegistration(t, db, 3003, 5006, 4003, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3003, 4003, 5005, now)
	seedContestInstanceTeamMember(t, db, 3003, 4003, 5006, now)
	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	service := newContestInstanceTestService(t, db)

	first, err := service.StartContestAWDService(context.Background(), 5005, 3003, 7003003)
	if err != nil {
		t.Fatalf("StartContestAWDService() first error = %v", err)
	}
	second, err := service.StartContestAWDService(context.Background(), 5006, 3003, 7003003)
	if err != nil {
		t.Fatalf("StartContestAWDService() second error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected shared instance reuse via awd service id, got first=%d second=%d", first.ID, second.ID)
	}
	if first.ChallengeID != 2003 {
		t.Fatalf("expected awd service to resolve challenge 2003, got %+v", first)
	}
}

func TestServiceStartContestAWDServicePersistsServiceIDOnInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	seedContestInstanceChallenge(t, db, 1004, 2004, now)
	seedContestInstanceAWDContest(t, db, 3004, 2004, now)
	seedContestInstanceAWDService(t, db, 7003004, 3004, 2004, now)
	seedContestInstanceTeam(t, db, 3004, 4004, 5007, now)
	seedContestInstanceRegistration(t, db, 3004, 5007, 4004, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3004, 4004, 5007, now)

	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartContestAWDService(context.Background(), 5007, 3004, 7003004)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}

	var instance model.Instance
	if err := db.First(&instance, resp.ID).Error; err != nil {
		t.Fatalf("load persisted instance: %v", err)
	}
	if instance.ServiceID == nil || *instance.ServiceID != 7003004 {
		t.Fatalf("expected instance.service_id=7003004, got %+v", instance.ServiceID)
	}
	expectedExpiry := now.Add(time.Hour)
	if !instance.ExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %s", expectedExpiry, instance.ExpiresAt)
	}
}

func TestServiceStartAdminContestAWDTeamServiceDoesNotRequireAdminRegistration(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1005, 2005, now)
	seedContestInstanceAWDContest(t, db, 3005, 2005, now)
	seedContestInstanceAWDService(t, db, 7003005, 3005, 2005, now)
	seedContestInstanceTeam(t, db, 3005, 4005, 5008, now)
	seedContestInstanceTeamMember(t, db, 3005, 4005, 5008, now)

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartAdminContestAWDTeamService(context.Background(), 3005, 4005, 7003005)
	if err != nil {
		t.Fatalf("StartAdminContestAWDTeamService() error = %v", err)
	}
	if resp.TeamID != 4005 || resp.ServiceID != 7003005 || resp.Instance == nil {
		t.Fatalf("unexpected admin awd instance response: %+v", resp)
	}

	var instance model.Instance
	if err := db.First(&instance, resp.Instance.ID).Error; err != nil {
		t.Fatalf("load admin-started instance: %v", err)
	}
	if instance.TeamID == nil || *instance.TeamID != 4005 {
		t.Fatalf("expected instance team_id=4005, got %+v", instance.TeamID)
	}
	if instance.ServiceID == nil || *instance.ServiceID != 7003005 {
		t.Fatalf("expected instance service_id=7003005, got %+v", instance.ServiceID)
	}
	if instance.UserID != 5008 {
		t.Fatalf("expected team captain to own runtime instance, got user_id=%d", instance.UserID)
	}
	if instance.ShareScope != model.InstanceSharingPerTeam {
		t.Fatalf("expected per-team share scope, got %s", instance.ShareScope)
	}
}

func TestServiceStartAdminContestAWDTeamServiceAllowsRegistrationForPrewarmRetry(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	seedContestInstanceChallenge(t, db, 10055, 20055, now)
	seedContestInstanceAWDContestWithStatus(t, db, 30055, 20055, model.ContestStatusRegistration, now)
	seedContestInstanceAWDService(t, db, 7003055, 30055, 20055, now)
	seedContestInstanceTeam(t, db, 30055, 40055, 50055, now)
	seedContestInstanceTeamMember(t, db, 30055, 40055, 50055, now)

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartAdminContestAWDTeamService(context.Background(), 30055, 40055, 7003055)
	if err != nil {
		t.Fatalf("StartAdminContestAWDTeamService() during registration error = %v", err)
	}
	if resp == nil || resp.Instance == nil {
		t.Fatalf("expected registration prewarm retry to return instance, got %+v", resp)
	}
}

func TestServicePrewarmAdminContestAWDInstancesStartsVisibleServicesForSelectedTeam(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	seedContestInstanceChallenge(t, db, 10100, 20100, now)
	seedContestInstanceChallenge(t, db, 10101, 20101, now)
	seedContestInstanceChallenge(t, db, 10102, 20102, now)
	seedContestInstanceAWDContestWithStatus(t, db, 30100, 20100, model.ContestStatusRegistration, now)
	seedContestInstanceAWDService(t, db, 70100, 30100, 20100, now)
	seedContestInstanceAWDService(t, db, 70101, 30100, 20101, now)
	seedContestInstanceAWDServiceWithVisibility(t, db, 70102, 30100, 20102, false, now)
	seedContestInstanceTeam(t, db, 30100, 40100, 50100, now)
	seedContestInstanceTeamMember(t, db, 30100, 40100, 50100, now)

	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	teamID := int64(40100)
	resp, err := service.PrewarmAdminContestAWDInstances(context.Background(), 30100, &dto.PrewarmAdminContestAWDInstancesReq{
		TeamID: &teamID,
	})
	if err != nil {
		t.Fatalf("PrewarmAdminContestAWDInstances() error = %v", err)
	}
	if resp.Summary.Total != 2 || resp.Summary.Started != 2 || resp.Summary.Reused != 0 || resp.Summary.Failed != 0 {
		t.Fatalf("unexpected prewarm summary: %+v", resp.Summary)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 visible service results, got %+v", resp.Results)
	}
	for _, item := range resp.Results {
		if item.TeamID != 40100 {
			t.Fatalf("expected selected team only, got %+v", item)
		}
		if item.ServiceID == 70102 {
			t.Fatalf("hidden service should not be prewarmed, got %+v", item)
		}
		if item.Outcome != "started" || item.Instance == nil {
			t.Fatalf("expected started instance result, got %+v", item)
		}
	}
}

func TestServicePrewarmAdminContestAWDInstancesReturnsReusedAndFailedResults(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	seedContestInstanceChallenge(t, db, 10110, 20110, now)
	seedContestInstanceAWDContestWithStatus(t, db, 30110, 20110, model.ContestStatusRegistration, now)
	seedContestInstanceAWDService(t, db, 70110, 30110, 20110, now)
	seedContestInstanceTeam(t, db, 30110, 40110, 50110, now)
	seedContestInstanceTeamMember(t, db, 30110, 40110, 50110, now)

	if err := db.Create(&model.Team{
		ID:         40111,
		ContestID:  30110,
		Name:       "Broken Team",
		CaptainID:  0,
		InviteCode: "broken",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create broken team: %v", err)
	}

	if err := ensureContestInstanceServiceIDColumn(db); err != nil {
		t.Fatalf("ensure instances.service_id column: %v", err)
	}

	contestID := int64(30110)
	teamID := int64(40110)
	serviceID := int64(70110)
	if err := db.Create(&model.Instance{
		ID:          90110,
		UserID:      50110,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 20110,
		ServiceID:   &serviceID,
		ShareScope:  model.InstanceSharingPerTeam,
		ContainerID: "existing-team-instance",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30110",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("seed shared instance: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.PrewarmAdminContestAWDInstances(context.Background(), 30110, &dto.PrewarmAdminContestAWDInstancesReq{})
	if err != nil {
		t.Fatalf("PrewarmAdminContestAWDInstances() error = %v", err)
	}
	if resp.Summary.Total != 2 || resp.Summary.Started != 0 || resp.Summary.Reused != 1 || resp.Summary.Failed != 1 {
		t.Fatalf("unexpected prewarm summary: %+v results=%#v", resp.Summary, resp.Results)
	}

	outcomes := make(map[int64]string, len(resp.Results))
	for _, item := range resp.Results {
		outcomes[item.TeamID] = item.Outcome
		if item.TeamID == 40110 && item.Instance == nil {
			t.Fatalf("expected reused instance for valid team, got %+v", item)
		}
		if item.TeamID == 40111 && strings.TrimSpace(item.ErrorMessage) == "" {
			t.Fatalf("expected failure message for broken team, got %+v", item)
		}
	}
	if outcomes[40110] != "reused" || outcomes[40111] != "failed" {
		t.Fatalf("unexpected outcomes: %+v", outcomes)
	}
}

func TestServicePrewarmAdminContestAWDInstancesRejectsNonRegistrationContest(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now().UTC()

	seedContestInstanceChallenge(t, db, 10120, 20120, now)
	seedContestInstanceAWDContestWithStatus(t, db, 30120, 20120, model.ContestStatusRunning, now)
	seedContestInstanceAWDService(t, db, 70120, 30120, 20120, now)
	seedContestInstanceTeam(t, db, 30120, 40120, 50120, now)
	seedContestInstanceTeamMember(t, db, 30120, 40120, 50120, now)

	service := newContestInstanceTestService(t, db)
	resp, err := service.PrewarmAdminContestAWDInstances(context.Background(), 30120, &dto.PrewarmAdminContestAWDInstancesReq{})
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected non-registration contest prewarm rejected, resp=%+v err=%v", resp, err)
	}
}

func TestServiceGetContestAWDInstanceOrchestrationReturnsTeamServiceMatrix(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1006, 2006, now)
	seedContestInstanceAWDContest(t, db, 3006, 2006, now)
	seedContestInstanceAWDService(t, db, 7003006, 3006, 2006, now)
	seedContestInstanceTeam(t, db, 3006, 4006, 5009, now)
	seedContestInstanceTeamMember(t, db, 3006, 4006, 5009, now)

	service := newContestInstanceTestService(t, db)
	started, err := service.StartAdminContestAWDTeamService(context.Background(), 3006, 4006, 7003006)
	if err != nil {
		t.Fatalf("StartAdminContestAWDTeamService() error = %v", err)
	}

	resp, err := service.GetContestAWDInstanceOrchestration(context.Background(), 3006)
	if err != nil {
		t.Fatalf("GetContestAWDInstanceOrchestration() error = %v", err)
	}
	if len(resp.Teams) != 1 || resp.Teams[0].TeamID != 4006 {
		t.Fatalf("expected one team in orchestration, got %+v", resp.Teams)
	}
	if len(resp.Services) != 1 || resp.Services[0].ServiceID != 7003006 {
		t.Fatalf("expected one service in orchestration, got %+v", resp.Services)
	}
	if len(resp.Instances) != 1 || resp.Instances[0].Instance == nil || resp.Instances[0].Instance.ID != started.Instance.ID {
		t.Fatalf("expected started instance in orchestration, got %+v", resp.Instances)
	}
}

func TestServiceStartChallengeSharedReusesPracticeInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceUser(t, db, 5101, now)
	seedContestInstanceUser(t, db, 5102, now)
	seedContestInstanceChallenge(t, db, 1101, 2101, now)
	if err := db.Model(&model.Challenge{}).
		Where("id = ?", 2101).
		Update("instance_sharing", model.InstanceSharingShared).Error; err != nil {
		t.Fatalf("update challenge sharing: %v", err)
	}

	service := newContestInstanceTestService(t, db)

	first, err := service.StartChallenge(context.Background(), 5101, 2101)
	if err != nil {
		t.Fatalf("StartChallenge() first error = %v", err)
	}
	second, err := service.StartChallenge(context.Background(), 5102, 2101)
	if err != nil {
		t.Fatalf("StartChallenge() second error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected shared practice instance reuse, got first=%d second=%d", first.ID, second.ID)
	}
}

func TestServiceStartChallengeSharedReusesPracticeInstanceAndRefreshesExpiry(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceUser(t, db, 5201, now)
	seedContestInstanceUser(t, db, 5202, now)
	seedContestInstanceChallenge(t, db, 1201, 2201, now)
	if err := db.Model(&model.Challenge{}).
		Where("id = ?", 2201).
		Update("instance_sharing", model.InstanceSharingShared).Error; err != nil {
		t.Fatalf("update challenge sharing: %v", err)
	}

	originalExpiry := now.Add(5 * time.Minute)
	if err := db.Create(&model.Instance{
		ID:          9201,
		UserID:      5201,
		ChallengeID: 2201,
		ShareScope:  model.InstanceSharingShared,
		ContainerID: "shared-practice-instance",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30009",
		ExpiresAt:   originalExpiry,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("seed shared instance: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	resp, err := service.StartChallenge(context.Background(), 5202, 2201)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.ID != 9201 {
		t.Fatalf("expected shared instance reuse, got %+v", resp)
	}

	var instance model.Instance
	if err := db.First(&instance, 9201).Error; err != nil {
		t.Fatalf("load reused instance: %v", err)
	}
	if !instance.ExpiresAt.After(originalExpiry) {
		t.Fatalf("expected shared instance expiry to be refreshed, before=%s after=%s", originalExpiry, instance.ExpiresAt)
	}
}

func TestServiceStartContestChallengePerTeamReusesTeamInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceUser(t, db, 5103, now)
	seedContestInstanceUser(t, db, 5104, now)
	seedContestInstanceChallenge(t, db, 1102, 2102, now)
	if err := db.Model(&model.Challenge{}).
		Where("id = ?", 2102).
		Update("instance_sharing", model.InstanceSharingPerTeam).Error; err != nil {
		t.Fatalf("update challenge sharing: %v", err)
	}
	seedContestInstanceJeopardyContest(t, db, 3102, 2102, now)
	seedContestInstanceTeam(t, db, 3102, 4102, 5103, now)
	seedContestInstanceRegistration(t, db, 3102, 5103, 4102, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceRegistration(t, db, 3102, 5104, 4102, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3102, 4102, 5103, now)
	seedContestInstanceTeamMember(t, db, 3102, 4102, 5104, now)

	service := newContestInstanceTestService(t, db)

	first, err := service.StartContestChallenge(context.Background(), 5103, 3102, 2102)
	if err != nil {
		t.Fatalf("StartContestChallenge() first error = %v", err)
	}
	second, err := service.StartContestChallenge(context.Background(), 5104, 3102, 2102)
	if err != nil {
		t.Fatalf("StartContestChallenge() second error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected per-team instance reuse, got first=%d second=%d", first.ID, second.ID)
	}
}

func TestServiceStartChallengeRejectsNoTargetChallenge(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	if err := db.Create(&model.Challenge{
		ID:         2201,
		Title:      "No Target",
		Category:   model.DimensionMisc,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     20,
		ImageID:    0,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := newContestInstanceTestService(t, db)
	_, err := service.StartChallenge(context.Background(), 5001, 2201)
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for no-target challenge, got %v", err)
	}
}

func newContestInstanceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Image{},
		&model.Challenge{},
		&model.ChallengeTopology{},
		&model.Contest{},
		&model.ContestAWDService{},
		&model.ContestChallenge{},
		&model.ContestRegistration{},
		&model.Team{},
		&model.TeamMember{},
		&model.Instance{},
		&model.AWDServiceOperation{},
		&model.AWDScopeControl{},
		&model.AWDDefenseWorkspace{},
		&model.PortAllocation{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("auto migrate contest instance test schema: %v", err)
	}
	return db
}

func ensureContestInstanceServiceIDColumn(db *gorm.DB) error {
	if db.Migrator().HasColumn(&model.Instance{}, "service_id") {
		return nil
	}
	return db.Exec("ALTER TABLE instances ADD COLUMN service_id integer").Error
}

type contestInstanceTestRuntimeService struct{}

func (contestInstanceTestRuntimeService) CleanupRuntime(context.Context, *model.Instance) error {
	return nil
}

func (contestInstanceTestRuntimeService) CreateTopology(_ context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if req == nil {
		return nil, nil
	}
	hostPort := req.ReservedHostPort
	accessURL := fmt.Sprintf("http://127.0.0.1:%d", hostPort)
	if req.DisableEntryPortPublishing {
		for _, node := range req.Nodes {
			if !node.IsEntryPoint || len(node.NetworkAliases) == 0 {
				continue
			}
			servicePort := node.ServicePort
			if servicePort <= 0 {
				servicePort = 8080
			}
			accessURL = fmt.Sprintf("http://%s:%d", strings.TrimSpace(node.NetworkAliases[0]), servicePort)
			break
		}
		hostPort = 0
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: fmt.Sprintf("contest-topology-%d", req.ReservedHostPort),
		NetworkID:          fmt.Sprintf("contest-network-%d", req.ReservedHostPort),
		AccessURL:          accessURL,
		RuntimeDetails: model.InstanceRuntimeDetails{
			Containers: []model.InstanceRuntimeContainer{{
				NodeKey:      "entry",
				ContainerID:  fmt.Sprintf("contest-topology-%d", req.ReservedHostPort),
				HostPort:     hostPort,
				ServicePort:  8080,
				IsEntryPoint: true,
			}},
		},
	}, nil
}

func (contestInstanceTestRuntimeService) CreateContainer(_ context.Context, _ string, _ map[string]string, reservedHostPort int) (string, string, int, int, error) {
	return fmt.Sprintf("contest-container-%d", reservedHostPort), fmt.Sprintf("contest-network-%d", reservedHostPort), reservedHostPort, 8080, nil
}

func (contestInstanceTestRuntimeService) InspectManagedContainer(context.Context, string) (*practiceports.ManagedContainerState, error) {
	return &practiceports.ManagedContainerState{
		Exists:  true,
		Running: true,
		Status:  "running",
	}, nil
}

func newContestInstanceTestService(t *testing.T, db *gorm.DB) *practicecmd.Service {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:30000")
	if err != nil {
		t.Fatalf("listen readiness server: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	instanceRepo := runtimeinfrarepo.NewRepository(db)
	return practicecmd.NewService(
		practiceinfra.NewRepository(db),
		challengeRepo,
		imageRepo,
		instanceRepo,
		contestInstanceTestRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultExposedPort:   8080,
				PublicHost:           "127.0.0.1",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				MaxExtends:           2,
				CreateTimeout:        time.Second,
			},
		},
		zap.NewNop()).
		SetContestScopeRepository(practiceinfra.NewContestScopeRepository(practiceinfra.NewRepository(db))).
		SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo)).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

}

func seedContestInstanceChallenge(t *testing.T, db *gorm.DB, imageID, challengeID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.Image{
		ID:        imageID,
		Name:      fmt.Sprintf("ctf/web-%d", imageID),
		Tag:       fmt.Sprintf("v%d", imageID),
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         challengeID,
		Title:      "AWD Service",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    imageID,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
}

func seedContestInstanceUser(t *testing.T, db *gorm.DB, userID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.User{
		ID:           userID,
		Username:     fmt.Sprintf("user-%d", userID),
		PasswordHash: "hash",
		Role:         model.RoleStudent,
		Status:       model.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func seedContestInstanceAWDContest(t *testing.T, db *gorm.DB, contestID, challengeID int64, now time.Time) {
	t.Helper()
	seedContestInstanceAWDContestWithStatus(t, db, contestID, challengeID, model.ContestStatusRunning, now)
}

func seedContestInstanceAWDContestWithStatus(t *testing.T, db *gorm.DB, contestID, challengeID int64, status string, now time.Time) {
	t.Helper()
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: challengeID,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
}

func seedContestInstanceAWDService(t *testing.T, db *gorm.DB, serviceID, contestID, challengeID int64, now time.Time) {
	t.Helper()
	seedContestInstanceAWDServiceWithVisibility(t, db, serviceID, contestID, challengeID, true, now)
}

func seedContestInstanceAWDServiceWithVisibility(t *testing.T, db *gorm.DB, serviceID, contestID, challengeID int64, visible bool, now time.Time) {
	t.Helper()
	var challenge model.Challenge
	if err := db.Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		t.Fatalf("load challenge for awd service snapshot: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ID:             serviceID,
		ContestID:      contestID,
		AWDChallengeID: challengeID,
		DisplayName:    "Bank Portal",
		Order:          1,
		IsVisible:      visible,
		ScoreConfig:    `{"points":100}`,
		RuntimeConfig:  `{"checker_type":"http_standard"}`,
		ServiceSnapshot: mustEncodeContestInstanceAWDServiceSnapshot(t, model.ContestAWDServiceSnapshot{
			Name:       "Bank Portal",
			Category:   challenge.Category,
			Difficulty: challenge.Difficulty,
			RuntimeConfig: map[string]any{
				"image_id":         challenge.ImageID,
				"instance_sharing": string(model.InstanceSharingPerTeam),
				"defense_workspace": map[string]any{
					"entry_mode":      "ssh",
					"seed_root":       "runtime/workspace",
					"workspace_roots": []string{"runtime/workspace/app"},
					"writable_roots":  []string{"runtime/workspace/app"},
					"readonly_roots":  []string{},
					"runtime_mounts": []map[string]any{
						{"source": "runtime/workspace/app", "target": "/workspace/app", "mode": "rw"},
					},
				},
			},
			FlagConfig: map[string]any{
				"flag_type":   challenge.FlagType,
				"flag_prefix": "flag",
			},
		}),
		ValidationState: model.AWDCheckerValidationStatePending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
}

func mustEncodeContestInstanceAWDServiceSnapshot(t *testing.T, snapshot model.ContestAWDServiceSnapshot) string {
	t.Helper()

	raw, err := model.EncodeContestAWDServiceSnapshot(snapshot)
	if err != nil {
		t.Fatalf("encode contest awd service snapshot: %v", err)
	}
	return raw
}

func seedContestInstanceJeopardyContest(t *testing.T, db *gorm.DB, contestID, challengeID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "Jeopardy Contest",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestChallenge{
		ContestID:   contestID,
		ChallengeID: challengeID,
		Points:      100,
		IsVisible:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
}

func seedContestInstanceTeam(t *testing.T, db *gorm.DB, contestID, teamID, captainID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Alpha",
		CaptainID:  captainID,
		InviteCode: "alpha",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
}

func seedContestInstanceRegistration(t *testing.T, db *gorm.DB, contestID, userID, teamID int64, status string, now time.Time) {
	t.Helper()
	teamIDCopy := teamID
	if err := db.Create(&model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamIDCopy,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest registration: %v", err)
	}
}

func seedContestInstanceTeamMember(t *testing.T, db *gorm.DB, contestID, teamID, userID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    userID,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
}
