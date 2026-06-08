package runtime_test

import (
	"context"
	"testing"
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	"ctf-platform/internal/shared/taxonomy"
)

func TestServiceGetUserInstancesIncludesChallengeMetadata(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:         101,
		Title:      "Matrix Web Challenge",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		FlagType:   challengecontracts.FlagTypeStatic,
		Status:     challengecontracts.ChallengeStatusPublished,
		Points:     100,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1001,
		UserID:      1,
		ChallengeID: 101,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30001",
		ExpiresAt:   now.Add(time.Hour),
		ExtendCount: 1,
		MaxExtends:  3,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %+v", items)
	}
	item := items[0]
	if item.ChallengeTitle != "Matrix Web Challenge" {
		t.Fatalf("expected challenge title, got %+v", item)
	}
	if item.Category != taxonomy.DimensionWeb {
		t.Fatalf("expected category %q, got %+v", taxonomy.DimensionWeb, item)
	}
	if item.Difficulty != taxonomy.DifficultyEasy {
		t.Fatalf("expected difficulty %q, got %+v", taxonomy.DifficultyEasy, item)
	}
	if item.FlagType != challengecontracts.FlagTypeStatic {
		t.Fatalf("expected flag type %q, got %+v", challengecontracts.FlagTypeStatic, item)
	}
	if item.RemainingExtends != 2 {
		t.Fatalf("expected remaining extends 2, got %+v", item)
	}
}

func TestServiceGetUserInstancesShowsContestSharedInstanceToTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(501)
	teamID := int64(601)

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:         102,
		Title:      "Shared AWD Challenge",
		Category:   taxonomy.DimensionPwn,
		Difficulty: taxonomy.DifficultyMedium,
		FlagType:   challengecontracts.FlagTypeDynamic,
		Status:     challengecontracts.ChallengeStatusPublished,
		Points:     150,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime Team",
		CaptainID:  1,
		InviteCode: "runtime",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1002,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1002 {
		t.Fatalf("expected teammate visible shared instance, got %+v", items)
	}
}

func TestServiceGetUserInstancesShowsPracticeSharedInstanceToAnyUser(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              103,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		Points:          80,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          1003,
		UserID:      1,
		ChallengeID: 103,
		ShareScope:  instancecontracts.ShareScopeShared,
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30003",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1003 {
		t.Fatalf("expected global shared instance visible to another user, got %+v", items)
	}
	if items[0].ShareScope != instancecontracts.ShareScopeShared {
		t.Fatalf("expected share scope to be returned, got %+v", items[0])
	}
}
