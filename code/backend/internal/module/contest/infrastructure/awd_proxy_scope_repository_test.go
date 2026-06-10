package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func TestAWDProxyScopeRepositoryReturnsCrossTeamRunningInstance(t *testing.T) {
	t.Parallel()

	db := newAWDProxyScopeRepositoryTestDB(t)
	now := time.Now().UTC()
	contestID := int64(9101)
	attackerTeamID := int64(9201)
	victimTeamID := int64(9202)
	serviceID := int64(9301)
	challengeID := int64(9401)
	instanceID := int64(9501)

	seedAWDProxyScopeRow(t, db, &contestentity.Contest{
		ID:        contestID,
		Title:     "AWD",
		Mode:      contestentity.ContestModeAWD,
		Status:    contestentity.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: attackerTeamID, ContestID: contestID, Name: "Red", CaptainID: 1001, InviteCode: "red", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: victimTeamID, ContestID: contestID, Name: "Blue", CaptainID: 1002, InviteCode: "blue", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.TeamMember{ContestID: contestID, TeamID: attackerTeamID, UserID: 1001, JoinedAt: now, CreatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDRound{ID: 9601, ContestID: contestID, RoundNumber: 1, Status: contestentity.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.ContestAWDService{
		ID:             serviceID,
		ContestID:      contestID,
		AWDChallengeID: challengeID,
		DisplayName:    "Web",
		IsVisible:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	seedAWDProxyScopeRow(t, db, &instancecontracts.Instance{
		ID:          instanceID,
		UserID:      1002,
		ContestID:   &contestID,
		TeamID:      &victimTeamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ContainerID: "ctr-blue-web",
		RuntimeDetails: `{
			"networks":[{"key":"default","name":"ctf-awd-contest-9101","network_id":"net-awd-contest-9101","shared":true}],
			"containers":[{"container_id":"ctr-blue-web","is_entry_point":true,"network_keys":["default"],"network_aliases":["awd-c9101-t9202-s9301"],"network_ips":{"ctf-awd-contest-9101":"172.30.10.20"}}]
		}`,
		ShareScope: instancecontracts.ShareScopePerTeam,
		Status:     instancecontracts.InstanceStatusRunning,
		AccessURL:  "http://awd-c9101-t9202-s9301:8080",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	scope, err := NewAWDRepository(db).FindAWDTargetProxyScope(context.Background(), 1001, contestID, serviceID, victimTeamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope == nil {
		t.Fatal("expected target scope")
	}
	if scope.InstanceID != instanceID || scope.AccessURL != "http://172.30.10.20:8080" {
		t.Fatalf("unexpected instance scope: %+v", scope)
	}
	if scope.AttackerTeamID != attackerTeamID || scope.VictimTeamID != victimTeamID || scope.ServiceID != serviceID || scope.AWDChallengeID != challengeID {
		t.Fatalf("unexpected AWD scope: %+v", scope)
	}
}

func TestAWDProxyScopeRepositoryRejectsOwnTeamTarget(t *testing.T) {
	t.Parallel()

	db := newAWDProxyScopeRepositoryTestDB(t)
	now := time.Now().UTC()
	contestID := int64(9102)
	teamID := int64(9203)
	serviceID := int64(9302)

	seedAWDProxyScopeRow(t, db, &contestentity.Contest{ID: contestID, Title: "AWD", Mode: contestentity.ContestModeAWD, Status: contestentity.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: teamID, ContestID: contestID, Name: "Red", CaptainID: 1003, InviteCode: "red-own", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 1003, JoinedAt: now, CreatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDRound{ID: 9602, ContestID: contestID, RoundNumber: 1, Status: contestentity.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 9402, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &instancecontracts.Instance{
		ID:          9502,
		UserID:      1003,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 9402,
		ServiceID:   &serviceID,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		Status:      instancecontracts.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39002",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	scope, err := NewAWDRepository(db).FindAWDTargetProxyScope(context.Background(), 1003, contestID, serviceID, teamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope != nil {
		t.Fatalf("expected own team target to be rejected, got %+v", scope)
	}
}

func TestAWDProxyScopeRepositoryHidesControlledTarget(t *testing.T) {
	t.Parallel()

	db := newAWDProxyScopeRepositoryTestDB(t)
	now := time.Now().UTC()
	contestID := int64(9110)
	attackerTeamID := int64(9210)
	victimTeamID := int64(9211)
	serviceID := int64(9310)

	seedAWDProxyScopeRow(t, db, &contestentity.Contest{ID: contestID, Title: "AWD", Mode: contestentity.ContestModeAWD, Status: contestentity.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: attackerTeamID, ContestID: contestID, Name: "Red", CaptainID: 1010, InviteCode: "red", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: victimTeamID, ContestID: contestID, Name: "Blue", CaptainID: 1011, InviteCode: "blue", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.TeamMember{ContestID: contestID, TeamID: attackerTeamID, UserID: 1010, JoinedAt: now, CreatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDRound{ID: 9610, ContestID: contestID, RoundNumber: 1, Status: contestentity.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 9410, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &instancecontracts.Instance{
		ID:          9510,
		UserID:      1011,
		ContestID:   &contestID,
		TeamID:      &victimTeamID,
		ChallengeID: 9410,
		ServiceID:   &serviceID,
		ShareScope:  instancecontracts.ShareScopePerTeam,
		Status:      instancecontracts.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39110",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDScopeControl{
		ContestID:   contestID,
		TeamID:      victimTeamID,
		ScopeType:   contestentity.AWDScopeControlScopeTeamService,
		ServiceID:   serviceID,
		ControlType: contestentity.AWDScopeControlTypeServiceDisabled,
		Reason:      "service_disabled",
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	scope, err := NewAWDRepository(db).FindAWDTargetProxyScope(context.Background(), 1010, contestID, serviceID, victimTeamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope != nil {
		t.Fatalf("expected controlled target scope to be hidden, got %+v", scope)
	}
}

func TestAWDProxyScopeRepositoryReturnsDefenseSSHWorkspace(t *testing.T) {
	t.Parallel()

	db := newAWDProxyScopeRepositoryTestDB(t)
	now := time.Now().UTC()
	contestID := int64(9103)
	teamID := int64(9204)
	serviceID := int64(9303)
	challengeID := int64(9403)
	instanceID := int64(9503)

	seedAWDProxyScopeRow(t, db, &contestentity.Contest{ID: contestID, Title: "AWD", Mode: contestentity.ContestModeAWD, Status: contestentity.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.Team{ID: teamID, ContestID: contestID, Name: "Red", CaptainID: 1004, InviteCode: "redssh", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 1004, JoinedAt: now, CreatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDRound{ID: 9603, ContestID: contestID, RoundNumber: 1, Status: contestentity.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &contestentity.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: challengeID, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDProxyScopeRow(t, db, &instancecontracts.Instance{
		ID:          instanceID,
		UserID:      1004,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ContainerID: "ctr-red-web",
		ShareScope:  instancecontracts.ShareScopePerTeam,
		Status:      instancecontracts.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39003",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	seedAWDProxyScopeRow(t, db, &contestentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instanceID,
		WorkspaceRevision: 7,
		Status:            contestentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-red-web",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	})

	scope, err := NewAWDRepository(db).FindAWDDefenseSSHScope(context.Background(), 1004, contestID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseSSHScope() error = %v", err)
	}
	if scope == nil {
		t.Fatal("expected defense ssh scope")
	}
	if scope.InstanceID != instanceID || scope.ContainerID != "workspace-red-web" || scope.WorkspaceRevision != 7 {
		t.Fatalf("unexpected defense scope: %+v", scope)
	}
	if scope.TeamID != teamID || scope.ServiceID != serviceID || scope.AWDChallengeID != challengeID {
		t.Fatalf("unexpected team/service scope: %+v", scope)
	}
}

func newAWDProxyScopeRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&contestentity.Contest{},
		&contestentity.Team{},
		&contestentity.TeamMember{},
		&contestentity.AWDRound{},
		&contestentity.ContestAWDService{},
		&contestentity.AWDDefenseWorkspace{},
		&contestentity.AWDScopeControl{},
		&instancecontracts.Instance{},
	); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}

func seedAWDProxyScopeRow(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("seed row: %v", err)
	}
}
