package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
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
	recommendFn func(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error)
}

func (s *studentReviewRecommendationStub) Recommend(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
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
		recommendFn: func(_ context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
			if userID != 101 || limit != 3 {
				return nil, nil
			}
			return &dto.RecommendationResp{
				WeakDimensions: []dto.RecommendationWeakDimension{
					{Dimension: "web", Severity: "medium", Confidence: 0.3},
				},
				Challenges: []*dto.ChallengeRecommendation{
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

	resp, err := service.GetStudentAttackSessions(context.Background(), 11, model.RoleTeacher, 101, &dto.TeacherAttackSessionQuery{
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
