package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type classInsightRepoStub struct {
	findUserByIDFn                 func(ctx context.Context, userID int64) (*model.User, error)
	getClassSummaryFn              func(ctx context.Context, className string, since time.Time) (*queryports.ClassSummary, error)
	getClassTrendFn                func(ctx context.Context, className string, since time.Time, days int) (*queryports.ClassTrend, error)
	listClassTeachingFactSnapshots func(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error)
}

func (s *classInsightRepoStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *classInsightRepoStub) GetClassSummary(ctx context.Context, className string, since time.Time) (*queryports.ClassSummary, error) {
	if s.getClassSummaryFn != nil {
		return s.getClassSummaryFn(ctx, className, since)
	}
	return nil, nil
}

func (s *classInsightRepoStub) GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*queryports.ClassTrend, error) {
	if s.getClassTrendFn != nil {
		return s.getClassTrendFn(ctx, className, since, days)
	}
	return &queryports.ClassTrend{}, nil
}

func (s *classInsightRepoStub) ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error) {
	if s.listClassTeachingFactSnapshots != nil {
		return s.listClassTeachingFactSnapshots(ctx, className, since)
	}
	return []teachingadvice.StudentFactSnapshot{}, nil
}

type classInsightRecommendationStub struct {
	recommendFn func(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error)
}

func (s *classInsightRecommendationStub) Recommend(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
	if s.recommendFn != nil {
		return s.recommendFn(ctx, userID, limit)
	}
	return nil, nil
}

func TestClassInsightQueryServiceGetClassSummaryUsesAccessibleClass(t *testing.T) {
	t.Parallel()

	repo := &classInsightRepoStub{
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		getClassSummaryFn: func(context.Context, string, time.Time) (*queryports.ClassSummary, error) {
			return &queryports.ClassSummary{
				ClassName:          "Class A",
				StudentCount:       3,
				AverageSolved:      2.5,
				ActiveStudentCount: 2,
				ActiveRate:         66,
				RecentEventCount:   9,
			}, nil
		},
	}

	service := NewClassInsightService(repo, repo, nil, nil)

	summary, err := service.GetClassSummary(context.Background(), 11, model.RoleTeacher, "Class A")
	if err != nil {
		t.Fatalf("GetClassSummary() error = %v", err)
	}
	if summary.ClassName != "Class A" {
		t.Fatalf("ClassName = %q, want Class A", summary.ClassName)
	}
	if summary.StudentCount != 3 || summary.ActiveStudentCount != 2 {
		t.Fatalf("summary = %+v, want mapped class summary fields", summary)
	}
}

func TestClassInsightQueryServiceGetClassReviewOnlyAttachesDimensionMatchedRecommendation(t *testing.T) {
	t.Parallel()

	nameAlice := "Alice"
	nameBob := "Bob"
	nameCarol := "Carol"

	repo := &classInsightRepoStub{
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		getClassSummaryFn: func(context.Context, string, time.Time) (*queryports.ClassSummary, error) {
			return &queryports.ClassSummary{
				ClassName:          "Class A",
				StudentCount:       3,
				AverageSolved:      2,
				ActiveStudentCount: 2,
				ActiveRate:         55,
				RecentEventCount:   12,
			}, nil
		},
		getClassTrendFn: func(context.Context, string, time.Time, int) (*queryports.ClassTrend, error) {
			return &queryports.ClassTrend{
				ClassName: "Class A",
				Points: []queryports.ClassTrendPoint{
					{Date: "2026-05-06", EventCount: 8, SolveCount: 4},
					{Date: "2026-05-12", EventCount: 5, SolveCount: 3},
				},
			}, nil
		},
		listClassTeachingFactSnapshots: func(context.Context, string, time.Time) ([]teachingadvice.StudentFactSnapshot, error) {
			return []teachingadvice.StudentFactSnapshot{
				{
					UserID:                 1,
					Username:               "alice",
					Name:                   &nameAlice,
					ActiveDays7d:           1,
					RecentEventCount7d:     1,
					CorrectSubmissionCount: 1,
					MaxWrongStreak:         4,
					Dimensions: []teachingadvice.DimensionFact{
						{Dimension: "web", ProfileScore: 0.2, AttemptCount: 4, SuccessCount: 0, EvidenceCount: 4},
					},
				},
				{
					UserID:                 2,
					Username:               "bob",
					Name:                   &nameBob,
					ActiveDays7d:           2,
					RecentEventCount7d:     3,
					CorrectSubmissionCount: 2,
					WriteupCount:           0,
					Dimensions: []teachingadvice.DimensionFact{
						{Dimension: "web", ProfileScore: 0.3, AttemptCount: 3, SuccessCount: 1, EvidenceCount: 3},
					},
				},
				{
					UserID:                 3,
					Username:               "carol",
					Name:                   &nameCarol,
					ActiveDays7d:           5,
					RecentEventCount7d:     8,
					CorrectSubmissionCount: 3,
					WriteupCount:           2,
					ApprovedReviewCount:    1,
					Dimensions: []teachingadvice.DimensionFact{
						{Dimension: "crypto", ProfileScore: 0.82, AttemptCount: 3, SuccessCount: 2, EvidenceCount: 3},
					},
				},
			}, nil
		},
	}
	recommendations := &classInsightRecommendationStub{
		recommendFn: func(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
			if userID != 1 || limit != 6 {
				return nil, nil
			}
			return &dto.RecommendationResp{
				Challenges: []*dto.ChallengeRecommendation{
					{
						ID:         99,
						Title:      "crypto-001",
						Category:   "crypto",
						Dimension:  "crypto",
						Difficulty: "easy",
						Summary:    "不应该挂到 web 聚焦建议上",
					},
					{
						ID:         101,
						Title:      "web-101",
						Category:   "web",
						Dimension:  "web",
						Difficulty: "easy",
						Summary:    "先补基础命中率",
					},
				},
			}, nil
		},
	}

	service := NewClassInsightService(repo, repo, recommendations, nil)

	review, err := service.GetClassReview(context.Background(), 11, model.RoleTeacher, "Class A")
	if err != nil {
		t.Fatalf("GetClassReview() error = %v", err)
	}
	if review.ClassName != "Class A" {
		t.Fatalf("ClassName = %q, want Class A", review.ClassName)
	}
	if len(review.Items) < 4 {
		t.Fatalf("Items = %+v, want multiple class review items", review.Items)
	}

	foundWeakDimensionRecommendation := false
	for _, item := range review.Items {
		switch item.Code {
		case "activity_risk":
			if len(item.Students) == 0 || item.Students[0].Username != "alice" {
				t.Fatalf("activity_risk students = %+v, want alice first", item.Students)
			}
			if item.Recommendation != nil {
				t.Fatalf("activity_risk recommendation = %+v, want no recommendation", item.Recommendation)
			}
		case "weak_dimension_cluster":
			if item.Recommendation == nil || item.Recommendation.Title != "web-101" {
				t.Fatalf("weak_dimension_cluster recommendation = %+v, want dimension-matched recommendation", item.Recommendation)
			}
			foundWeakDimensionRecommendation = true
		}
	}

	if !foundWeakDimensionRecommendation {
		t.Fatalf("review items = %+v, want weak_dimension_cluster recommendation", review.Items)
	}
}
