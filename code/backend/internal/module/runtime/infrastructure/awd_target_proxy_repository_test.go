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
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{
		ID:             serviceID,
		ContestID:      contestID,
		AWDChallengeID: challengeID,
		DisplayName:    "Web",
		IsVisible:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	seedAWDTargetProxyRow(t, db, &model.Instance{
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
		ShareScope: model.InstanceSharingPerTeam,
		Status:     model.InstanceStatusRunning,
		AccessURL:  "http://awd-c9101-t9202-s9301:8080",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	scope, err := NewRepository(db).FindAWDTargetProxyScope(context.Background(), 1001, contestID, serviceID, victimTeamID)
	if err != nil {
		t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
	}
	if scope == nil {
		t.Fatal("expected target scope")
	}
	if scope.InstanceID != instanceID || scope.AccessURL != "http://172.30.10.20:8080" {
		t.Fatalf("unexpected instance scope: %+v", scope)
	}
	if scope.AttackerTeamID != attackerTeamID || scope.VictimTeamID != victimTeamID {
		t.Fatalf("unexpected team scope: %+v", scope)
	}
	if scope.ServiceID != serviceID || scope.AWDChallengeID != challengeID {
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
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 9402, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
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

func TestFindAWDDefenseSSHScopeReturnsOwnTeamRunningInstance(t *testing.T) {
	t.Parallel()

	db := newAWDTargetProxyRepositoryTestDB(t)
	now := time.Now()
	contestID := int64(9103)
	teamID := int64(9204)
	serviceID := int64(9303)
	challengeID := int64(9403)
	instanceID := int64(9503)

	seedAWDTargetProxyRow(t, db, &model.Contest{ID: contestID, Title: "AWD", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: teamID, ContestID: contestID, Name: "Red", CaptainID: 1004, InviteCode: "redssh", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 1004, JoinedAt: now, CreatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9603, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{
		ID:             serviceID,
		ContestID:      contestID,
		AWDChallengeID: challengeID,
		DisplayName:    "Web",
		IsVisible:      true,
		RuntimeConfig:  `{"challenge_runtime":{"defense_scope":{"editable_paths":["docker/challenge_app.py"],"protected_paths":["docker/app.py","docker/ctf_runtime.py","docker/check/check.py","challenge.yml"]}}}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	seedAWDTargetProxyRow(t, db, &model.Instance{
		ID:          instanceID,
		UserID:      1004,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ContainerID: "ctr-red-web",
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39003",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	seedAWDTargetProxyRow(t, db, &model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instanceID,
		WorkspaceRevision: 7,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-red-web",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	})

	scope, err := NewRepository(db).FindAWDDefenseSSHScope(context.Background(), 1004, contestID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseSSHScope() error = %v", err)
	}
	if scope == nil {
		t.Fatal("expected defense ssh scope")
	}
	if scope.InstanceID != instanceID || scope.ContainerID != "workspace-red-web" {
		t.Fatalf("unexpected instance scope: %+v", scope)
	}
	if scope.TeamID != teamID || scope.ServiceID != serviceID || scope.AWDChallengeID != challengeID || scope.WorkspaceRevision != 7 {
		t.Fatalf("unexpected team/service scope: %+v", scope)
	}
}

func TestFindAWDDefenseSSHScopeRejectsOtherTeamInstance(t *testing.T) {
	t.Parallel()

	db := newAWDTargetProxyRepositoryTestDB(t)
	now := time.Now()
	contestID := int64(9104)
	ownTeamID := int64(9205)
	otherTeamID := int64(9206)
	serviceID := int64(9304)
	challengeID := int64(9404)

	seedAWDTargetProxyRow(t, db, &model.Contest{ID: contestID, Title: "AWD", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: ownTeamID, ContestID: contestID, Name: "Red", CaptainID: 1005, InviteCode: "ownssh", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Team{ID: otherTeamID, ContestID: contestID, Name: "Blue", CaptainID: 1006, InviteCode: "othssh", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: ownTeamID, UserID: 1005, JoinedAt: now, CreatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9604, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: challengeID, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
	seedAWDTargetProxyRow(t, db, &model.Instance{
		ID:          9504,
		UserID:      1006,
		ContestID:   &contestID,
		TeamID:      &otherTeamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ContainerID: "ctr-blue-web",
		ShareScope:  model.InstanceSharingPerTeam,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39004",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	scope, err := NewRepository(db).FindAWDDefenseSSHScope(context.Background(), 1005, contestID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseSSHScope() error = %v", err)
	}
	if scope != nil {
		t.Fatalf("expected other team instance to be rejected, got %+v", scope)
	}
}

func TestFindAWDTargetProxyScopeReturnsNilWhenScopeControlled(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		teamID      int64
		scopeType   string
		controlType string
		serviceID   int64
	}{
		{
			name:        "attacker_team_retired",
			teamID:      9210,
			scopeType:   model.AWDScopeControlScopeTeam,
			controlType: model.AWDScopeControlTypeRetired,
			serviceID:   0,
		},
		{
			name:        "attacker_service_disabled",
			teamID:      9210,
			scopeType:   model.AWDScopeControlScopeTeamService,
			controlType: model.AWDScopeControlTypeServiceDisabled,
			serviceID:   9310,
		},
		{
			name:        "victim_team_retired",
			teamID:      9211,
			scopeType:   model.AWDScopeControlScopeTeam,
			controlType: model.AWDScopeControlTypeRetired,
			serviceID:   0,
		},
		{
			name:        "victim_service_disabled",
			teamID:      9211,
			scopeType:   model.AWDScopeControlScopeTeamService,
			controlType: model.AWDScopeControlTypeServiceDisabled,
			serviceID:   9310,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newAWDTargetProxyRepositoryTestDB(t)
			now := time.Now().UTC()
			contestID := int64(9110)
			attackerTeamID := int64(9210)
			victimTeamID := int64(9211)
			serviceID := int64(9310)

			seedAWDTargetProxyRow(t, db, &model.Contest{ID: contestID, Title: "AWD", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.Team{ID: attackerTeamID, ContestID: contestID, Name: "Red", CaptainID: 1010, InviteCode: "red", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.Team{ID: victimTeamID, ContestID: contestID, Name: "Blue", CaptainID: 1011, InviteCode: "blue", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: attackerTeamID, UserID: 1010, JoinedAt: now, CreatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9610, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: 9410, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.Instance{
				ID:          9510,
				UserID:      1011,
				ContestID:   &contestID,
				TeamID:      &victimTeamID,
				ChallengeID: 9410,
				ServiceID:   &serviceID,
				ContainerID: "ctr-blue-web",
				ShareScope:  model.InstanceSharingPerTeam,
				Status:      model.InstanceStatusRunning,
				AccessURL:   "http://127.0.0.1:39110",
				ExpiresAt:   now.Add(time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			seedAWDTargetProxyRow(t, db, &model.AWDScopeControl{
				ContestID:   contestID,
				TeamID:      tc.teamID,
				ScopeType:   tc.scopeType,
				ServiceID:   tc.serviceID,
				ControlType: tc.controlType,
				Reason:      tc.name,
				CreatedAt:   now,
				UpdatedAt:   now,
			})

			scope, err := NewRepository(db).FindAWDTargetProxyScope(context.Background(), 1010, contestID, serviceID, victimTeamID)
			if err != nil {
				t.Fatalf("FindAWDTargetProxyScope() error = %v", err)
			}
			if scope != nil {
				t.Fatalf("expected controlled target scope to be hidden, got %+v", scope)
			}
		})
	}
}

func TestFindAWDDefenseSSHScopeReturnsNilWhenScopeControlled(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		scopeType   string
		controlType string
		serviceID   int64
	}{
		{
			name:        "team_retired",
			scopeType:   model.AWDScopeControlScopeTeam,
			controlType: model.AWDScopeControlTypeRetired,
			serviceID:   0,
		},
		{
			name:        "service_disabled",
			scopeType:   model.AWDScopeControlScopeTeamService,
			controlType: model.AWDScopeControlTypeServiceDisabled,
			serviceID:   9311,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newAWDTargetProxyRepositoryTestDB(t)
			now := time.Now().UTC()
			contestID := int64(9111)
			teamID := int64(9212)
			serviceID := int64(9311)
			challengeID := int64(9411)
			instanceID := int64(9511)

			seedAWDTargetProxyRow(t, db, &model.Contest{ID: contestID, Title: "AWD", Mode: model.ContestModeAWD, Status: model.ContestStatusRunning, StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.Team{ID: teamID, ContestID: contestID, Name: "Red", CaptainID: 1012, InviteCode: "redssh", MaxMembers: 4, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 1012, JoinedAt: now, CreatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.AWDRound{ID: 9611, ContestID: contestID, RoundNumber: 1, Status: model.AWDRoundStatusRunning, StartedAt: &now, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.ContestAWDService{ID: serviceID, ContestID: contestID, AWDChallengeID: challengeID, DisplayName: "Web", IsVisible: true, CreatedAt: now, UpdatedAt: now})
			seedAWDTargetProxyRow(t, db, &model.Instance{
				ID:          instanceID,
				UserID:      1012,
				ContestID:   &contestID,
				TeamID:      &teamID,
				ChallengeID: challengeID,
				ServiceID:   &serviceID,
				ContainerID: "ctr-red-web",
				ShareScope:  model.InstanceSharingPerTeam,
				Status:      model.InstanceStatusRunning,
				AccessURL:   "http://127.0.0.1:39111",
				ExpiresAt:   now.Add(time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			seedAWDTargetProxyRow(t, db, &model.AWDDefenseWorkspace{
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				InstanceID:        instanceID,
				WorkspaceRevision: 2,
				Status:            model.AWDDefenseWorkspaceStatusRunning,
				ContainerID:       "workspace-red-web",
				SeedSignature:     "seed:v1",
				CreatedAt:         now,
				UpdatedAt:         now,
			})
			seedAWDTargetProxyRow(t, db, &model.AWDScopeControl{
				ContestID:   contestID,
				TeamID:      teamID,
				ScopeType:   tc.scopeType,
				ServiceID:   tc.serviceID,
				ControlType: tc.controlType,
				Reason:      tc.name,
				CreatedAt:   now,
				UpdatedAt:   now,
			})

			scope, err := NewRepository(db).FindAWDDefenseSSHScope(context.Background(), 1012, contestID, serviceID)
			if err != nil {
				t.Fatalf("FindAWDDefenseSSHScope() error = %v", err)
			}
			if scope != nil {
				t.Fatalf("expected controlled defense scope to be hidden, got %+v", scope)
			}
		})
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
		&model.AWDDefenseWorkspace{},
		&model.AWDScopeControl{},
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
