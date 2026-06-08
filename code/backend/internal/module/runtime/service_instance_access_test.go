package runtime_test

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	"ctf-platform/internal/shared/taxonomy"
)

func TestServiceDestroyInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(301)
	teamID := int64(401)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Alpha", CaptainID: 1, InviteCode: "alpha", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          901,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 101,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	if err := service.DestroyInstance(context.Background(), 901, 2); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 901)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != instanceentity.InstanceStatusStopping {
		t.Fatalf("expected stopping status, got %s", instance.Status)
	}
}

func TestServiceExtendInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(302)
	teamID := int64(402)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Beta", CaptainID: 1, InviteCode: "beta", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	initialExpiry := now.Add(time.Hour)
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          902,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		ContainerID: "contest-shared-extend",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   initialExpiry,
	})

	resp, err := service.ExtendInstance(context.Background(), 902, 2)
	if err != nil {
		t.Fatalf("ExtendInstance() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected extend response")
	}
	if resp.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1, got %d", resp.RemainingExtends)
	}

	instance, err := repo.FindByID(context.Background(), 902)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if !instance.ExpiresAt.After(initialExpiry) {
		t.Fatalf("expected expiry to be extended, got %s", instance.ExpiresAt)
	}
	if instance.ExtendCount != 1 {
		t.Fatalf("expected extend count 1, got %d", instance.ExtendCount)
	}
}

func TestServiceDestroyInstanceRejectsAWDTeamServiceInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(303)
	teamID := int64(403)
	serviceID := int64(503)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Gamma", CaptainID: 1, InviteCode: "gamma", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          905,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 105,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	err := service.DestroyInstance(context.Background(), 905, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for awd team service destroy, got %v", err)
	}
}

func TestServiceExtendInstanceRejectsAWDTeamServiceInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(304)
	teamID := int64(404)
	serviceID := int64(504)

	if err := repo.db.Create(&contestcontracts.Team{ID: teamID, ContestID: contestID, Name: "Delta", CaptainID: 1, InviteCode: "delta", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&contestcontracts.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          906,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 106,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	_, err := service.ExtendInstance(context.Background(), 906, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for awd team service extend, got %v", err)
	}
}

func TestServiceDestroyInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              903,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          903,
		UserID:      1,
		ChallengeID: 903,
		ShareScope:  instancecontracts.ShareScopeShared,
		ContainerID: "shared-ctr",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	err := service.DestroyInstance(context.Background(), 903, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared destroy, got %v", err)
	}
}

func TestServiceExtendInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&runtimeChallengeTestRow{
		ID:              904,
		Title:           "Shared Practice",
		Category:        taxonomy.DimensionWeb,
		Difficulty:      taxonomy.DifficultyEasy,
		FlagType:        challengecontracts.FlagTypeStatic,
		Status:          challengecontracts.ChallengeStatusPublished,
		InstanceSharing: challengecontracts.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &instanceentity.Instance{
		ID:          904,
		UserID:      1,
		ChallengeID: 904,
		ShareScope:  instancecontracts.ShareScopeShared,
		ContainerID: "shared-ctr",
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	_, err := service.ExtendInstance(context.Background(), 904, 2)
	if err == nil || err.Error() != apperror.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared extend, got %v", err)
	}
}
