package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
)

type stubTeachingQueryService struct{}

func (stubTeachingQueryService) ListClasses(context.Context, int64, string, *teachingqueryqueries.TeacherClassListInput) ([]teachingqueryqueries.TeacherClassItem, int64, int, int, error) {
	return nil, 0, 0, 0, nil
}

func (stubTeachingQueryService) ListStudents(context.Context, int64, string, *teachingqueryqueries.TeacherStudentDirectoryInput) ([]teachingqueryqueries.TeacherStudentItem, int64, int, int, error) {
	return nil, 0, 0, 0, nil
}

func (stubTeachingQueryService) ListClassStudents(context.Context, int64, string, string, *teachingqueryqueries.TeacherStudentListInput) ([]teachingqueryqueries.TeacherStudentItem, error) {
	return nil, nil
}

type stubTeachingOverviewService struct{}

func (stubTeachingOverviewService) GetOverview(context.Context, int64, string) (*teachingqueryqueries.TeacherOverviewResp, error) {
	return nil, nil
}

type stubTeachingClassInsightService struct{}

func (stubTeachingClassInsightService) GetClassSummary(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassSummary, error) {
	return nil, nil
}

func (stubTeachingClassInsightService) GetClassTrend(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassTrend, error) {
	return nil, nil
}

func (stubTeachingClassInsightService) GetClassReview(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassReview, error) {
	return nil, nil
}

type captureTeachingClassInsightService struct {
	lastQuery *teachingqueryqueries.TeacherClassInsightInput
}

func (s *captureTeachingClassInsightService) GetClassSummary(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassSummary, error) {
	return nil, nil
}

func (s *captureTeachingClassInsightService) GetClassTrend(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassTrend, error) {
	return nil, nil
}

func (s *captureTeachingClassInsightService) GetClassReview(_ context.Context, _ int64, _ string, _ string, query *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassReview, error) {
	s.lastQuery = query
	return &teachingqueryqueries.TeacherClassReview{}, nil
}

type responseTeachingClassInsightService struct{}

func (responseTeachingClassInsightService) GetClassSummary(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassSummary, error) {
	return &teachingqueryqueries.TeacherClassSummary{
		ClassName:          "Class A",
		StudentCount:       3,
		AverageSolved:      2.5,
		ActiveStudentCount: 2,
		ActiveRate:         66.7,
		RecentEventCount:   9,
	}, nil
}

func (responseTeachingClassInsightService) GetClassTrend(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassTrend, error) {
	return nil, nil
}

func (responseTeachingClassInsightService) GetClassReview(context.Context, int64, string, string, *teachingqueryqueries.TeacherClassInsightInput) (*teachingqueryqueries.TeacherClassReview, error) {
	name := "Alice"
	return &teachingqueryqueries.TeacherClassReview{
		ClassName: "Class A",
		Items: []teachingqueryqueries.TeacherClassReviewItem{
			{
				Code:      "weak_dimension_cluster",
				Severity:  "warning",
				Summary:   "web 基础薄弱",
				Dimension: "web",
				Students: []teachingqueryqueries.TeacherReviewStudentRef{
					{ID: 1, Username: "alice", Name: &name},
				},
				Recommendation: &teachingqueryqueries.TeacherRecommendationItem{
					ChallengeID: 101,
					Title:       "web-101",
					Category:    "web",
					Difficulty:  "easy",
					Summary:     "先补基础命中率",
				},
			},
		},
	}, nil
}

type captureTeachingStudentReviewService struct {
	lastEvidenceQuery      *teachingqueryqueries.TeacherEvidenceInput
	lastAttackSessionQuery *teachingqueryqueries.TeacherAttackSessionInput
}

func (s *captureTeachingStudentReviewService) GetStudentProgress(context.Context, int64, string, int64) (*teachingqueryqueries.TeacherProgressResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentRecommendations(context.Context, int64, string, int64, int) (*teachingqueryqueries.TeacherRecommendationResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentTimeline(context.Context, int64, string, int64, int, int) (*teachingqueryqueries.TimelineResp, error) {
	return nil, nil
}

func (s *captureTeachingStudentReviewService) GetStudentEvidence(_ context.Context, _ int64, _ string, _ int64, query *teachingqueryqueries.TeacherEvidenceInput) (*teachingqueryqueries.TeacherEvidenceResp, error) {
	s.lastEvidenceQuery = query
	return &teachingqueryqueries.TeacherEvidenceResp{}, nil
}

func (s *captureTeachingStudentReviewService) GetStudentAttackSessions(_ context.Context, _ int64, _ string, _ int64, query *teachingqueryqueries.TeacherAttackSessionInput) (*teachingqueryqueries.TeacherAttackSessionResp, error) {
	s.lastAttackSessionQuery = query
	return &teachingqueryqueries.TeacherAttackSessionResp{}, nil
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

func TestGetClassSummaryKeepsTeacherHTTPContract(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	handler := NewHandler(stubTeachingQueryService{}, stubTeachingOverviewService{}, responseTeachingClassInsightService{}, &captureTeachingStudentReviewService{})

	router := gin.New()
	router.GET("/api/v1/teacher/classes/:name/summary", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "teacher", Role: model.RoleTeacher})
		handler.GetClassSummary(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/classes/Class%20A/summary", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Data teachingqueryqueries.TeacherClassSummary `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Data.ClassName != "Class A" || body.Data.StudentCount != 3 || body.Data.RecentEventCount != 9 {
		t.Fatalf("unexpected summary payload = %+v", body.Data)
	}
}

func TestGetClassReviewKeepsTeacherHTTPContract(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	handler := NewHandler(stubTeachingQueryService{}, stubTeachingOverviewService{}, responseTeachingClassInsightService{}, &captureTeachingStudentReviewService{})

	router := gin.New()
	router.GET("/api/v1/teacher/classes/:name/review", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "teacher", Role: model.RoleTeacher})
		handler.GetClassReview(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/classes/Class%20A/review", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Data teachingqueryqueries.TeacherClassReview `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Data.ClassName != "Class A" || len(body.Data.Items) != 1 {
		t.Fatalf("unexpected review payload = %+v", body.Data)
	}
	if got := body.Data.Items[0]; got.Code != "weak_dimension_cluster" || got.Recommendation == nil || got.Recommendation.Title != "web-101" {
		t.Fatalf("unexpected review item = %+v", got)
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
