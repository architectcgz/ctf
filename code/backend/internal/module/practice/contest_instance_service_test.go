package practice_test

import (
	"context"
	"fmt"
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
	practiceModule "ctf-platform/internal/module/practice"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeadapters "ctf-platform/internal/testutil/runtimeadapters"
	"ctf-platform/pkg/errcode"
)

func TestServiceStartContestChallengeAWDCreatesAndReusesTeamInstance(t *testing.T) {
	db := newContestInstanceTestDB(t)
	now := time.Now()

	seedContestInstanceChallenge(t, db, 1001, 2001, now)
	seedContestInstanceAWDContest(t, db, 3001, 2001, now)
	seedContestInstanceTeam(t, db, 3001, 4001, 5001, now)
	seedContestInstanceRegistration(t, db, 3001, 5001, 4001, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceRegistration(t, db, 3001, 5002, 4001, model.ContestRegistrationStatusApproved, now)
	seedContestInstanceTeamMember(t, db, 3001, 4001, 5001, now)
	seedContestInstanceTeamMember(t, db, 3001, 4001, 5002, now)

	service := newContestInstanceTestService(db)

	first, err := service.StartContestChallenge(context.Background(), 5001, 3001, 2001)
	if err != nil {
		t.Fatalf("StartContestChallenge() first error = %v", err)
	}
	if first.ID == 0 {
		t.Fatalf("expected created instance id, got %+v", first)
	}

	second, err := service.StartContestChallenge(context.Background(), 5002, 3001, 2001)
	if err != nil {
		t.Fatalf("StartContestChallenge() second error = %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected shared instance reuse, got first=%d second=%d", first.ID, second.ID)
	}

	var instance model.Instance
	if err := db.First(&instance, first.ID).Error; err != nil {
		t.Fatalf("load created instance: %v", err)
	}
	if instance.ContestID == nil || *instance.ContestID != 3001 {
		t.Fatalf("expected contest scoped instance, got %+v", instance)
	}
	if instance.TeamID == nil || *instance.TeamID != 4001 {
		t.Fatalf("expected team scoped instance, got %+v", instance)
	}

	runtimeCleanupService := runtimeapp.NewRuntimeCleanupService(nil, nil)
	instanceService := runtimeapp.NewInstanceService(runtimeinfrarepo.NewRepository(db), runtimeCleanupService, &config.ContainerConfig{
		MaxExtends:     2,
		ExtendDuration: 30 * time.Minute,
	}, nil)
	visible, err := instanceService.GetUserInstancesWithContext(context.Background(), 5002)
	if err != nil {
		t.Fatalf("GetUserInstancesWithContext() error = %v", err)
	}
	if len(visible) != 1 || visible[0].ID != first.ID {
		t.Fatalf("expected teammate to see shared instance, got %+v", visible)
	}

	accessURL, err := instanceService.GetAccessURLWithContext(context.Background(), first.ID, 5002)
	if err != nil {
		t.Fatalf("GetAccessURLWithContext() error = %v", err)
	}
	if accessURL != first.AccessURL {
		t.Fatalf("expected teammate to access shared instance, got %q", accessURL)
	}
}

func TestServiceStartContestChallengeAWDReturnsExistingTeamInstance(t *testing.T) {
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
		ContainerID: "existing-team-instance",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30001",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("seed shared instance: %v", err)
	}

	service := newContestInstanceTestService(db)
	resp, err := service.StartContestChallenge(context.Background(), 5004, 3002, 2002)
	if err != nil {
		t.Fatalf("StartContestChallenge() error = %v", err)
	}
	if resp.ID != 9002 {
		t.Fatalf("expected existing shared instance, got %+v", resp)
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

	service := newContestInstanceTestService(db)
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
		&model.Image{},
		&model.Challenge{},
		&model.ChallengeTopology{},
		&model.Contest{},
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

func newContestInstanceTestService(db *gorm.DB) *practiceModule.Service {
	challengeRepo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	instanceRepo := runtimeinfrarepo.NewRepository(db)
	runtimeCleanupService := runtimeapp.NewRuntimeCleanupService(nil, nil)
	runtimeProvisioningService := runtimeapp.NewProvisioningService(instanceRepo, nil, &config.ContainerConfig{
		PortRangeStart:       30000,
		PortRangeEnd:         30010,
		DefaultExposedPort:   8080,
		PublicHost:           "127.0.0.1",
		DefaultTTL:           time.Hour,
		MaxConcurrentPerUser: 3,
		MaxExtends:           2,
		CreateTimeout:        time.Second,
	}, nil)
	return practiceModule.NewService(
		practiceModule.NewRepository(db),
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
