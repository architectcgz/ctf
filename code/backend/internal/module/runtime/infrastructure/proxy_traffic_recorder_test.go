package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/model"
)

func TestProxyTrafficEventRecorderPrefersServiceChallengeMetadata(t *testing.T) {
	t.Parallel()

	db := newProxyTrafficRecorderTestDB(t)
	now := time.Now()
	contestID := int64(901)
	victimTeamID := int64(90111)
	attackerTeamID := int64(90112)
	serviceID := int64(90121)

	seedProxyTrafficRecorderRow(t, db, &model.Contest{
		ID:        contestID,
		Title:     "Runtime AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.AWDRound{
		ID:           90101,
		ContestID:    contestID,
		RoundNumber:  1,
		Status:       model.AWDRoundStatusRunning,
		StartedAt:    &now,
		AttackScore:  50,
		DefenseScore: 50,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.Team{
		ID:         victimTeamID,
		ContestID:  contestID,
		Name:       "Victim",
		CaptainID:  5001,
		InviteCode: "victim-team",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.Team{
		ID:         attackerTeamID,
		ContestID:  contestID,
		Name:       "Attacker",
		CaptainID:  5002,
		InviteCode: "attacker-team",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    victimTeamID,
		UserID:    5001,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.TeamMember{
		ContestID: contestID,
		TeamID:    attackerTeamID,
		UserID:    5002,
		JoinedAt:  now,
		CreatedAt: now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  9012,
		DisplayName:     "Bank Portal",
		Order:           1,
		IsVisible:       true,
		ValidationState: model.AWDCheckerValidationStatePassed,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.Instance{
		ID:          99001,
		UserID:      5001,
		ContestID:   &contestID,
		TeamID:      &victimTeamID,
		ChallengeID: 9011,
		ServiceID:   &serviceID,
		ContainerID: "ctr-runtime-proxy",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:39001",
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	recorder := NewProxyTrafficEventRecorder(db)
	if err := recorder.RecordRuntimeProxyTrafficEvent(context.Background(), 99001, 5002, "GET", "/bank", 200); err != nil {
		t.Fatalf("RecordRuntimeProxyTrafficEvent() error = %v", err)
	}

	var event model.AWDTrafficEvent
	if err := db.First(&event).Error; err != nil {
		t.Fatalf("load awd traffic event: %v", err)
	}
	if event.AWDChallengeID != 9012 {
		t.Fatalf("expected traffic event awd_challenge_id=9012 from contest awd service, got %+v", event)
	}
	if event.ServiceID != serviceID {
		t.Fatalf("expected traffic event service_id=%d from runtime instance, got %+v", serviceID, event)
	}
	if event.VictimTeamID != victimTeamID || event.AttackerTeamID != attackerTeamID {
		t.Fatalf("unexpected traffic event teams: %+v", event)
	}
}

func TestProxyTrafficEventRecorderRecordsExplicitAWDAttackScope(t *testing.T) {
	t.Parallel()

	db := newProxyTrafficRecorderTestDB(t)
	now := time.Now()
	contestID := int64(902)
	victimTeamID := int64(90211)
	attackerTeamID := int64(90212)
	serviceID := int64(90221)

	seedProxyTrafficRecorderRow(t, db, &model.Contest{
		ID:        contestID,
		Title:     "Runtime AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.AWDRound{
		ID:           90201,
		ContestID:    contestID,
		RoundNumber:  1,
		Status:       model.AWDRoundStatusRunning,
		StartedAt:    &now,
		AttackScore:  50,
		DefenseScore: 50,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.Team{
		ID:         victimTeamID,
		ContestID:  contestID,
		Name:       "Victim",
		CaptainID:  5001,
		InviteCode: "victim-team-2",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	seedProxyTrafficRecorderRow(t, db, &model.Team{
		ID:         attackerTeamID,
		ContestID:  contestID,
		Name:       "Attacker",
		CaptainID:  5002,
		InviteCode: "attacker-team-2",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	recorder := NewProxyTrafficEventRecorder(db)
	err := recorder.RecordAWDProxyTrafficEvent(context.Background(), model.AWDProxyTrafficEventInput{
		ContestID:      contestID,
		AttackerTeamID: attackerTeamID,
		VictimTeamID:   victimTeamID,
		ServiceID:      serviceID,
		AWDChallengeID: 90231,
		Method:         "POST",
		Path:           "/api/flag",
		StatusCode:     200,
	})
	if err != nil {
		t.Fatalf("RecordAWDProxyTrafficEvent() error = %v", err)
	}

	var event model.AWDTrafficEvent
	if err := db.First(&event).Error; err != nil {
		t.Fatalf("load awd traffic event: %v", err)
	}
	if event.RoundID != 90201 || event.ContestID != contestID {
		t.Fatalf("unexpected round scope: %+v", event)
	}
	if event.AttackerTeamID != attackerTeamID || event.VictimTeamID != victimTeamID {
		t.Fatalf("unexpected teams: %+v", event)
	}
	if event.ServiceID != serviceID || event.AWDChallengeID != 90231 {
		t.Fatalf("unexpected service scope: %+v", event)
	}
	if event.Method != "POST" || event.Path != "/api/flag" || event.StatusCode != 200 {
		t.Fatalf("unexpected request metadata: %+v", event)
	}
}

func newProxyTrafficRecorderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("%s/%s.sqlite", t.TempDir(), t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Contest{},
		&model.AWDRound{},
		&model.Team{},
		&model.TeamMember{},
		&model.ContestAWDService{},
		&model.Instance{},
		&model.AWDTrafficEvent{},
	); err != nil {
		t.Fatalf("migrate proxy traffic recorder tables: %v", err)
	}
	return db
}

func seedProxyTrafficRecorderRow(t *testing.T, db *gorm.DB, value any) {
	t.Helper()
	if err := db.Create(value).Error; err != nil {
		t.Fatalf("create test row: %v", err)
	}
}
