package queries

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/evidence"
)

type overviewRepoStub struct {
	findUserByIDFn          func(ctx context.Context, userID int64) (*model.User, error)
	countStudentsByClassFn  func(ctx context.Context, className string) (int64, error)
	listClassesFn           func(ctx context.Context, offset, limit int) ([]readmodelports.ClassItem, error)
	listStudentsByClassFn   func(ctx context.Context, className, keyword, studentNo string, since time.Time) ([]readmodelports.StudentItem, error)
	listStudentsByClassesFn func(ctx context.Context, classNames []string, keyword, studentNo string, since time.Time) ([]readmodelports.StudentItem, error)
	getClassSummaryFn       func(ctx context.Context, className string, since time.Time) (*readmodelports.ClassSummary, error)
	getOverviewTrendFn      func(ctx context.Context, classNames []string, since time.Time, days int) (*readmodelports.OverviewTrend, error)
}

func (s *overviewRepoStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if s.findUserByIDFn != nil {
		return s.findUserByIDFn(ctx, userID)
	}
	return nil, nil
}

func (s *overviewRepoStub) CountStudentsByClass(ctx context.Context, className string) (int64, error) {
	if s.countStudentsByClassFn != nil {
		return s.countStudentsByClassFn(ctx, className)
	}
	return 0, nil
}

func (s *overviewRepoStub) CountClasses(context.Context) (int64, error) {
	return 0, nil
}

func (s *overviewRepoStub) ListClasses(ctx context.Context, offset, limit int) ([]readmodelports.ClassItem, error) {
	if s.listClassesFn != nil {
		return s.listClassesFn(ctx, offset, limit)
	}
	return []readmodelports.ClassItem{}, nil
}

func (s *overviewRepoStub) ListStudents(
	context.Context,
	string,
	string,
	string,
	string,
	string,
	time.Time,
	int,
	int,
) ([]readmodelports.StudentItem, int64, error) {
	return []readmodelports.StudentItem{}, 0, nil
}

func (s *overviewRepoStub) ListStudentsByClass(
	ctx context.Context,
	className, keyword, studentNo string,
	since time.Time,
) ([]readmodelports.StudentItem, error) {
	if s.listStudentsByClassFn != nil {
		return s.listStudentsByClassFn(ctx, className, keyword, studentNo, since)
	}
	return []readmodelports.StudentItem{}, nil
}

func (s *overviewRepoStub) CountPublishedChallenges(context.Context) (int64, error) {
	return 0, nil
}

func (s *overviewRepoStub) CountSolvedChallenges(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *overviewRepoStub) GetCategoryProgress(context.Context, int64) ([]readmodelports.ProgressRow, error) {
	return []readmodelports.ProgressRow{}, nil
}

func (s *overviewRepoStub) GetDifficultyProgress(context.Context, int64) ([]readmodelports.ProgressRow, error) {
	return []readmodelports.ProgressRow{}, nil
}

func (s *overviewRepoStub) GetStudentTimeline(context.Context, int64, int, int) ([]readmodelports.TimelineEventRecord, error) {
	return []readmodelports.TimelineEventRecord{}, nil
}

func (s *overviewRepoStub) GetStudentEvidence(context.Context, int64, evidence.Query) ([]readmodelports.EvidenceEventRecord, error) {
	return []readmodelports.EvidenceEventRecord{}, nil
}

func (s *overviewRepoStub) GetClassSummary(
	ctx context.Context,
	className string,
	since time.Time,
) (*readmodelports.ClassSummary, error) {
	if s.getClassSummaryFn != nil {
		return s.getClassSummaryFn(ctx, className, since)
	}
	return nil, nil
}

func (s *overviewRepoStub) GetClassTrend(context.Context, string, time.Time, int) (*readmodelports.ClassTrend, error) {
	return &readmodelports.ClassTrend{}, nil
}

func (s *overviewRepoStub) ListClassTeachingFactSnapshots(
	context.Context,
	string,
	time.Time,
) ([]teachingadvice.StudentFactSnapshot, error) {
	return []teachingadvice.StudentFactSnapshot{}, nil
}

func (s *overviewRepoStub) ListStudentsByClasses(
	ctx context.Context,
	classNames []string,
	keyword, studentNo string,
	since time.Time,
) ([]readmodelports.StudentItem, error) {
	if s.listStudentsByClassesFn != nil {
		return s.listStudentsByClassesFn(ctx, classNames, keyword, studentNo, since)
	}
	return []readmodelports.StudentItem{}, nil
}

func (s *overviewRepoStub) GetOverviewTrend(
	ctx context.Context,
	classNames []string,
	since time.Time,
	days int,
) (*readmodelports.OverviewTrend, error) {
	if s.getOverviewTrendFn != nil {
		return s.getOverviewTrendFn(ctx, classNames, since, days)
	}
	return &readmodelports.OverviewTrend{}, nil
}

