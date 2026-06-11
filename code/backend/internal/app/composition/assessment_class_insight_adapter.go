package composition

import (
	"context"
	"time"

	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	queryports "ctf-platform/internal/module/teaching_analysis/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type teachingClassInsightSource interface {
	GetClassSummary(ctx context.Context, className string, since time.Time) (*queryports.ClassSummary, error)
	GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*queryports.ClassTrend, error)
	ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error)
}

type assessmentClassInsightAdapter struct {
	source teachingClassInsightSource
}

var _ assessmentports.AssessmentClassInsightRepository = (*assessmentClassInsightAdapter)(nil)

func newAssessmentClassInsightAdapter(source teachingClassInsightSource) assessmentports.AssessmentClassInsightRepository {
	if source == nil {
		return nil
	}
	return &assessmentClassInsightAdapter{source: source}
}

func (a *assessmentClassInsightAdapter) GetClassSummary(ctx context.Context, className string, since time.Time) (*assessmentdomain.ClassInsightSummary, error) {
	summary, err := a.source.GetClassSummary(ctx, className, since)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, nil
	}
	return &assessmentdomain.ClassInsightSummary{
		ClassName:          summary.ClassName,
		StudentCount:       summary.StudentCount,
		AverageSolved:      summary.AverageSolved,
		ActiveStudentCount: summary.ActiveStudentCount,
		ActiveRate:         summary.ActiveRate,
		RecentEventCount:   summary.RecentEventCount,
	}, nil
}

func (a *assessmentClassInsightAdapter) GetClassTrend(ctx context.Context, className string, since time.Time, days int) (*assessmentdomain.ClassInsightTrend, error) {
	trend, err := a.source.GetClassTrend(ctx, className, since, days)
	if err != nil {
		return nil, err
	}
	if trend == nil {
		return nil, nil
	}
	points := make([]assessmentdomain.ClassInsightTrendPoint, 0, len(trend.Points))
	for _, point := range trend.Points {
		points = append(points, assessmentdomain.ClassInsightTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return &assessmentdomain.ClassInsightTrend{
		ClassName: trend.ClassName,
		Points:    points,
	}, nil
}

func (a *assessmentClassInsightAdapter) ListClassTeachingFactSnapshots(ctx context.Context, className string, since time.Time) ([]teachingadvice.StudentFactSnapshot, error) {
	return a.source.ListClassTeachingFactSnapshots(ctx, className, since)
}
