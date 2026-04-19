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
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeadapters "ctf-platform/internal/testutil/runtimeadapters"
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
	now := time.Now()

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

	var persistedServiceID int64
	if err := db.Raw("SELECT service_id FROM instances WHERE id = ?", resp.ID).
		Scan(&persistedServiceID).Error; err != nil {
		t.Fatalf("load persisted instance service_id: %v", err)
	}
	if persistedServiceID != 7003004 {
		t.Fatalf("expected instance.service_id=7003004, got %d", persistedServiceID)
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

	first, err := service.StartChallengeWithContext(context.Background(), 5101, 2101)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() first error = %v", err)
	}
	second, err := service.StartChallengeWithContext(context.Background(), 5102, 2101)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() second error = %v", err)
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
	resp, err := service.StartChallengeWithContext(context.Background(), 5202, 2201)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() error = %v", err)
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
	_, err := service.StartChallengeWithContext(context.Background(), 5001, 2201)
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
	runtimeCleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil)
	runtimeProvisioningService := runtimecmd.NewProvisioningService(instanceRepo, nil, &config.ContainerConfig{
		PortRangeStart:       30000,
		PortRangeEnd:         30010,
		DefaultExposedPort:   8080,
		PublicHost:           "127.0.0.1",
		DefaultTTL:           time.Hour,
		MaxConcurrentPerUser: 3,
		MaxExtends:           2,
		CreateTimeout:        time.Second,
	}, nil)
	return practicecmd.NewService(
		practiceinfra.NewRepository(db),
		challengeRepo,
		imageRepo,
		instanceRepo,
		runtimeadapters.NewPracticeRuntimeService(runtimeCleanupService, runtimeProvisioningService),
		nil,
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
		zap.NewNop(),
	)
}

func seedContestInstanceChallenge(t *testing.T, db *gorm.DB, imageID, challengeID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.Image{
		ID:        imageID,
		Name:      "ctf/web",
		Tag:       "v1",
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
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
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

func seedContestInstanceAWDService(t *testing.T, db *gorm.DB, serviceID, contestID, challengeID int64, now time.Time) {
	t.Helper()
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		ChallengeID:     challengeID,
		DisplayName:     "Bank Portal",
		Order:           1,
		IsVisible:       true,
		ScoreConfig:     `{"points":100}`,
		RuntimeConfig:   `{"checker_type":"http_standard"}`,
		ValidationState: model.AWDCheckerValidationStatePending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create contest awd service: %v", err)
	}
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
