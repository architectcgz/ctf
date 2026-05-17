package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/teaching/evidence"
)

type studentReviewRepoStub struct {
	findUserByIDFn             func(ctx context.Context, userID int64) (*model.User, error)
	countPublishedChallengesFn func(ctx context.Context) (int64, error)
	countSolvedChallengesFn    func(ctx context.Context, userID int64) (int64, error)
	getCategoryProgressFn      func(ctx context.Context, userID int64) ([]queryports.ProgressRow, error)
	getDifficultyProgressFn    func(ctx context.Context, userID int64) ([]queryports.ProgressRow, error)
	getStudentTimelineFn       func(ctx context.Context, userID int64, limit, offset int) ([]queryports.TimelineEventRecord, error)
	getStudentEvidenceFn       func(ctx context.Context, userID int64, query evidence.Query) ([]queryports.EvidenceEventRecord, error)
}

func (s *studentReviewRepoStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *studentReviewRepoStub) CountPublishedChallenges(ctx context.Context) (int64, error) {
	if s.countPublishedChallengesFn != nil {
		return s.countPublishedChallengesFn(ctx)
	}
	return 0, nil
}

func (s *studentReviewRepoStub) CountSolvedChallenges(ctx context.Context, userID int64) (int64, error) {
	if s.countSolvedChallengesFn != nil {
		return s.countSolvedChallengesFn(ctx, userID)
	}
	return 0, nil
}

func (s *studentReviewRepoStub) GetCategoryProgress(ctx context.Context, userID int64) ([]queryports.ProgressRow, error) {
	if s.getCategoryProgressFn != nil {
		return s.getCategoryProgressFn(ctx, userID)
	}
	return []queryports.ProgressRow{}, nil
}

func (s *studentReviewRepoStub) GetDifficultyProgress(ctx context.Context, userID int64) ([]queryports.ProgressRow, error) {
	if s.getDifficultyProgressFn != nil {
		return s.getDifficultyProgressFn(ctx, userID)
	}
	return []queryports.ProgressRow{}, nil
}

func (s *studentReviewRepoStub) GetStudentTimeline(ctx context.Context, userID int64, limit, offset int) ([]queryports.TimelineEventRecord, error) {
	if s.getStudentTimelineFn != nil {
		return s.getStudentTimelineFn(ctx, userID, limit, offset)
	}
	return []queryports.TimelineEventRecord{}, nil
}

func (s *studentReviewRepoStub) GetStudentEvidence(ctx context.Context, userID int64, query evidence.Query) ([]queryports.EvidenceEventRecord, error) {
	if s.getStudentEvidenceFn != nil {
		return s.getStudentEvidenceFn(ctx, userID, query)
	}
	return []queryports.EvidenceEventRecord{}, nil
}

type studentReviewRecommendationStub struct {
	recommendFn func(ctx context.Context, userID int64, limit int) (*assessmentcontracts.Recommendation, error)
}

func (s *studentReviewRecommendationStub) Recommend(ctx context.Context, userID int64, limit int) (*assessmentcontracts.Recommendation, error) {
	if s.recommendFn != nil {
		return s.recommendFn(ctx, userID, limit)
	}
	return nil, nil
}

