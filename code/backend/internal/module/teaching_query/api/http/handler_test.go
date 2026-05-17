package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
)

type stubTeachingQueryService struct{}

func (stubTeachingQueryService) ListClasses(context.Context, int64, string, *teachingqueryqueries.TeacherClassListInput) ([]dto.TeacherClassItem, int64, int, int, error) {
	return nil, 0, 0, 0, nil
}

func (stubTeachingQueryService) ListStudents(context.Context, int64, string, *teachingqueryqueries.TeacherStudentDirectoryInput) ([]dto.TeacherStudentItem, int64, int, int, error) {
	return nil, 0, 0, 0, nil
}

func (stubTeachingQueryService) ListClassStudents(context.Context, int64, string, string, *teachingqueryqueries.TeacherStudentListInput) ([]dto.TeacherStudentItem, error) {
	return nil, nil
}

type stubTeachingOverviewService struct{}

func (stubTeachingOverviewService) GetOverview(context.Context, int64, string) (*dto.TeacherOverviewResp, error) {
	return nil, nil
}

type stubTeachingClassInsightService struct{}

func (stubTeachingClassInsightService) GetClassSummary(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassSummaryResp, error) {
	return nil, nil
}

func (stubTeachingClassInsightService) GetClassTrend(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassTrendResp, error) {
	return nil, nil
}

func (stubTeachingClassInsightService) GetClassReview(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassReviewResp, error) {
	return nil, nil
}

type captureTeachingClassInsightService struct {
	lastQuery *teachingqueryqueries.TeacherClassInsightInput
}

func (s *captureTeachingClassInsightService) GetClassSummary(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassSummaryResp, error) {
	return nil, nil
}

func (s *captureTeachingClassInsightService) GetClassTrend(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassTrendResp, error) {
	return nil, nil
}

func (s *captureTeachingClassInsightService) GetClassReview(_ context.Context, _ int64, _ string, _ string, query *teachingqueryqueries.TeacherClassInsightInput) (*dto.TeacherClassReviewResp, error) {
	s.lastQuery = query
	return &dto.TeacherClassReviewResp{}, nil
}

type captureTeachingStudentReviewService struct {
	lastEvidenceQuery      *teachingqueryqueries.TeacherEvidenceInput
	lastAttackSessionQuery *teachingqueryqueries.TeacherAttackSessionInput
}

func (s *captureTeachingStudentReviewService) GetStudentProgress(context.Context, int64, string, int64) (*dto.TeacherProgressResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentRecommendations(context.Context, int64, string, int64, int) (*dto.TeacherRecommendationResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentTimeline(context.Context, int64, string, int64, int, int) (*teachingqueryqueries.TimelineResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentEvidence(_ context.Context, _ int64, _ string, _ int64, query *teachingqueryqueries.TeacherEvidenceInput) (*dto.TeacherEvidenceResp, error) {
	s.lastEvidenceQuery = query
	return &dto.TeacherEvidenceResp{}, nil
}

func (s *captureTeachingStudentReviewService) GetStudentAttackSessions(_ context.Context, _ int64, _ string, _ int64, query *teachingqueryqueries.TeacherAttackSessionInput) (*dto.TeacherAttackSessionResp, error) {
	s.lastAttackSessionQuery = query
	return &dto.TeacherAttackSessionResp{}, nil
}

func TestGetStudentEvidenceBindsQueryIntoTeachingInput(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	studentReview := &captureTeachingStudentReviewService{}
	handler := NewHandler(stubTeachingQueryService{}, stubTeachingOverviewService{}, stubTeachingClassInsightService{}, studentReview)

	router := gin.New()
	router.GET("/api/v1/teacher/students/:id/evidence", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "teacher", Role: model.RoleTeacher})
		handler.GetStudentEvidence(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/students/101/evidence?challenge_id=8&contest_id=12&round_id=4&event_type=challenge_submission&from=2026-05-01T00:00:00Z&to=2026-05-02T00:00:00Z&limit=10&offset=2", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if studentReview.lastEvidenceQuery == nil {
		t.Fatal("expected evidence query to be passed into student review service")
	}
	if studentReview.lastEvidenceQuery.EventType != "challenge_submission" || studentReview.lastEvidenceQuery.Limit != 10 || studentReview.lastEvidenceQuery.Offset != 2 {
		t.Fatalf("unexpected evidence query = %+v", studentReview.lastEvidenceQuery)
	}
	if studentReview.lastEvidenceQuery.ChallengeID == nil || *studentReview.lastEvidenceQuery.ChallengeID != 8 {
		t.Fatalf("unexpected challenge id = %+v", studentReview.lastEvidenceQuery)
	}
	expectedFrom := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if studentReview.lastEvidenceQuery.From == nil || !studentReview.lastEvidenceQuery.From.Equal(expectedFrom) {
		t.Fatalf("unexpected from = %+v, want %s", studentReview.lastEvidenceQuery.From, expectedFrom)
	}
}

func TestGetClassReviewBindsWindowIntoTeachingInput(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	classInsight := &captureTeachingClassInsightService{}
	handler := NewHandler(stubTeachingQueryService{}, stubTeachingOverviewService{}, classInsight, &captureTeachingStudentReviewService{})

	router := gin.New()
	router.GET("/api/v1/teacher/classes/:name/review", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "teacher", Role: model.RoleTeacher})
		handler.GetClassReview(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/classes/Class%20A/review?from_date=2026-05-01&to_date=2026-05-03", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if classInsight.lastQuery == nil {
		t.Fatal("expected class insight query to be passed into service")
	}
	if classInsight.lastQuery.FromDate != "2026-05-01" || classInsight.lastQuery.ToDate != "2026-05-03" {
		t.Fatalf("unexpected class insight query = %+v", classInsight.lastQuery)
	}
}

func TestGetStudentAttackSessionsBindsQueryIntoTeachingInput(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	studentReview := &captureTeachingStudentReviewService{}
	handler := NewHandler(stubTeachingQueryService{}, stubTeachingOverviewService{}, stubTeachingClassInsightService{}, studentReview)

	router := gin.New()
	router.GET("/api/v1/teacher/students/:id/attack-sessions", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "teacher", Role: model.RoleTeacher})
		handler.GetStudentAttackSessions(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/students/101/attack-sessions?mode=awd&contest_id=12&round_id=4&result=success&with_events=false&limit=5&offset=1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if studentReview.lastAttackSessionQuery == nil {
		t.Fatal("expected attack session query to be passed into student review service")
	}
	if studentReview.lastAttackSessionQuery.Mode != "awd" || studentReview.lastAttackSessionQuery.Result != "success" {
		t.Fatalf("unexpected attack session query = %+v", studentReview.lastAttackSessionQuery)
	}
	if studentReview.lastAttackSessionQuery.WithEvents == nil || *studentReview.lastAttackSessionQuery.WithEvents {
		t.Fatalf("unexpected with_events binding = %+v", studentReview.lastAttackSessionQuery.WithEvents)
	}
}
