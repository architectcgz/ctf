package reporting

import "strings"

type summaryLine struct {
	Label string
	Value string
}

func safeTrendPoints(trend *ClassReportTrend) []ClassReportTrendPoint {
	if trend == nil {
		return []ClassReportTrendPoint{}
	}
	return trend.Points
}

func safeReviewItems(review *ClassReportReview) []ClassReportReviewItem {
	if review == nil {
		return []ClassReportReviewItem{}
	}
	return review.Items
}

func reviewStudentNames(students []ClassReportReviewStudentRef) string {
	names := make([]string, 0, len(students))
	for _, student := range students {
		if student.Name != nil && strings.TrimSpace(*student.Name) != "" {
			names = append(names, strings.TrimSpace(*student.Name))
			continue
		}
		if strings.TrimSpace(student.Username) != "" {
			names = append(names, strings.TrimSpace(student.Username))
		}
	}
	return strings.Join(names, ", ")
}

func reviewSummaryActiveRate(summary *ClassReportSummary) float64 {
	if summary == nil {
		return 0
	}
	return summary.ActiveRate
}

func reviewSummaryRecentEvents(summary *ClassReportSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.RecentEventCount
}
