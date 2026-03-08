package teacher

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type stubRecommendationProvider struct {
	resp  *dto.RecommendationResp
	err   error
	calls []int64
}

func (s *stubRecommendationProvider) Recommend(userID int64, _ int) (*dto.RecommendationResp, error) {
	s.calls = append(s.calls, userID)
	return s.resp, s.err
}

func setupTeacherTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Submission{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	now := time.Now()
	users := []model.User{
		{ID: 1, Username: "teacher-a", TeacherNo: "T-1001", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Username: "student-a", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Username: "student-b", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 4, Username: "admin", Role: model.RoleAdmin, Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	challenges := []model.Challenge{
		{ID: 11, Title: "web-1", Category: "web", Difficulty: "easy", Points: 100, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 12, Title: "pwn-1", Category: "pwn", Difficulty: "medium", Points: 200, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge: %v", err)
		}
	}

	submissions := []model.Submission{
		{UserID: 2, ChallengeID: 11, IsCorrect: true, SubmittedAt: now},
		{UserID: 3, ChallengeID: 12, IsCorrect: true, SubmittedAt: now},
	}
	for _, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}

	return db
}

func TestServiceListClassesTeacherScoped(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	items, err := service.ListClasses(context.Background(), 1, model.RoleTeacher)
	if err != nil {
		t.Fatalf("ListClasses() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "Class A" || items[0].StudentCount != 1 {
		t.Fatalf("unexpected classes: %+v", items)
	}
}

func TestServiceListClassStudentsFiltersByStudentNo(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	items, err := service.ListClassStudents(context.Background(), 1, model.RoleTeacher, "Class A", &dto.TeacherStudentQuery{
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListClassStudents() error = %v", err)
	}
	if len(items) != 1 || items[0].Username != "student-a" {
		t.Fatalf("unexpected students: %+v", items)
	}
	if items[0].StudentNo == nil || *items[0].StudentNo != "S-1001" {
		t.Fatalf("expected student no to be returned, got %+v", items[0])
	}
}

func TestServiceGetStudentProgressForbiddenForOtherClass(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	_, err := service.GetStudentProgress(context.Background(), 1, model.RoleTeacher, 3)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceGetStudentProgressForAdmin(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	progress, err := service.GetStudentProgress(context.Background(), 4, model.RoleAdmin, 3)
	if err != nil {
		t.Fatalf("GetStudentProgress() error = %v", err)
	}
	if progress.TotalChallenges != 2 || progress.SolvedChallenges != 1 {
		t.Fatalf("unexpected progress summary: %+v", progress)
	}
	if progress.ByDifficulty["medium"].Solved != 1 {
		t.Fatalf("unexpected difficulty summary: %+v", progress.ByDifficulty)
	}
}

func TestServiceGetStudentRecommendationsMapsResponse(t *testing.T) {
	db := setupTeacherTestDB(t)
	reco := &stubRecommendationProvider{
		resp: &dto.RecommendationResp{
			Challenges: []*dto.ChallengeRecommendation{
				{
					ID:         12,
					Title:      "pwn-1",
					Category:   "pwn",
					Difficulty: "medium",
					Reason:     "针对薄弱维度：PWN",
				},
			},
		},
	}
	service := NewService(NewRepository(db), reco, nil)

	items, err := service.GetStudentRecommendations(context.Background(), 1, model.RoleTeacher, 2, 6)
	if err != nil {
		t.Fatalf("GetStudentRecommendations() error = %v", err)
	}
	if len(items) != 1 || items[0].ChallengeID != 12 || len(reco.calls) != 1 || reco.calls[0] != 2 {
		t.Fatalf("unexpected recommendation result: %+v calls=%v", items, reco.calls)
	}
}
