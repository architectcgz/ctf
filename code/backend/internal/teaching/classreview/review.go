package classreview

import (
	"context"

	"ctf-platform/internal/dto"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

const recommendationLimit = 6

type Input struct {
	ClassName        string
	ActiveRate       float64
	RecentEventCount int64
	HasTrend         bool
	TrendEventDelta  int64
	TrendSolveDelta  int64
	Snapshots        []teachingadvice.StudentFactSnapshot
}

type RecommendationResolver interface {
	Resolve(ctx context.Context, candidateIDs []int64, dimension string, limit int) *dto.TeacherRecommendationItem
}

type RecommendationResolverFunc func(ctx context.Context, candidateIDs []int64, dimension string, limit int) *dto.TeacherRecommendationItem

func (f RecommendationResolverFunc) Resolve(ctx context.Context, candidateIDs []int64, dimension string, limit int) *dto.TeacherRecommendationItem {
	if f == nil {
		return nil
	}
	return f(ctx, candidateIDs, dimension, limit)
}

func BuildResponse(ctx context.Context, input Input, resolver RecommendationResolver) *dto.TeacherClassReviewResp {
	evaluations := make(map[int64]teachingadvice.StudentEvaluation, len(input.Snapshots))
	studentRefs := make(map[int64]dto.TeacherReviewStudentRef, len(input.Snapshots))
	for _, snapshot := range input.Snapshots {
		evaluations[snapshot.UserID] = teachingadvice.EvaluateStudent(snapshot)
		studentRefs[snapshot.UserID] = dto.TeacherReviewStudentRef{
			ID:       snapshot.UserID,
			Username: snapshot.Username,
			Name:     snapshot.Name,
		}
	}

	var trend *teachingadvice.ClassTrendSnapshot
	if input.HasTrend {
		trend = &teachingadvice.ClassTrendSnapshot{
			EventDelta: input.TrendEventDelta,
			SolveDelta: input.TrendSolveDelta,
		}
	}

	adviceItems := teachingadvice.BuildClassReview(
		input.ClassName,
		teachingadvice.ClassSummarySnapshot{
			ClassName:        input.ClassName,
			StudentCount:     len(input.Snapshots),
			ActiveRate:       input.ActiveRate,
			RecentEventCount: input.RecentEventCount,
		},
		trend,
		input.Snapshots,
		evaluations,
	)

	items := make([]dto.TeacherClassReviewItem, 0, len(adviceItems))
	for _, adviceItem := range adviceItems {
		item := dto.TeacherClassReviewItem{
			Code:        adviceItem.Code,
			Severity:    string(adviceItem.Severity),
			Summary:     adviceItem.Summary,
			Evidence:    adviceItem.Evidence,
			Action:      adviceItem.Action,
			ReasonCodes: append([]string(nil), adviceItem.ReasonCodes...),
			Dimension:   adviceItem.Dimension,
			Students:    reviewStudentRefsByIDs(studentRefs, adviceItem.StudentIDs),
		}
		if resolver != nil && adviceItem.RecommendationStudentID != nil {
			if recommendation := resolver.Resolve(
				ctx,
				prioritizedStudentIDs(*adviceItem.RecommendationStudentID, adviceItem.StudentIDs),
				adviceItem.Dimension,
				recommendationLimit,
			); recommendation != nil {
				item.Recommendation = recommendation
			}
		}
		items = append(items, item)
	}

	return &dto.TeacherClassReviewResp{
		ClassName: input.ClassName,
		Items:     items,
	}
}

func prioritizedStudentIDs(primary int64, studentIDs []int64) []int64 {
	ids := make([]int64, 0, len(studentIDs)+1)
	seen := make(map[int64]struct{}, len(studentIDs)+1)
	appendID := func(id int64) {
		if id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	appendID(primary)
	for _, studentID := range studentIDs {
		appendID(studentID)
	}
	return ids
}

func reviewStudentRefsByIDs(
	refsByID map[int64]dto.TeacherReviewStudentRef,
	studentIDs []int64,
) []dto.TeacherReviewStudentRef {
	refs := make([]dto.TeacherReviewStudentRef, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		ref, ok := refsByID[studentID]
		if !ok {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}