func TestStudentReviewQueryServiceGetStudentProgressUsesAccessibleStudent(t *testing.T) {
	t.Parallel()

	repo := &studentReviewRepoStub{
		findUserByIDFn: func(_ context.Context, userID int64) (*model.User, error) {
			switch userID {
			case 11:
				return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			case 101:
				return &model.User{ID: 101, Role: model.RoleStudent, ClassName: "Class A"}, nil
			default:
				return nil, nil
			}
		},
		countPublishedChallengesFn: func(context.Context) (int64, error) {
			return 20, nil
		},
		countSolvedChallengesFn: func(context.Context, int64) (int64, error) {
			return 5, nil
		},
		getCategoryProgressFn: func(context.Context, int64) ([]queryports.ProgressRow, error) {
			return []queryports.ProgressRow{{Key: "web", Total: 10, Solved: 3}}, nil
		},
		getDifficultyProgressFn: func(context.Context, int64) ([]queryports.ProgressRow, error) {
			return []queryports.ProgressRow{{Key: "easy", Total: 8, Solved: 4}}, nil
		},
	}

	service := NewStudentReviewService(repo, repo, nil)

	progress, err := service.GetStudentProgress(context.Background(), 11, model.RoleTeacher, 101)
	if err != nil {
		t.Fatalf("GetStudentProgress() error = %v", err)
	}
	if progress.TotalChallenges != 20 || progress.SolvedChallenges != 5 {
		t.Fatalf("progress totals = %+v, want challenges=20 solved=5", progress)
	}
	if progress.ByCategory["web"].Solved != 3 {
		t.Fatalf("category breakdown = %+v, want web solved=3", progress.ByCategory)
	}
	if progress.ByDifficulty["easy"].Total != 8 {
		t.Fatalf("difficulty breakdown = %+v, want easy total=8", progress.ByDifficulty)
	}
}

func TestStudentReviewQueryServiceGetStudentRecommendationsMapsResult(t *testing.T) {
	t.Parallel()

	repo := &studentReviewRepoStub{
		findUserByIDFn: func(_ context.Context, userID int64) (*model.User, error) {
			switch userID {
			case 11:
				return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			case 101:
				return &model.User{ID: 101, Role: model.RoleStudent, ClassName: "Class A"}, nil
			default:
				return nil, nil
			}
		},
	}
	recommendations := &studentReviewRecommendationStub{
		recommendFn: func(_ context.Context, userID int64, limit int) (*assessmentcontracts.Recommendation, error) {
			if userID != 101 || limit != 3 {
				return nil, nil
			}
			return &assessmentcontracts.Recommendation{
				WeakDimensions: []assessmentcontracts.RecommendationWeakDimension{
					{Dimension: "web", Severity: "medium", Confidence: 0.3},
				},
				Challenges: []*assessmentcontracts.ChallengeRecommendation{
					{ID: 7, Title: "web-101", Category: "web", Difficulty: "easy"},
				},
			}, nil
		},
	}

	service := NewStudentReviewService(repo, repo, recommendations)

	resp, err := service.GetStudentRecommendations(context.Background(), 11, model.RoleTeacher, 101, 3)
	if err != nil {
		t.Fatalf("GetStudentRecommendations() error = %v", err)
	}
	if len(resp.WeakDimensions) != 1 || resp.WeakDimensions[0].Dimension != "web" {
		t.Fatalf("WeakDimensions = %+v, want web suggestion", resp.WeakDimensions)
	}
	if len(resp.Challenges) != 1 || resp.Challenges[0].Title != "web-101" {
		t.Fatalf("Challenges = %+v, want mapped web-101 recommendation", resp.Challenges)
	}
}