func TestOverviewQueryServiceGetOverviewBuildsScopeSummary(t *testing.T) {
	t.Parallel()

	repo := &overviewRepoStub{
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: 11, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		countStudentsByClassFn: func(context.Context, string) (int64, error) {
			return 2, nil
		},
		listStudentsByClassesFn: func(context.Context, []string, string, string, time.Time) ([]readmodelports.StudentItem, error) {
			weakCrypto := "crypto"
			weakPwn := "pwn"
			className := "Class A"
			nameAlice := "Alice"
			nameBob := "Bob"
			return []readmodelports.StudentItem{
				{
					ID:               1,
					Username:         "alice",
					Name:             &nameAlice,
					ClassName:        &className,
					SolvedCount:      4,
					TotalScore:       320,
					RecentEventCount: 0,
					WeakDimension:    &weakCrypto,
				},
				{
					ID:               2,
					Username:         "bob",
					Name:             &nameBob,
					ClassName:        &className,
					SolvedCount:      2,
					TotalScore:       180,
					RecentEventCount: 3,
					WeakDimension:    &weakPwn,
				},
			}, nil
		},
		listStudentsByClassFn: func(context.Context, string, string, string, time.Time) ([]readmodelports.StudentItem, error) {
			weakCrypto := "crypto"
			weakPwn := "pwn"
			className := "Class A"
			return []readmodelports.StudentItem{
				{ID: 1, Username: "alice", ClassName: &className, SolvedCount: 4, TotalScore: 320, RecentEventCount: 0, WeakDimension: &weakCrypto},
				{ID: 2, Username: "bob", ClassName: &className, SolvedCount: 2, TotalScore: 180, RecentEventCount: 3, WeakDimension: &weakPwn},
			}, nil
		},
		getClassSummaryFn: func(context.Context, string, time.Time) (*readmodelports.ClassSummary, error) {
			return &readmodelports.ClassSummary{
				ClassName:          "Class A",
				StudentCount:       2,
				AverageSolved:      3,
				ActiveStudentCount: 1,
				ActiveRate:         50,
				RecentEventCount:   3,
			}, nil
		},
		getOverviewTrendFn: func(context.Context, []string, time.Time, int) (*readmodelports.OverviewTrend, error) {
			return &readmodelports.OverviewTrend{
				Points: []readmodelports.OverviewTrendPoint{
					{Date: "2026-05-06", ActiveStudentCount: 1, EventCount: 2, SolveCount: 1},
					{Date: "2026-05-07", ActiveStudentCount: 1, EventCount: 1, SolveCount: 1},
				},
			}, nil
		},
	}

	service := NewOverviewService(repo)

	overview, err := service.GetOverview(context.Background(), 11, model.RoleTeacher)
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}

	if overview.Summary.ClassCount != 1 {
		t.Fatalf("ClassCount = %d, want 1", overview.Summary.ClassCount)
	}
	if overview.Summary.StudentCount != 2 {
		t.Fatalf("StudentCount = %d, want 2", overview.Summary.StudentCount)
	}
	if overview.Summary.ActiveStudentCount != 1 {
		t.Fatalf("ActiveStudentCount = %d, want 1", overview.Summary.ActiveStudentCount)
	}
	if overview.Summary.RiskStudentCount != 1 {
		t.Fatalf("RiskStudentCount = %d, want 1", overview.Summary.RiskStudentCount)
	}
	if len(overview.FocusClasses) != 1 || overview.FocusClasses[0].ClassName != "Class A" {
		t.Fatalf("FocusClasses = %+v, want Class A", overview.FocusClasses)
	}
	if len(overview.FocusStudents) != 1 || overview.FocusStudents[0].Username != "alice" {
		t.Fatalf("FocusStudents = %+v, want alice as risk student", overview.FocusStudents)
	}
	if overview.SpotlightStudent == nil || overview.SpotlightStudent.Username != "alice" {
		t.Fatalf("SpotlightStudent = %+v, want alice", overview.SpotlightStudent)
	}
	if len(overview.WeakDimensions) != 2 || overview.WeakDimensions[0].Dimension != "crypto" {
		t.Fatalf("WeakDimensions = %+v, want crypto first", overview.WeakDimensions)
	}
	if len(overview.Trend.Points) != 2 || overview.Trend.Points[0].Date != "2026-05-06" {
		t.Fatalf("Trend.Points = %+v, want mapped overview trend points", overview.Trend.Points)
	}
}

func TestOverviewQueryServiceGetOverviewWithoutAccessibleClassReturnsEmptyScope(t *testing.T) {
	t.Parallel()

	repo := &overviewRepoStub{
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: 22, Role: model.RoleTeacher, ClassName: ""}, nil
		},
	}

	service := NewOverviewService(repo)

	overview, err := service.GetOverview(context.Background(), 22, model.RoleTeacher)
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}

	if overview.Summary.StudentCount != 0 {
		t.Fatalf("StudentCount = %d, want 0", overview.Summary.StudentCount)
	}
	if len(overview.FocusClasses) != 0 || len(overview.FocusStudents) != 0 || len(overview.WeakDimensions) != 0 {
		t.Fatalf("overview should be empty, got %+v", overview)
	}
}
