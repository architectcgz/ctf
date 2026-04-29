package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
)

func TestAWDRepositoryRecordRuntimeProxyTrafficEventPrefersServiceChallengeMetadata(t *testing.T) {
	db := testsupport.SetupAWDTestDB(t)
	now := time.Now()

	testsupport.CreateAWDContestFixture(t, db, 901, now)
	testsupport.CreateAWDRoundFixture(t, db, 90101, 901, 1, 50, 50, now)
	testsupport.CreateAWDChallengeFixture(t, db, 9011, now)
	testsupport.CreateAWDChallengeFixture(t, db, 9012, now)
	testsupport.CreateAWDContestChallengeFixture(t, db, 901, 9012, now)
	testsupport.CreateAWDTeamFixture(t, db, 90111, 901, "Victim", now)
	testsupport.CreateAWDTeamFixture(t, db, 90112, 901, "Attacker", now)
	testsupport.CreateAWDTeamMemberFixture(t, db, 901, 90111, 5001, now)
	testsupport.CreateAWDTeamMemberFixture(t, db, 901, 90112, 5002, now)

	serviceID := testsupport.DefaultAWDContestServiceID(901, 9012)
	contestID := int64(901)
	victimTeamID := int64(90111)
	if err := db.Create(&model.Instance{
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
	}).Error; err != nil {
		t.Fatalf("create awd runtime instance: %v", err)
	}

	repo := contestinfra.NewAWDRepository(db)
	if err := repo.RecordRuntimeProxyTrafficEvent(context.Background(), 99001, 5002, "GET", "/bank", 200); err != nil {
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
	if event.VictimTeamID != 90111 || event.AttackerTeamID != 90112 {
		t.Fatalf("unexpected traffic event teams: %+v", event)
	}
}
