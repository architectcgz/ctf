package commands

import (
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestWriteupServiceUpsertSubmissionAndReview(t *testing.T) {
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
	service := NewWriteupService(repo)
	queryService := challengeqry.NewWriteupService(repo)

	draft, err := service.UpsertSubmission(challengeItem.ID, student.ID, &dto.UpsertSubmissionWriteupReq{
		Title:            "草稿版解题记录",
		Content:          "先枚举路由，再找注入点",
		SubmissionStatus: model.SubmissionWriteupStatusDraft,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission draft error = %v", err)
	}
	if draft.SubmissionStatus != model.SubmissionWriteupStatusDraft || draft.SubmittedAt != nil {
		t.Fatalf("unexpected draft submission: %+v", draft)
	}

	submitted, err := service.UpsertSubmission(challengeItem.ID, student.ID, &dto.UpsertSubmissionWriteupReq{
		Title:            "正式版解题记录",
		Content:          "1. 枚举接口\n2. 找到注入点\n3. 读取 flag",
		SubmissionStatus: model.SubmissionWriteupStatusSubmitted,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission submitted error = %v", err)
	}
	if submitted.SubmissionStatus != model.SubmissionWriteupStatusSubmitted || submitted.SubmittedAt == nil {
		t.Fatalf("unexpected submitted writeup: %+v", submitted)
	}
	if submitted.ReviewStatus != model.SubmissionWriteupReviewPending {
		t.Fatalf("unexpected review status after submit: %+v", submitted)
	}

	reviewed, err := service.ReviewSubmission(submitted.ID, teacher.ID, model.RoleTeacher, &dto.ReviewSubmissionWriteupReq{
		ReviewStatus:  model.SubmissionWriteupReviewExcellent,
		ReviewComment: "链路完整，可以作为班级优秀样例。",
	})
	if err != nil {
		t.Fatalf("ReviewSubmission() error = %v", err)
	}
	if reviewed.ReviewStatus != model.SubmissionWriteupReviewExcellent || reviewed.ReviewedAt == nil {
		t.Fatalf("unexpected reviewed submission: %+v", reviewed)
	}

	mine, err := queryService.GetMySubmission(student.ID, challengeItem.ID)
	if err != nil {
		t.Fatalf("GetMySubmission() error = %v", err)
	}
	if mine.ReviewComment != "链路完整，可以作为班级优秀样例。" {
		t.Fatalf("unexpected my submission review comment: %+v", mine)
	}

	detail, err := queryService.GetTeacherSubmission(submitted.ID, teacher.ID, model.RoleTeacher)
	if err != nil {
		t.Fatalf("GetTeacherSubmission() error = %v", err)
	}
	if detail.StudentUsername != student.Username || detail.ChallengeTitle != challengeItem.Title {
		t.Fatalf("unexpected teacher detail: %+v", detail)
	}
}
