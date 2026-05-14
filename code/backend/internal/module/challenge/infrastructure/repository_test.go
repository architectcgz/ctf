package infrastructure

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/testsupport"
	"reflect"
	"testing"
	"time"
)

func TestRepositoryCreate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	challenge := &model.Challenge{Title: "Test", Status: "draft"}
	err := repo.Create(context.Background(), challenge)
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

	challenges, total, err := repo.List(context.Background(), &dto.ChallengeQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(challenges) != 2 {
		t.Fatalf("unexpected count: %d", len(challenges))
	}
}

func TestRepositoryListPublishedUsesOnlyJeopardyChallenges(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	normalChallenge := &model.Challenge{
		Title:  "Normal Web",
		Status: model.ChallengeStatusPublished,
	}
	awdChallenge := &model.AWDChallenge{
		Name:   "AWD Web",
		Status: model.AWDChallengeStatusPublished,
	}
	if err := db.Create(normalChallenge).Error; err != nil {
		t.Fatalf("create normal challenge: %v", err)
	}
	if err := db.Create(awdChallenge).Error; err != nil {
		t.Fatalf("create awd challenge: %v", err)
	}
	now := time.Now()
	contest := &model.Contest{
		Title:     "AWD Contest",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
	}
	if err := db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestAWDService{
		ContestID:      contest.ID,
		AWDChallengeID: awdChallenge.ID,
		DisplayName:    "AWD Web",
		IsVisible:      true,
	}).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	challenges, total, err := repo.ListPublished(context.Background(), &dto.ChallengeQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListPublished() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(challenges) != 1 {
		t.Fatalf("unexpected count: %d", len(challenges))
	}
	if challenges[0].ID != normalChallenge.ID {
		t.Fatalf("expected normal challenge, got %+v", challenges[0])
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

func TestRepositoryFindPublishedForRecommendation(t *testing.T) {
	db := testsupport.SetupTagTestDB(t)
	repo := NewRepository(db)

	solved := &model.Challenge{
		Title:      "Solved Pwn",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyBeginner,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
	}
	fresh := &model.Challenge{
		Title:      "Fresh Pwn",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     120,
		Status:     model.ChallengeStatusPublished,
	}
	tagged := &model.Challenge{
		Title:      "Tagged Web",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     150,
		Status:     model.ChallengeStatusPublished,
	}
	ignoredDraft := &model.Challenge{
		Title:      "Draft Pwn",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyBeginner,
		Points:     90,
		Status:     model.ChallengeStatusDraft,
	}
	for _, challenge := range []*model.Challenge{solved, fresh, tagged, ignoredDraft} {
		if err := db.Create(challenge).Error; err != nil {
			t.Fatalf("create challenge %s: %v", challenge.Title, err)
		}
	}

	knowledgeTag := &model.Tag{
		Name: "pwn",
		Type: model.TagTypeKnowledge,
	}
	if err := db.Create(knowledgeTag).Error; err != nil {
		t.Fatalf("create knowledge tag: %v", err)
	}
	if err := db.Create(&model.ChallengeTag{
		ChallengeID: tagged.ID,
		TagID:       knowledgeTag.ID,
	}).Error; err != nil {
		t.Fatalf("create challenge tag: %v", err)
	}

	items, err := repo.FindPublishedForRecommendation(context.Background(), 5, []string{"pwn"}, "", []int64{solved.ID})
	if err != nil {
		t.Fatalf("FindPublishedForRecommendation() error = %v", err)
	}

	gotTitles := make([]string, 0, len(items))
	for _, item := range items {
		gotTitles = append(gotTitles, item.Title)
	}

	wantTitles := []string{"Fresh Pwn", "Tagged Web"}
	if !reflect.DeepEqual(gotTitles, wantTitles) {
		t.Fatalf("unexpected recommendation titles: got=%v want=%v", gotTitles, wantTitles)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 recommendation items, got %+v", items)
	}
	if items[0].RecommendationDimension != "pwn" {
		t.Fatalf("expected direct category match to expose pwn recommendation dimension, got %+v", items[0])
	}
	if items[1].RecommendationDimension != "pwn" {
		t.Fatalf("expected tagged recommendation dimension pwn, got %+v", items[1])
	}
}

func TestRepositoryFindPublishedForRecommendationPrefersClosestDifficultyMatch(t *testing.T) {
	db := testsupport.SetupTagTestDB(t)
	repo := NewRepository(db)

	beginner := &model.Challenge{
		Title:      "Pwn Beginner",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyBeginner,
		Points:     90,
		Status:     model.ChallengeStatusPublished,
	}
	easy := &model.Challenge{
		Title:      "Pwn Easy",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     120,
		Status:     model.ChallengeStatusPublished,
	}
	medium := &model.Challenge{
		Title:      "Pwn Medium",
		Category:   "pwn",
		Difficulty: model.ChallengeDifficultyMedium,
		Points:     150,
		Status:     model.ChallengeStatusPublished,
	}
	for _, challenge := range []*model.Challenge{beginner, easy, medium} {
		if err := db.Create(challenge).Error; err != nil {
			t.Fatalf("create challenge %s: %v", challenge.Title, err)
		}
	}

	items, err := repo.FindPublishedForRecommendation(
		context.Background(),
		5,
		[]string{"pwn"},
		model.ChallengeDifficultyEasy,
		nil,
	)
	if err != nil {
		t.Fatalf("FindPublishedForRecommendation() error = %v", err)
	}

	gotTitles := make([]string, 0, len(items))
	for _, item := range items {
		gotTitles = append(gotTitles, item.Title)
	}

	wantTitles := []string{"Pwn Easy", "Pwn Beginner", "Pwn Medium"}
	if !reflect.DeepEqual(gotTitles, wantTitles) {
		t.Fatalf("unexpected recommendation titles: got=%v want=%v", gotTitles, wantTitles)
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

func TestRepositoryCreateAndListAWDChallenges(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := NewRepository(db)

	template := &model.AWDChallenge{
		Name:           "Bank Portal AWD",
		Slug:           "bank-portal-awd",
		Category:       "web",
		Difficulty:     "hard",
		ServiceType:    model.AWDServiceTypeWebHTTP,
		DeploymentMode: model.AWDDeploymentModeSingleContainer,
		Status:         model.AWDChallengeStatusDraft,
	}

	if err := repo.CreateAWDChallenge(context.Background(), template); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if template.ID == 0 {
		t.Fatal("template ID should be set")
	}

	items, total, err := repo.ListAWDChallenges(context.Background(), &dto.AWDChallengeQuery{
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
