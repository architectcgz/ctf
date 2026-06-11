package commands

import (
	"time"

	assessmentreporting "ctf-platform/internal/module/assessment/application/reporting"
)

type personalReportData = assessmentreporting.PersonalReportData
type classReportData = assessmentreporting.ClassReportData
type classReportWindow = assessmentreporting.ClassReportWindow
type classReportSummary = assessmentreporting.ClassReportSummary
type classReportTrendPoint = assessmentreporting.ClassReportTrendPoint
type classReportTrend = assessmentreporting.ClassReportTrend
type classReportReviewStudentRef = assessmentreporting.ClassReportReviewStudentRef
type classReportRecommendationItem = assessmentreporting.ClassReportRecommendationItem
type classReportReviewItem = assessmentreporting.ClassReportReviewItem
type classReportReview = assessmentreporting.ClassReportReview
type contestExportData = assessmentreporting.ContestExportData
type contestExportMeta = assessmentreporting.ContestExportMeta
type ReviewArchiveData = assessmentreporting.ReviewArchiveData
type ReviewArchiveStudent = assessmentreporting.ReviewArchiveStudent

var reportNow = func() time.Time {
	return time.Now().UTC()
}
