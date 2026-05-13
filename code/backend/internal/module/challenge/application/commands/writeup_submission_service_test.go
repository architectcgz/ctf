package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestWriteupServiceUpsertSubmissionCommunityLifecycle(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	teacher := &model.User{
		Username:  "teacher_a",
		Role:      model.RoleTeacher,
		ClassName: "ClassA",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := teacher.SetPassword("Password123"); err != nil {
		t.Fatalf("set teacher password: %v", err)
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}
	student := &model.User{
		Username:  "student_a",
		Role:      model.RoleStudent,
		ClassName: "ClassA",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := student.SetPassword("Password123"); err != nil {
		t.Fatalf("set student password: %v", err)
	}
	if err := db.Create(student).Error; err != nil {
		t.Fatalf("create student: %v", err)
	}
	challengeItem := &model.Challenge{
		Title:       "web-301",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyEasy,
		Points:      100,
		ImageID:     1,
		Status:      model.ChallengeStatusPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	writeupRepo := challengeinfra.NewWriteupServiceRepository(repo)
	service := NewWriteupService(writeupRepo)
	queryService := challengeqry.NewWriteupService(writeupRepo)

	emptyMine, err := queryService.GetMySubmission(context.Background(), student.ID, challengeItem.ID)
	if err != nil {
		t.Fatalf("GetMySubmission() before upsert error = %v", err)
	}
	if emptyMine != nil {
		t.Fatalf("expected nil submission before upsert, got %+v", emptyMine)
	}

	draft, err := service.UpsertSubmission(context.Background(), challengeItem.ID, student.ID, UpsertSubmissionWriteupInput{
		Title:            "草稿版解题记录",
		Content:          "先枚举路由，再找注入点",
		SubmissionStatus: model.SubmissionWriteupStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission draft error = %v", err)
	}
	if draft.SubmissionStatus != model.SubmissionWriteupStatusDraft || draft.PublishedAt != nil {
		t.Fatalf("unexpected draft submission: %+v", draft)
	}
	if draft.VisibilityStatus != model.SubmissionWriteupVisibilityVisible {
		t.Fatalf("unexpected draft visibility status: %+v", draft)
	}

	if _, err := service.UpsertSubmission(context.Background(), challengeItem.ID, student.ID, UpsertSubmissionWriteupInput{
		Title:            "未解题直接发布",
		Content:          "这一步应该被拦住",
		SubmissionStatus: model.SubmissionWriteupStatusPublished,
	}); err == nil {
		t.Fatalf("expected publish before solve to be forbidden")
	}

	solvedSubmission := &model.Submission{
		UserID:       student.ID,
		ChallengeID:  challengeItem.ID,
		IsCorrect:    true,
		ReviewStatus: model.SubmissionReviewStatusNotRequired,
		Score:        challengeItem.Points,
		SubmittedAt:  now,
		UpdatedAt:    now,
	}
	if err := db.Create(solvedSubmission).Error; err != nil {
		t.Fatalf("create solved submission: %v", err)
	}

	published, err := service.UpsertSubmission(context.Background(), challengeItem.ID, student.ID, UpsertSubmissionWriteupInput{
		Title:            "正式版解题记录",
		Content:          "1. 枚举接口\n2. 找到注入点\n3. 读取 flag",
		SubmissionStatus: model.SubmissionWriteupStatusPublished,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission published error = %v", err)
	}
	if published.SubmissionStatus != model.SubmissionWriteupStatusPublished || published.PublishedAt == nil {
		t.Fatalf("unexpected published writeup: %+v", published)
	}
	if published.VisibilityStatus != model.SubmissionWriteupVisibilityVisible {
		t.Fatalf("unexpected published visibility status: %+v", published)
	}

	mine, err := queryService.GetMySubmission(context.Background(), student.ID, challengeItem.ID)
	if err != nil {
		t.Fatalf("GetMySubmission() error = %v", err)
	}
	if mine.PublishedAt == nil || mine.SubmissionStatus != model.SubmissionWriteupStatusPublished {
		t.Fatalf("unexpected my published submission payload: %+v", mine)
	}

	detail, err := queryService.GetTeacherSubmission(context.Background(), published.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("GetTeacherSubmission() error = %v", err)
	}
	if detail.StudentUsername != student.Username || detail.ChallengeTitle != challengeItem.Title {
		t.Fatalf("unexpected teacher detail: %+v", detail)
	}
	if detail.PublishedAt == nil || detail.IsRecommended {
		t.Fatalf("unexpected teacher community detail: %+v", detail)
	}
}

func TestWriteupServiceCommunityModerationAndOfficialRecommendation(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	admin := &model.User{Username: "admin_a", Role: model.RoleAdmin, Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := admin.SetPassword("Password123"); err != nil {
		t.Fatalf("set admin password: %v", err)
	}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	teacher := &model.User{Username: "teacher_a", Role: model.RoleTeacher, ClassName: "ClassA", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := teacher.SetPassword("Password123"); err != nil {
		t.Fatalf("set teacher password: %v", err)
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("create teacher: %v", err)
	}

	otherTeacher := &model.User{Username: "teacher_b", Role: model.RoleTeacher, ClassName: "ClassB", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := otherTeacher.SetPassword("Password123"); err != nil {
		t.Fatalf("set other teacher password: %v", err)
	}
	if err := db.Create(otherTeacher).Error; err != nil {
		t.Fatalf("create other teacher: %v", err)
	}

	student := &model.User{Username: "student_a", Role: model.RoleStudent, ClassName: "ClassA", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := student.SetPassword("Password123"); err != nil {
		t.Fatalf("set student password: %v", err)
	}
	if err := db.Create(student).Error; err != nil {
		t.Fatalf("create student: %v", err)
	}

	challengeItem := &model.Challenge{
		Title:       "web-302",
		Description: "desc",
		Category:    model.DimensionWeb,
		Difficulty:  model.ChallengeDifficultyEasy,
		Points:      100,
		ImageID:     1,
		Status:      model.ChallengeStatusPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(challengeItem).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	repo := challengeinfra.NewRepository(db)
	writeupRepo := challengeinfra.NewWriteupServiceRepository(repo)
	service := NewWriteupService(writeupRepo)

	if _, err := service.Upsert(context.Background(), challengeItem.ID, admin.ID, UpsertOfficialWriteupInput{
		Title:      "Official",
		Content:    "official content",
		Visibility: model.WriteupVisibilityPublic,
	}); err != nil {
		t.Fatalf("Upsert official writeup() error = %v", err)
	}

	if err := db.Create(&model.Submission{
		UserID:       student.ID,
		ChallengeID:  challengeItem.ID,
		IsCorrect:    true,
		ReviewStatus: model.SubmissionReviewStatusNotRequired,
		Score:        challengeItem.Points,
		SubmittedAt:  now,
		UpdatedAt:    now,
	}).Error; err != nil {
		t.Fatalf("create solved submission: %v", err)
	}

	published, err := service.UpsertSubmission(context.Background(), challengeItem.ID, student.ID, UpsertSubmissionWriteupInput{
		Title:            "社区题解",
		Content:          "community content",
		SubmissionStatus: model.SubmissionWriteupStatusPublished,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission published error = %v", err)
	}

	official, err := service.RecommendOfficial(context.Background(), challengeItem.ID, admin.ID)
	if err != nil {
		t.Fatalf("RecommendOfficial() error = %v", err)
	}
	if !official.IsRecommended {
		t.Fatalf("expected official writeup to be recommended, got %+v", official)
	}

	community, err := service.RecommendCommunity(context.Background(), published.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("RecommendCommunity() error = %v", err)
	}
	if !community.IsRecommended {
		t.Fatalf("expected community writeup to be recommended, got %+v", community)
	}

	if _, err := service.HideCommunity(context.Background(), published.ID, otherTeacher.ID, model.RoleTeacher); err == nil {
		t.Fatalf("expected teacher from another class to be forbidden")
	}

	hidden, err := service.HideCommunity(context.Background(), published.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("HideCommunity() error = %v", err)
	}
	if hidden.VisibilityStatus != model.SubmissionWriteupVisibilityHidden {
		t.Fatalf("expected hidden community writeup, got %+v", hidden)
	}

	restored, err := service.RestoreCommunity(context.Background(), published.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("RestoreCommunity() error = %v", err)
	}
	if restored.VisibilityStatus != model.SubmissionWriteupVisibilityVisible {
		t.Fatalf("expected restored community writeup, got %+v", restored)
	}

	unrecommendedCommunity, err := service.UnrecommendCommunity(context.Background(), published.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("UnrecommendCommunity() error = %v", err)
	}
	if unrecommendedCommunity.IsRecommended {
		t.Fatalf("expected community writeup recommendation to be cleared, got %+v", unrecommendedCommunity)
	}

	unrecommendedOfficial, err := service.UnrecommendOfficial(context.Background(), challengeItem.ID, admin.ID)
	if err != nil {
		t.Fatalf("UnrecommendOfficial() error = %v", err)
	}
	if unrecommendedOfficial.IsRecommended {
		t.Fatalf("expected official writeup recommendation to be cleared, got %+v", unrecommendedOfficial)
	}
}