func TestStudentReviewQueryServiceGetStudentTimelineMapsFields(t *testing.T) {
	t.Parallel()

	correct := true
	points := 120
	timestamp := time.Date(2026, 5, 12, 9, 30, 0, 0, time.UTC)
	repo := &studentReviewRepoStub{
		findUserByIDFn: func(_ context.Context, userID int64) (*model.User, error) {
			switch userID {
			case 11:
				return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			case 101:
				return &model.User{ID: 101, Role: model.RoleStudent, ClassName: "Class A"}, nil
			default:
				return nil, nil
			}
		},
		getStudentTimelineFn: func(_ context.Context, userID int64, limit, offset int) ([]queryports.TimelineEventRecord, error) {
			if userID != 101 || limit != 50 || offset != 5 {
				t.Fatalf("unexpected timeline query user=%d limit=%d offset=%d", userID, limit, offset)
			}
			return []queryports.TimelineEventRecord{{
				Type:        "flag_submit",
				ChallengeID: 8,
				Title:       "web-101",
				Timestamp:   timestamp,
				IsCorrect:   &correct,
				Points:      &points,
				Detail:      "submit flag",
			}}, nil
		},
	}

	service := NewStudentReviewService(repo, repo, nil)

	resp, err := service.GetStudentTimeline(context.Background(), 11, model.RoleTeacher, 101, 50, 5)
	if err != nil {
		t.Fatalf("GetStudentTimeline() error = %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("expected one timeline event, got %+v", resp.Events)
	}
	event := resp.Events[0]
	if event.ChallengeID != 8 || event.Title != "web-101" || event.Timestamp != timestamp {
		t.Fatalf("unexpected timeline event identity fields: %+v", event)
	}
	if event.IsCorrect == nil || !*event.IsCorrect {
		t.Fatalf("expected is_correct=true, got %+v", event.IsCorrect)
	}
	if event.Points == nil || *event.Points != points {
		t.Fatalf("expected points=%d, got %+v", points, event.Points)
	}
}

func TestStudentReviewQueryServiceGetStudentTimelineNormalizesNilEventsToEmptySlice(t *testing.T) {
	t.Parallel()

	repo := &studentReviewRepoStub{
		findUserByIDFn: func(_ context.Context, userID int64) (*model.User, error) {
			switch userID {
			case 11:
				return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			case 101:
				return &model.User{ID: 101, Role: model.RoleStudent, ClassName: "Class A"}, nil
			default:
				return nil, nil
			}
		},
		getStudentTimelineFn: func(_ context.Context, userID int64, limit, offset int) ([]queryports.TimelineEventRecord, error) {
			return nil, nil
		},
	}

	service := NewStudentReviewService(repo, repo, nil)

	resp, err := service.GetStudentTimeline(context.Background(), 11, model.RoleTeacher, 101, 20, 0)
	if err != nil {
		t.Fatalf("GetStudentTimeline() error = %v", err)
	}
	if resp.Events == nil {
		t.Fatalf("expected empty events slice, got nil")
	}
	if len(resp.Events) != 0 {
		t.Fatalf("expected no timeline events, got %+v", resp.Events)
	}
}

func TestStudentReviewQueryServiceGetStudentAttackSessionsBuildsSummary(t *testing.T) {
	t.Parallel()

	withEvents := false
	start := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	repo := &studentReviewRepoStub{
		findUserByIDFn: func(_ context.Context, userID int64) (*model.User, error) {
			switch userID {
			case 11:
				return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
			case 101:
				return &model.User{ID: 101, Role: model.RoleStudent, ClassName: "Class A"}, nil
			default:
				return nil, nil
			}
		},
		getStudentEvidenceFn: func(_ context.Context, userID int64, query evidence.Query) ([]queryports.EvidenceEventRecord, error) {
			if userID != 101 || query.ChallengeID != nil || query.ContestID != nil || query.RoundID != nil {
				t.Fatalf("unexpected evidence query = %+v for user=%d", query, userID)
			}
			return []queryports.EvidenceEventRecord{
				{Type: "instance_proxy_request", ChallengeID: 8, Title: "web-101", Timestamp: start, Detail: "open proxy"},
				{Type: "challenge_submission", ChallengeID: 8, Title: "web-101", Timestamp: start.Add(2 * time.Minute), Detail: "submit flag", Meta: map[string]any{"is_correct": true}},
			}, nil
		},
	}

	service := NewStudentReviewService(repo, repo, nil)

	resp, err := service.GetStudentAttackSessions(context.Background(), 11, model.RoleTeacher, 101, &TeacherAttackSessionInput{
		WithEvents: &withEvents,
	})
	if err != nil {
		t.Fatalf("GetStudentAttackSessions() error = %v", err)
	}
	if resp.Summary.TotalSessions != 1 || resp.Summary.SuccessCount != 1 || resp.Summary.EventCount != 2 {
		t.Fatalf("summary = %+v, want one successful session with two events", resp.Summary)
	}
	if len(resp.Sessions) != 1 || resp.Sessions[0].Result != "success" {
		t.Fatalf("sessions = %+v, want one successful session", resp.Sessions)
	}
	if resp.Sessions[0].Events != nil {
		t.Fatalf("events = %+v, want hidden events when WithEvents=false", resp.Sessions[0].Events)
	}
}
