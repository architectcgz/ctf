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

	if err := db.AutoMigrate(
		&model.User{},
		&model.Challenge{},
		&model.Submission{},
		&model.Instance{},
		&model.SkillProfile{},
		&model.AuditLog{},
		&model.ChallengeHint{},
		&model.ChallengeHintUnlock{},
	); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	now := time.Now()
	users := []model.User{
		{ID: 1, Username: "teacher-a", TeacherNo: "T-1001", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Username: "alice", Name: "Alice Zhang", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Username: "bob", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
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

	profiles := []model.SkillProfile{
		{UserID: 2, Dimension: model.DimensionCrypto, Score: 0.2, UpdatedAt: now},
		{UserID: 2, Dimension: model.DimensionWeb, Score: 0.7, UpdatedAt: now},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("seed skill profile: %v", err)
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
	if len(items) != 1 || items[0].Username != "alice" {
		t.Fatalf("unexpected students: %+v", items)
	}
	if items[0].Name == nil || *items[0].Name != "Alice Zhang" {
		t.Fatalf("expected name in response, got %+v", items[0])
	}
	if items[0].StudentNo == nil || *items[0].StudentNo != "S-1001" {
		t.Fatalf("expected student no to be returned, got %+v", items[0])
	}
	if items[0].SolvedCount != 1 || items[0].TotalScore != 100 {
		t.Fatalf("expected solved/score summary, got %+v", items[0])
	}
	if items[0].RecentEventCount != 1 {
		t.Fatalf("expected recent event count to be returned, got %+v", items[0])
	}
	if items[0].WeakDimension == nil || *items[0].WeakDimension != model.DimensionCrypto {
		t.Fatalf("expected weak dimension to be returned, got %+v", items[0])
	}
}

func TestServiceListClassStudentsFiltersByKeyword(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	items, err := service.ListClassStudents(context.Background(), 1, model.RoleTeacher, "Class A", &dto.TeacherStudentQuery{
		Keyword: "zhang",
	})
	if err != nil {
		t.Fatalf("ListClassStudents() error = %v", err)
	}
	if len(items) != 1 || items[0].Username != "alice" {
		t.Fatalf("unexpected students: %+v", items)
	}
}

func TestServiceGetClassSummaryForTeacher(t *testing.T) {
	db := setupTeacherTestDB(t)
	now := time.Now()

	events := []model.Submission{
		{UserID: 2, ChallengeID: 11, IsCorrect: true, SubmittedAt: now.Add(-2 * time.Hour)},
		{UserID: 2, ChallengeID: 12, IsCorrect: false, SubmittedAt: now.Add(-90 * time.Minute)},
	}
	for _, item := range events {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed summary submission: %v", err)
		}
	}
	instance := model.Instance{
		ID:          31,
		UserID:      2,
		ChallengeID: 11,
		Status:      model.InstanceStatusStopped,
		CreatedAt:   now.Add(-3 * time.Hour),
		UpdatedAt:   now.Add(-30 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed summary instance: %v", err)
	}
	hint := model.ChallengeHint{
		ID:          51,
		ChallengeID: 11,
		Level:       1,
		Title:       "先看注入点",
		Content:     "try id param",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&hint).Error; err != nil {
		t.Fatalf("seed summary hint: %v", err)
	}
	if err := db.Create(&model.ChallengeHintUnlock{
		UserID:          2,
		ChallengeID:     11,
		ChallengeHintID: hint.ID,
		UnlockedAt:      now.Add(-45 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed summary hint unlock: %v", err)
	}

	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	summary, err := service.GetClassSummary(context.Background(), 1, model.RoleTeacher, "Class A")
	if err != nil {
		t.Fatalf("GetClassSummary() error = %v", err)
	}
	if summary.StudentCount != 1 {
		t.Fatalf("expected student_count=1, got %+v", summary)
	}
	if summary.ActiveStudentCount != 1 {
		t.Fatalf("expected active_student_count=1, got %+v", summary)
	}
	if summary.RecentEventCount != 6 {
		t.Fatalf("expected recent_event_count=6, got %+v", summary)
	}
	if summary.AverageSolved != 1 {
		t.Fatalf("expected average_solved=1, got %+v", summary)
	}
	if summary.ActiveRate != 100 {
		t.Fatalf("expected active_rate=100, got %+v", summary)
	}
}

func TestServiceGetClassSummaryForbiddenForOtherClass(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	_, err := service.GetClassSummary(context.Background(), 1, model.RoleTeacher, "Class B")
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceGetClassTrendForTeacher(t *testing.T) {
	db := setupTeacherTestDB(t)
	now := time.Now()

	records := []model.Submission{
		{UserID: 2, ChallengeID: 11, IsCorrect: true, SubmittedAt: now.Add(-2 * time.Hour)},
		{UserID: 2, ChallengeID: 12, IsCorrect: false, SubmittedAt: now.Add(-26 * time.Hour)},
	}
	for _, item := range records {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed trend submission: %v", err)
		}
	}
	instances := []model.Instance{
		{ID: 41, UserID: 2, ChallengeID: 11, Status: model.InstanceStatusRunning, CreatedAt: now.Add(-90 * time.Minute), UpdatedAt: now.Add(-90 * time.Minute)},
		{ID: 42, UserID: 2, ChallengeID: 11, Status: model.InstanceStatusStopped, CreatedAt: now.Add(-28 * time.Hour), UpdatedAt: now.Add(-25 * time.Hour)},
	}
	for _, item := range instances {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed trend instance: %v", err)
		}
	}
	hint := model.ChallengeHint{
		ID:          61,
		ChallengeID: 11,
		Level:       1,
		Title:       "观察回显",
		Content:     "focus on response",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&hint).Error; err != nil {
		t.Fatalf("seed trend hint: %v", err)
	}
	if err := db.Create(&model.ChallengeHintUnlock{
		UserID:          2,
		ChallengeID:     11,
		ChallengeHintID: hint.ID,
		UnlockedAt:      now.Add(-80 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed trend hint unlock: %v", err)
	}

	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)
	trend, err := service.GetClassTrend(context.Background(), 1, model.RoleTeacher, "Class A")
	if err != nil {
		t.Fatalf("GetClassTrend() error = %v", err)
	}
	if len(trend.Points) != 7 {
		t.Fatalf("expected 7 trend points, got %+v", trend.Points)
	}

	var totalEvents int64
	var totalSolves int64
	var activeDays int
	for _, point := range trend.Points {
		totalEvents += point.EventCount
		totalSolves += point.SolveCount
		if point.ActiveStudentCount > 0 {
			activeDays++
		}
	}
	if totalEvents != 7 {
		t.Fatalf("expected total events=7, got %+v", trend.Points)
	}
	if totalSolves != 2 {
		t.Fatalf("expected total solves=2, got %+v", trend.Points)
	}
	if activeDays < 2 {
		t.Fatalf("expected at least 2 active days, got %+v", trend.Points)
	}
}

func TestServiceGetClassTrendForbiddenForOtherClass(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	_, err := service.GetClassTrend(context.Background(), 1, model.RoleTeacher, "Class B")
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceGetClassReviewBuildsRecommendations(t *testing.T) {
	db := setupTeacherTestDB(t)
	now := time.Now()

	student := model.User{
		ID:        5,
		Username:  "charlie",
		Name:      "Charlie Li",
		StudentNo: "S-1003",
		Role:      model.RoleStudent,
		ClassName: "Class A",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("seed extra student: %v", err)
	}
	if err := db.Create(&model.SkillProfile{
		UserID:    5,
		Dimension: model.DimensionCrypto,
		Score:     0.1,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed extra skill profile: %v", err)
	}

	reco := &stubRecommendationProvider{
		resp: &dto.RecommendationResp{
			Challenges: []*dto.ChallengeRecommendation{
				{
					ID:         11,
					Title:      "web-1",
					Category:   "web",
					Difficulty: "easy",
					Reason:     "适合先补基础训练",
				},
			},
		},
	}
	service := NewService(NewRepository(db), reco, nil)

	review, err := service.GetClassReview(context.Background(), 1, model.RoleTeacher, "Class A")
	if err != nil {
		t.Fatalf("GetClassReview() error = %v", err)
	}
	if review.ClassName != "Class A" {
		t.Fatalf("unexpected class review: %+v", review)
	}
	if len(review.Items) < 3 {
		t.Fatalf("expected at least 3 review items, got %+v", review.Items)
	}
	if review.Items[1].Title != "优先补薄弱维度" || review.Items[1].Recommendation == nil {
		t.Fatalf("expected weak-dimension review recommendation, got %+v", review.Items[1])
	}
	if review.Items[2].Title != "先跟进重点学生" || len(review.Items[2].Students) == 0 {
		t.Fatalf("expected focus-students review item, got %+v", review.Items[2])
	}
	if len(reco.calls) == 0 || reco.calls[0] != 5 {
		t.Fatalf("expected recommendations to target weakest student first, calls=%v", reco.calls)
	}
}

func TestServiceGetClassReviewForbiddenForOtherClass(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	_, err := service.GetClassReview(context.Background(), 1, model.RoleTeacher, "Class B")
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
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

func TestServiceGetStudentTimelineForTeacher(t *testing.T) {
	db := setupTeacherTestDB(t)
	now := time.Now()
	instance := model.Instance{
		ID:          21,
		UserID:      2,
		ChallengeID: 11,
		Status:      model.InstanceStatusRunning,
		CreatedAt:   now.Add(-2 * time.Hour),
		UpdatedAt:   now.Add(-2 * time.Hour),
	}
	submission := model.Submission{
		UserID:      2,
		ChallengeID: 11,
		IsCorrect:   true,
		SubmittedAt: now.Add(-time.Hour),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Create(&submission).Error; err != nil {
		t.Fatalf("seed timeline submission: %v", err)
	}
	hint := model.ChallengeHint{
		ID:          71,
		ChallengeID: 11,
		Level:       1,
		Title:       "先看回显",
		Content:     "payload in query",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&hint).Error; err != nil {
		t.Fatalf("seed timeline hint: %v", err)
	}
	if err := db.Create(&model.ChallengeHintUnlock{
		UserID:          2,
		ChallengeID:     11,
		ChallengeHintID: hint.ID,
		UnlockedAt:      now.Add(-90 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed timeline hint unlock: %v", err)
	}
	instanceExtendDetail := `{"id":"21"}`
	if err := db.Create(&model.AuditLog{
		UserID:       ptrInt64(2),
		Action:       model.AuditActionUpdate,
		ResourceType: "instance",
		ResourceID:   ptrInt64(21),
		Detail:       instanceExtendDetail,
		CreatedAt:    now.Add(-30 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed timeline extend audit: %v", err)
	}
	challengeViewDetail := `{"id":"11"}`
	if err := db.Create(&model.AuditLog{
		UserID:       ptrInt64(2),
		Action:       model.AuditActionRead,
		ResourceType: "challenge_detail",
		ResourceID:   ptrInt64(11),
		Detail:       challengeViewDetail,
		CreatedAt:    now.Add(-150 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed timeline detail-view audit: %v", err)
	}
	if err := db.Create(&model.AuditLog{
		UserID:       ptrInt64(2),
		Action:       model.AuditActionRead,
		ResourceType: "instance_access",
		ResourceID:   ptrInt64(21),
		Detail:       `{"id":"21"}`,
		CreatedAt:    now.Add(-20 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed timeline instance-access audit: %v", err)
	}
	if err := db.Create(&model.AuditLog{
		UserID:       ptrInt64(2),
		Action:       model.AuditActionSubmit,
		ResourceType: "instance_proxy_request",
		ResourceID:   ptrInt64(21),
		Detail:       `{"method":"POST","target_path":"/submit","status":201,"payload_preview":"payload=' OR 1=1 --"}`,
		CreatedAt:    now.Add(-10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("seed timeline proxy audit: %v", err)
	}

	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	timeline, err := service.GetStudentTimeline(context.Background(), 1, model.RoleTeacher, 2, 10, 0)
	if err != nil {
		t.Fatalf("GetStudentTimeline() error = %v", err)
	}
	if len(timeline.Events) != 8 {
		t.Fatalf("expected 8 timeline events, got %+v", timeline.Events)
	}
	var foundSolvedDetail bool
	var foundHintDetail bool
	var foundExtendDetail bool
	var foundChallengeViewDetail bool
	var foundInstanceAccessDetail bool
	var foundProxyTraceDetail bool
	for _, event := range timeline.Events {
		if event.Type == "flag_submit" && event.Title == "web-1" && event.Points != nil && *event.Points == 100 && event.Detail != "" {
			foundSolvedDetail = true
		}
		if event.Type == "hint_unlock" && event.Detail != "" {
			foundHintDetail = true
		}
		if event.Type == "instance_extend" && event.Detail != "" {
			foundExtendDetail = true
		}
		if event.Type == "challenge_detail_view" && event.Detail != "" {
			foundChallengeViewDetail = true
		}
		if event.Type == "instance_access" && event.Detail != "" {
			foundInstanceAccessDetail = true
		}
		if event.Type == "instance_proxy_request" && event.Detail != "" {
			foundProxyTraceDetail = true
		}
	}
	if !foundSolvedDetail {
		t.Fatalf("expected solved timeline detail, got %+v", timeline.Events)
	}
	if !foundHintDetail {
		t.Fatalf("expected hint unlock detail, got %+v", timeline.Events)
	}
	if !foundExtendDetail {
		t.Fatalf("expected instance extend detail, got %+v", timeline.Events)
	}
	if !foundChallengeViewDetail {
		t.Fatalf("expected challenge detail view, got %+v", timeline.Events)
	}
	if !foundInstanceAccessDetail {
		t.Fatalf("expected instance access detail, got %+v", timeline.Events)
	}
	if !foundProxyTraceDetail {
		t.Fatalf("expected proxy trace detail, got %+v", timeline.Events)
	}
}

func ptrInt64(value int64) *int64 {
	return &value
}

func TestServiceGetStudentTimelineForbiddenForOtherClass(t *testing.T) {
	db := setupTeacherTestDB(t)
	service := NewService(NewRepository(db), &stubRecommendationProvider{}, nil)

	_, err := service.GetStudentTimeline(context.Background(), 1, model.RoleTeacher, 3, 10, 0)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
