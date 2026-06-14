package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/apperror"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentreporting "ctf-platform/internal/module/assessment/application/reporting"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	"ctf-platform/internal/shared/taxonomy"
	"ctf-platform/internal/teaching/classreview"
	"ctf-platform/internal/teaching/classwindow"
)

func (s *ReportService) generatePersonalReport(ctx context.Context, reportID, userID int64, format string) (string, time.Time, error) {
	data, err := s.buildPersonalReportData(ctx, userID)
	if err != nil {
		return "", time.Time{}, err
	}

	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypePersonal, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(output.LocalPath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateClassReport(ctx context.Context, reportID int64, className, format string, window classwindow.Range) (string, time.Time, error) {
	data, err := s.buildClassReportData(ctx, className, window)
	if err != nil {
		return "", time.Time{}, err
	}

	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypeClass, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(output.LocalPath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateContestExport(ctx context.Context, reportID, contestID int64, format string) (string, time.Time, error) {
	data, err := s.buildContestExportData(ctx, contestID)
	if err != nil {
		return "", time.Time{}, err
	}

	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypeContest, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(output.LocalPath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateStudentReviewArchive(ctx context.Context, reportID, studentID int64, format string) (string, time.Time, error) {
	data, err := s.buildStudentReviewArchiveData(ctx, studentID)
	if err != nil {
		return "", time.Time{}, err
	}

	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypeReview, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(output.LocalPath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewArchive(ctx context.Context, reportID int64, archive *assessmentqry.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypeAWDReviewArchive, assessmententity.ReportFormatZIP)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := assessmentreporting.RenderAWDReviewArchiveZip(output.LocalPath, archive); err != nil {
		return "", time.Time{}, err
	}
	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewReport(ctx context.Context, reportID int64, archive *assessmentqry.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	output, err := s.reportFilePath(ctx, reportID, assessmententity.ReportTypeAWDReviewReport, assessmententity.ReportFormatPDF)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := assessmentreporting.RenderAWDReviewReportPDF(output.LocalPath, archive); err != nil {
		return "", time.Time{}, err
	}
	return output.StorageKey, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) buildPersonalReportData(ctx context.Context, userID int64) (*personalReportData, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	skillProfileResp, err := s.assessmentService.GetSkillProfile(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	stats, err := s.personalRepo.GetPersonalStats(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	dimensionStats, err := s.personalRepo.ListPersonalDimensionStats(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &personalReportData{
		User:           user,
		SkillProfile:   skillProfileResp.Dimensions,
		Stats:          stats,
		DimensionStats: dimensionStats,
	}, nil
}

func (s *ReportService) buildClassReportData(ctx context.Context, className string, window classwindow.Range) (*classReportData, error) {
	totalStudents, err := s.classRepo.CountClassStudents(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	avgScore, err := s.classRepo.GetClassAverageScore(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	dimensionAverages, err := s.classRepo.ListClassDimensionAverages(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	topStudents, err := s.classRepo.ListClassTopStudents(ctx, className, 10)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	categoryDistribution, err := s.classRepo.ListClassCategoryDistribution(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	difficultyDistribution, err := s.classRepo.ListClassDifficultyDistribution(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	contestMigration, err := s.classRepo.GetClassContestMigrationSummary(ctx, className)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var summaryResp *classReportSummary
	var trendResp *classReportTrend
	var reviewResp *classReportReview
	if s.classInsightRepo != nil {
		summary, err := s.classInsightRepo.GetClassSummary(ctx, className, window.Since)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		summaryResp = mapClassSummary(summary)

		trend, err := s.classInsightRepo.GetClassTrend(ctx, className, window.StartOfDay, window.Days)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}
		trendResp = mapClassTrend(trend)

		snapshots, err := s.classInsightRepo.ListClassTeachingFactSnapshots(ctx, className, window.Since)
		if err != nil {
			return nil, apperror.ErrInternal.WithCause(err)
		}

		var trendEventDelta int64
		var trendSolveDelta int64
		hasTrend := trend != nil && len(trend.Points) >= 2
		if hasTrend {
			first := trend.Points[0]
			last := trend.Points[len(trend.Points)-1]
			trendEventDelta = last.EventCount - first.EventCount
			trendSolveDelta = last.SolveCount - first.SolveCount
		}
		reviewResp = mapClassReview(classreview.BuildResponse(ctx, classreview.Input{
			ClassName:        className,
			ActiveRate:       summary.ActiveRate,
			RecentEventCount: summary.RecentEventCount,
			HasTrend:         hasTrend,
			TrendEventDelta:  trendEventDelta,
			TrendSolveDelta:  trendSolveDelta,
			Snapshots:        snapshots,
		}, nil))
	}

	return &classReportData{
		ClassName:         className,
		Window:            classReportWindow{FromDate: window.FromDate, ToDate: window.ToDate, Days: window.Days},
		TotalStudents:     totalStudents,
		AverageScore:      avgScore,
		DimensionAverages: assessmentdomain.FillMissingDimensionAverages(dimensionAverages),
		TopStudents:       topStudents,
		Summary:           summaryResp,
		Trend:             trendResp,
		Review:            reviewResp,
		CategoryDistribution: assessmentdomain.FillMissingDistributionStats(
			categoryDistribution,
			taxonomy.AllDimensions,
		),
		DifficultyDistribution: assessmentdomain.FillMissingDistributionStats(
			difficultyDistribution,
			assessmentdomain.ClassReportDifficultyOrder(),
		),
		ContestMigration: derefClassContestMigration(contestMigration),
	}, nil
}

func (s *ReportService) buildContestExportData(ctx context.Context, contestID int64) (*contestExportData, error) {
	contest, err := s.contestRepo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentContestNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	scoreboard, err := s.contestExportRepo.ListContestScoreboard(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	challenges, err := s.contestExportRepo.ListContestChallenges(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	teams, err := s.contestExportRepo.ListContestTeams(ctx, contestID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &contestExportData{
		GeneratedAt: reportNow(),
		Contest: contestExportMeta{
			ID:          contest.ID,
			Title:       contest.Title,
			Description: contest.Description,
			Mode:        contest.Mode,
			Status:      contest.Status,
			StartTime:   contest.StartTime,
			EndTime:     contest.EndTime,
			FreezeTime:  contest.FreezeTime,
		},
		Scoreboard: scoreboard,
		Challenges: challenges,
		Teams:      teams,
	}, nil
}

func mapClassSummary(summary *assessmentdomain.ClassInsightSummary) *classReportSummary {
	if summary == nil {
		return nil
	}
	return &classReportSummary{
		ClassName:          summary.ClassName,
		StudentCount:       summary.StudentCount,
		AverageSolved:      summary.AverageSolved,
		ActiveStudentCount: summary.ActiveStudentCount,
		ActiveRate:         summary.ActiveRate,
		RecentEventCount:   summary.RecentEventCount,
	}
}

func mapClassTrend(trend *assessmentdomain.ClassInsightTrend) *classReportTrend {
	if trend == nil {
		return nil
	}
	points := make([]classReportTrendPoint, 0, len(trend.Points))
	for _, point := range trend.Points {
		points = append(points, classReportTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return &classReportTrend{
		ClassName: trend.ClassName,
		Points:    points,
	}
}

func mapClassReview(review *classreview.Response) *classReportReview {
	if review == nil {
		return nil
	}
	items := make([]classReportReviewItem, 0, len(review.Items))
	for _, item := range review.Items {
		items = append(items, classReportReviewItem{
			Code:           item.Code,
			Severity:       item.Severity,
			Summary:        item.Summary,
			Evidence:       item.Evidence,
			Action:         item.Action,
			ReasonCodes:    append([]string(nil), item.ReasonCodes...),
			Dimension:      item.Dimension,
			Students:       mapReviewStudents(item.Students),
			Recommendation: mapReviewRecommendation(item.Recommendation),
		})
	}
	return &classReportReview{
		ClassName: review.ClassName,
		Items:     items,
	}
}

func mapReviewStudents(students []classreview.ReviewStudentRef) []classReportReviewStudentRef {
	items := make([]classReportReviewStudentRef, 0, len(students))
	for _, student := range students {
		items = append(items, classReportReviewStudentRef{
			ID:       student.ID,
			Username: student.Username,
			Name:     student.Name,
		})
	}
	return items
}

func mapReviewRecommendation(item *classreview.RecommendationItem) *classReportRecommendationItem {
	if item == nil {
		return nil
	}
	return &classReportRecommendationItem{
		ChallengeID:    item.ChallengeID,
		Title:          item.Title,
		Category:       item.Category,
		Difficulty:     item.Difficulty,
		Dimension:      item.Dimension,
		DifficultyBand: item.DifficultyBand,
		Severity:       item.Severity,
		ReasonCodes:    append([]string(nil), item.ReasonCodes...),
		Summary:        item.Summary,
		Evidence:       item.Evidence,
	}
}

func derefClassContestMigration(summary *assessmentdomain.ClassContestMigrationSummary) assessmentdomain.ClassContestMigrationSummary {
	if summary == nil {
		return assessmentdomain.ClassContestMigrationSummary{}
	}
	return *summary
}
