package commands

import (
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentreporting "ctf-platform/internal/module/assessment/application/reporting"
)

func RenderAWDReviewArchiveZip(targetPath string, archive *assessmentqry.TeacherAWDReviewArchiveResp) error {
	return assessmentreporting.RenderAWDReviewArchiveZip(targetPath, archive)
}

func RenderAWDReviewReportPDF(targetPath string, archive *assessmentqry.TeacherAWDReviewArchiveResp) error {
	return assessmentreporting.RenderAWDReviewReportPDF(targetPath, archive)
}
