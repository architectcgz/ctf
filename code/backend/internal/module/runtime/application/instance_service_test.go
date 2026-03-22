package application_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
)

type noopRuntimeCleaner struct{}

func (noopRuntimeCleaner) CleanupRuntimeWithContext(context.Context, *model.Instance) error {
	return nil
}

func TestInstanceServiceGetUserInstancesShowsContestSharedInstanceToTeamMember(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()
	contestID := int64(501)
	teamID := int64(601)

	seedInstanceServiceChallenge(t, db, &model.Challenge{
		ID:         102,
		Title:      "Shared AWD Challenge",
		Category:   model.DimensionPwn,
		Difficulty: model.ChallengeDifficultyMedium,
		FlagType:   model.FlagTypeDynamic,
		Status:     model.ChallengeStatusPublished,
		Points:     150,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeam(t, db, &model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime Team",
		CaptainID:  1,
		InviteCode: "runtime",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedInstanceServiceTeamMember(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedInstanceServiceInstance(t, db, &model.Instance{
		ID:          1002,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	service := runtimeapp.NewInstanceService(
		runtimeinfrarepo.NewRepository(db),
		noopRuntimeCleaner{},
		&config.ContainerConfig{MaxExtends: 2, ExtendDuration: 30 * time.Minute},
		nil,
	)

	items, err := service.GetUserInstancesWithContext(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstancesWithContext() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1002 {
		t.Fatalf("expected teammate visible shared instance, got %+v", items)
	}
}

func TestInstanceServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: model.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	service := runtimeapp.NewInstanceService(
		runtimeinfrarepo.NewRepository(db),
		noopRuntimeCleaner{},
		&config.ContainerConfig{MaxExtends: 2, ExtendDuration: 30 * time.Minute},
		nil,
	)

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(items), items)
	}
	if items[0].StudentUsername != "alice" || items[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", items[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func TestInstanceServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	db := newInstanceServiceTestDB(t)
	now := time.Now()

	seedInstanceServiceUser(t, db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 2, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceUser(t, db, &model.User{ID: 3, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceChallenge(t, db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 201, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstanceServiceInstance(t, db, &model.Instance{ID: 202, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	service := runtimeapp.NewInstanceService(
		runtimeinfrarepo.NewRepository(db),
		noopRuntimeCleaner{},
		&config.ContainerConfig{MaxExtends: 2, ExtendDuration: 30 * time.Minute},
		nil,
	)

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, model.RoleTeacher); err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	var instance model.Instance
	if err := db.First(&instance, 201).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

func newInstanceServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Instance{}, &model.PortAllocation{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Team{}, &model.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return db
}

func seedInstanceServiceUser(t *testing.T, db *gorm.DB, user *model.User) {
	t.Helper()
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func seedInstanceServiceChallenge(t *testing.T, db *gorm.DB, challenge *model.Challenge) {
	t.Helper()
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
}

func seedInstanceServiceTeam(t *testing.T, db *gorm.DB, team *model.Team) {
	t.Helper()
	if err := db.Create(team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
}

func seedInstanceServiceTeamMember(t *testing.T, db *gorm.DB, member *model.TeamMember) {
	t.Helper()
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
}

func seedInstanceServiceInstance(t *testing.T, db *gorm.DB, instance *model.Instance) {
	t.Helper()
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
}
