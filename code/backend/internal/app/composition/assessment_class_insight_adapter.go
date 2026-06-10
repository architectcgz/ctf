package composition

import (
	"context"
	"time"

	assessmentports "ctf-platform/internal/module/assessment/ports"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type assessmentClassInsightAdapter struct {
	repo queryports.TeachingClassInsightRepository
}

var _ assessmentports.AssessmentClassInsightRepository = assessmentClassInsightAdapter{}

func newAssessmentClassInsightAdapter(repo queryports.TeachingClassInsightRepository) assessmentports.AssessmentClassInsightRepository {
	if repo == nil {
		return nil
	}
	return assessmentClassInsightAdapter{repo: repo}
}

func (a assessmentClassInsightAdapter) GetClassSummary(ctx context.Context, className string, since time.Time) (*assessmentports.ClassInsightSummary, error) {
	summary, err := a.repo.GetClassSummary(ctx, className, since)
	if err != nil || summary == nil {
		return nil, err
	}
	return &assessmentports.ClassInsightSummary{
		ClassName:          summary.ClassName,
		StudentCount:       summary.StudentCount,
		AverageSolved:      summary.AverageSolved,
		ActiveStudentCount: summary.ActiveStudentCount,
		ActiveRate:         summary.ActiveRate,
		RecentEventCount:   summary.RecentEventCount,
	}, nil
}

func (a assessmentClassInsightAdapter) GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*assessmentports.ClassInsightTrend, error) {
	trend, err := a.repo.GetClassTrend(ctx, className, since, days)
	if err != nil || trend == nil {
		return nil, err
	}
	points := make([]assessmentports.ClassInsightTrendPoint, 0, len(trend.Points))
	for _, point := range trend.Points {
		points = append(points, assessmentports.ClassInsightTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return &assessmentports.ClassInsightTrend{
		ClassName: trend.ClassName,
		Points:    points,
	}, nil
}

func (a assessmentClassInsightAdapter) ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error) {
	return a.repo.ListClassTeachingFactSnapshots(ctx, className, since)
}
