package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestFindAWDTargetProxyScopeReturnsCrossTeamRunningInstance(t *testing.T) {
	t.Parallel()

	db := newAWDTargetProxyRepositoryTestDB(t)
	now := time.Now()
	contestID := int64(9101)
	attackerTeamID := int64(9201)
	victimTeamID := int64(9202)
	serviceID := int64(9301)
	challengeID := int64(9401)
	instanceID := int64(9501)

	seedAWDTargetProxyRow(t, db, &model.Contest{
		ID:        contestID,
		Title:     "AWD",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: attackerTeamID, ContestID: contestID, Name: "Red", CaptainID: 1001, InviteCode: "red", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: victimTeamID, ContestID: contestID, Name: "Blue", CaptainID: 1002, InviteCode: "blue", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: attackerTeamID, UserID: 1001, JoinedAt: now, CreatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9601, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, ChallengeID: challengeID, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Instance{
		ID:          instanceID,
		UserID:      1002,
		ContestID:   &contestID,
		TeamID:      &victimTeamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ContainerID: "ctr-blue-web",
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39001",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	scope, err := NewRepository(db).FindAWDTargetProxyScope(context.Background(), 1001, contestID, serviceID, victimTeamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope == nil {
		t.Fatal("expected target scope")
	}
	if scope.InstanceID != instanceID || scope.AccessURL != "http://127.0.0.1:39001" {
		t.Fatalf("unexpected instance scope: %+v", scope)
	}
	if scope.AttackerTeamID != attackerTeamID || scope.VictimTeamID != victimTeamID {
		t.Fatalf("unexpected team scope: %+v", scope)
	}
	if scope.ServiceID != serviceID || scope.ChallengeID != challengeID {
		t.Fatalf("unexpected service scope: %+v", scope)
	}
}

func TestFindAWDTargetProxyScopeRejectsOwnTeamTarget(t *testing.T) {
	t.Parallel()

	db := newAWDTargetProxyRepositoryTestDB(t)
	now := time.Now()
	contestID := int64(9102)
	teamID := int64(9203)
	serviceID := int64(9302)

	seedAWDTargetProxyRow(t, db, &model.Contest{ID: contestID, Title: "AWD", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: teamID, ContestID: contestID, Name: "Red", CaptainID: 1003, InviteCode: "red-own", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 1003, JoinedAt: now, CreatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9602, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, ChallengeID: 9402, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Instance{
		ID:          9502,
		UserID:      1003,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 9402,
		ServiceID:   &serviceID,
		ContainerID: "ctr-red-web",
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39002",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	scope, err := NewRepository(db).FindAWDTargetProxyScope(context.Background(), 1003, contestID, serviceID, teamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope != nil {
		t.Fatalf("expected own team target to be rejected, got %+v", scope)
	}
}

func newAWDTargetProxyRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Contest{},
		&model.Team{},
		&model.TeamMember{},
		&model.AWDRound{},
		&model.ContestAWDService{},
		&model.Instance{},
	); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}

func seedAWDTargetProxyRow(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("seed row: %v", err)
	}
}
