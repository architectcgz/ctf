package queries

import (
	"ctf-platform/internal/dto"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/teaching/classreview"
)

func toTeacherClassSummary(source *queryports.ClassSummary) *TeacherClassSummary {
	if source == nil {
		return nil
	}
	return &TeacherClassSummary{
		ClassName:          source.ClassName,
		StudentCount:       source.StudentCount,
		AverageSolved:      source.AverageSolved,
		ActiveStudentCount: source.ActiveStudentCount,
		ActiveRate:         source.ActiveRate,
		RecentEventCount:   source.RecentEventCount,
	}
}

func toTeacherClassTrend(source *queryports.ClassTrend) *TeacherClassTrend {
	if source == nil {
		return nil
	}
	points := make([]TeacherClassTrendPoint, 0, len(source.Points))
	for _, point := range source.Points {
		points = append(points, TeacherClassTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return &TeacherClassTrend{
		ClassName: source.ClassName,
		Points:    points,
	}
}

func toTeacherClassReview(source *classreview.Response) *TeacherClassReview {
	if source == nil {
		return nil
	}
	items := make([]TeacherClassReviewItem, 0, len(source.Items))
	for _, item := range source.Items {
		items = append(items, TeacherClassReviewItem{
			Code:           item.Code,
			Severity:       item.Severity,
			Summary:        item.Summary,
			Evidence:       item.Evidence,
			Action:         item.Action,
			ReasonCodes:    append([]string(nil), item.ReasonCodes...),
			Dimension:      item.Dimension,
			Students:       toTeacherReviewStudentRefs(item.Students),
			Recommendation: toTeacherRecommendationItem(item.Recommendation),
		})
	}
	return &TeacherClassReview{
		ClassName: source.ClassName,
		Items:     items,
	}
}

func toTeacherReviewStudentRefs(source []classreview.ReviewStudentRef) []TeacherReviewStudentRef {
	refs := make([]TeacherReviewStudentRef, 0, len(source))
	for _, item := range source {
		refs = append(refs, TeacherReviewStudentRef{
			ID:       item.ID,
			Username: item.Username,
			Name:     item.Name,
		})
	}
	return refs
}

func toTeacherRecommendationItem(source *classreview.RecommendationItem) *dto.TeacherRecommendationItem {
	if source == nil {
		return nil
	}
	return &dto.TeacherRecommendationItem{
		ChallengeID:    source.ChallengeID,
		Title:          source.Title,
		Category:       source.Category,
		Difficulty:     source.Difficulty,
		Dimension:      source.Dimension,
		DifficultyBand: source.DifficultyBand,
		Severity:       source.Severity,
		ReasonCodes:    append([]string(nil), source.ReasonCodes...),
		Summary:        source.Summary,
		Evidence:       source.Evidence,
	}
}
