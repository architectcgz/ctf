package infrastructure

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/testsupport"
	"testing"
	"time"
)

func TestRepositoryCreate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	err := repo.Create(challenge)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if challenge.ID == 0 {
		t.Fatal("ID should be set")
	}
}

func TestRepositoryFindByID(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test"}
	db.Create(challenge)

	found, err := repo.FindByID(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Title != "Test" {
		t.Fatalf("unexpected title: %s", found.Title)
	}
}

func TestRepositoryList(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	db.Create(&model.Challenge{Title: "C1", Category: "web"})
	db.Create(&model.Challenge{Title: "C2", Category: "pwn"})

	challenges, total, err := repo.ListWithContext(context.Background(), &dto.ChallengeQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListWithContext() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(challenges) != 2 {
		t.Fatalf("unexpected count: %d", len(challenges))
	}
}

func TestRepositoryHasRunningInstances(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test"}
	db.Create(challenge)
	db.Create(&model.Instance{ChallengeID: challenge.ID, Status: "running"})

	has, err := repo.HasRunningInstances(context.Background(), challenge.ID)
	if err != nil {
		t.Fatalf("HasRunningInstances() error = %v", err)
	}
	if !has {
		t.Fatal("should have running instances")
	}
}

func TestRepositoryPublishCheckJobLifecycle(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)
	user := &model.User{Username: "teacher", PasswordHash: "x", Role: model.RoleTeacher}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	challenge := &model.Challenge{Title: "Test", Status: model.ChallengeStatusDraft, CreatedBy: &user.ID}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	ctx := context.Background()
	job := &model.ChallengePublishCheckJob{
		ChallengeID:   challenge.ID,
		RequestedBy:   user.ID,
		Status:        model.ChallengePublishCheckStatusPending,
		RequestSource: "admin_publish",
	}
	if err := repo.CreatePublishCheckJob(ctx, job); err != nil {
		t.Fatalf("CreatePublishCheckJob() error = %v", err)
	}

	active, err := repo.FindActivePublishCheckJobByChallengeID(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("FindActivePublishCheckJobByChallengeID() error = %v", err)
	}
	if active.ID != job.ID {
		t.Fatalf("unexpected active job id: %d", active.ID)
	}

	byID, err := repo.FindPublishCheckJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("FindPublishCheckJobByID() error = %v", err)
	}
	if byID.ID != job.ID {
		t.Fatalf("unexpected job by id: %+v", byID)
	}

	jobs, err := repo.ListPendingPublishCheckJobs(ctx, 10)
	if err != nil {
		t.Fatalf("ListPendingPublishCheckJobs() error = %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("unexpected pending jobs: %+v", jobs)
	}

	startedAt := time.Now()
	started, err := repo.TryStartPublishCheckJob(ctx, job.ID, startedAt)
	if err != nil {
		t.Fatalf("TryStartPublishCheckJob() error = %v", err)
	}
	if !started {
		t.Fatal("expected job to start")
	}

	latest, err := repo.FindLatestPublishCheckJobByChallengeID(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("FindLatestPublishCheckJobByChallengeID() error = %v", err)
	}
	if latest.Status != model.ChallengePublishCheckStatusRunning {
		t.Fatalf("unexpected latest status: %s", latest.Status)
	}

	latest.Status = model.ChallengePublishCheckStatusFailed
	latest.FailureSummary = "runtime failed"
	if err := repo.UpdatePublishCheckJob(ctx, latest); err != nil {
		t.Fatalf("UpdatePublishCheckJob() error = %v", err)
	}

	updated, err := repo.FindLatestPublishCheckJobByChallengeID(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("FindLatestPublishCheckJobByChallengeID() after update error = %v", err)
	}
	if updated.FailureSummary != "runtime failed" {
		t.Fatalf("unexpected failure summary: %s", updated.FailureSummary)
	}
}

func TestRepositorySchemaOmitsHintUnlockArtifacts(t *testing.T) {
	db := testsupport.SetupTestDB(t)

	if db.Migrator().HasTable("challenge_hint_unlocks") {
		t.Fatal("expected challenge_hint_unlocks table to be removed")
	}
	if db.Migrator().HasColumn(&model.ChallengeHint{}, "cost_points") {
		t.Fatal("expected challenge_hints.cost_points column to be removed")
	}
}

func TestRepositoryCreateAndListAWDServiceTemplates(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	template := &model.AWDServiceTemplate{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     "hard",
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDServiceTemplateStatusDraft,
	}

	if err := repo.CreateAWDServiceTemplate(context.Background(), template); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if template.ID == 0 {
		t.Fatal("template ID should be set")
	}

	items, total, err := repo.ListAWDServiceTemplates(context.Background(), &dto.AWDServiceTemplateQuery{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected item count: %d", len(items))
	}
	if items[0].Slug != "bank-portal-awd" {
		t.Fatalf("unexpected template slug: %s", items[0].Slug)
	}
}
