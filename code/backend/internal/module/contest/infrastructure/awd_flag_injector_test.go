package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/module/contest/testsupport"
)

type stubAWDContainerFileWriter struct {
	writes map[string]map[string]string
}

func (s *stubAWDContainerFileWriter) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	if s.writes == nil {
		s.writes = make(map[string]map[string]string)
	}
	if s.writes[containerID] == nil {
		s.writes[containerID] = make(map[string]string)
	}
	s.writes[containerID][filePath] = string(content)
	return nil
}

func TestDockerAWDFlagInjectorInjectsAllRunningTeamContainers(t *testing.T) {
	db := testsupport.SetupAWDTestDB(t)
	now := time.Now()

	testsupport.CreateAWDContestFixture(t, db, 10, now)
	testsupport.CreateAWDChallengeFixture(t, db, 1001, now)
	testsupport.CreateAWDTeamFixture(t, db, 1011, 10, "Alpha", now)
	testsupport.CreateAWDTeamMemberFixture(t, db, 10, 1011, 5001, now)
	testsupport.CreateAWDTeamMemberFixture(t, db, 10, 1011, 5002, now)
	testsupport.CreateAWDRoundFixture(t, db, 10001, 10, 1, 50, 50, now)

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{
			{ContainerID: "ctr-main"},
			{ContainerID: "ctr-sidecar"},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}
	if err := db.Create(&model.Instance{
		ID:             9001,
		UserID:         5001,
		ChallengeID:    1001,
		ContainerID:    "ctr-main",
		RuntimeDetails: runtimeDetails,
		Status:         model.InstanceStatusRunning,
		ExpiresAt:      now.Add(time.Hour),
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("create first instance: %v", err)
	}
	if err := db.Create(&model.Instance{
		ID:          9002,
		UserID:      5002,
		ChallengeID: 1001,
		ContainerID: "ctr-second-user",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create second instance: %v", err)
	}

	writer := &stubAWDContainerFileWriter{}
	injector := contestinfra.NewDockerAWDFlagInjector(db, writer, zap.NewNop())

	err = injector.InjectRoundFlags(context.Background(), &model.Contest{ID: 10}, &model.AWDRound{ID: 10001}, []contestports.AWDFlagAssignment{
		{TeamID: 1011, ChallengeID: 1001, Flag: "awd{round-flag}"},
	})
	if err != nil {
		t.Fatalf("InjectRoundFlags() error = %v", err)
	}

	if got := writer.writes["ctr-main"]["/flag/flag.txt"]; got != "awd{round-flag}" {
		t.Fatalf("unexpected main container write: %+v", writer.writes)
	}
	if got := writer.writes["ctr-sidecar"]["/flag/flag.txt"]; got != "awd{round-flag}" {
		t.Fatalf("unexpected sidecar write: %+v", writer.writes)
	}
	if got := writer.writes["ctr-second-user"]["/flag/flag.txt"]; got != "awd{round-flag}" {
		t.Fatalf("unexpected second user write: %+v", writer.writes)
	}
}

func TestDockerAWDFlagInjectorInjectsContestScopedTeamInstanceWithoutTeamMemberFallback(t *testing.T) {
	db := testsupport.SetupAWDTestDB(t)
	now := time.Now()

	testsupport.CreateAWDContestFixture(t, db, 20, now)
	testsupport.CreateAWDChallengeFixture(t, db, 2001, now)
	testsupport.CreateAWDRoundFixture(t, db, 20001, 20, 1, 50, 50, now)

	contestID := int64(20)
	teamID := int64(2011)
	if err := db.Create(&model.Instance{
		ID:          9901,
		UserID:      9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2001,
		ContainerID: "ctr-team-owned",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create team scoped instance: %v", err)
	}

	writer := &stubAWDContainerFileWriter{}
	injector := contestinfra.NewDockerAWDFlagInjector(db, writer, zap.NewNop())

	err := injector.InjectRoundFlags(context.Background(), &model.Contest{ID: 20}, &model.AWDRound{ID: 20001}, []contestports.AWDFlagAssignment{
		{TeamID: 2011, ChallengeID: 2001, Flag: "awd{contest-scoped}"},
	})
	if err != nil {
		t.Fatalf("InjectRoundFlags() error = %v", err)
	}

	if got := writer.writes["ctr-team-owned"]["/flag/flag.txt"]; got != "awd{contest-scoped}" {
		t.Fatalf("unexpected team scoped container write: %+v", writer.writes)
	}
}
