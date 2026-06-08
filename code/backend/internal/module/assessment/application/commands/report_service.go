package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentconfig "ctf-platform/internal/module/assessment/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingquerycontracts "ctf-platform/internal/module/teaching_query/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/internal/teaching/classreview"
	"ctf-platform/internal/teaching/classwindow"
	teachingevidence "ctf-platform/internal/teaching/evidence"
)

type ReportService struct {
	lifecycleRepo     assessmentports.AssessmentReportLifecycleRepository
	userRepo          assessmentports.AssessmentReportUserLookupRepository
	contestRepo       assessmentports.AssessmentReportContestLookupRepository
	personalRepo      assessmentports.AssessmentPersonalReportRepository
	classRepo         assessmentports.AssessmentClassReportRepository
	classInsightRepo  assessmentports.AssessmentClassInsightRepository
	contestExportRepo assessmentports.AssessmentContestExportRepository
	reviewArchiveRepo assessmentports.AssessmentReviewArchiveRepository
	assessmentService assessmentports.AssessmentProfileReader
	awdReviewBuilder  AWDReviewExportBuilder
	config            config.ReportConfig
	logger            *zap.Logger
	workerPool        chan struct{}
	baseCtx           context.Context
	cancel            context.CancelFunc
	tasks             sync.WaitGroup
}

type personalReportData struct {
	User           *assessmentdomain.ReportUser
	SkillProfile   []*assessmentcontracts.SkillDimension
	Stats          *assessmentdomain.PersonalReportStats
	DimensionStats []assessmentdomain.ReportDimensionStat
}

type classReportData struct {
	ClassName              string                                        `json:"class_name"`
	Window                 classReportWindow                             `json:"window"`
	TotalStudents          int                                           `json:"total_students"`
	AverageScore           float64                                       `json:"average_score"`
	DimensionAverages      []assessmentdomain.ClassDimensionAverage      `json:"dimension_averages"`
	TopStudents            []assessmentdomain.ClassTopStudent            `json:"top_students"`
	Summary                *teachingquerycontracts.TeacherClassSummary   `json:"summary,omitempty"`
	Trend                  *teachingquerycontracts.TeacherClassTrend     `json:"trend,omitempty"`
	Review                 *teachingquerycontracts.TeacherClassReview    `json:"review,omitempty"`
	CategoryDistribution   []assessmentdomain.ClassDistributionStat      `json:"category_distribution"`
	DifficultyDistribution []assessmentdomain.ClassDistributionStat      `json:"difficulty_distribution"`
	ContestMigration       assessmentdomain.ClassContestMigrationSummary `json:"contest_migration"`
}

type classReportWindow struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	Days     int    `json:"days"`
}

type contestExportData struct {
	GeneratedAt time.Time                                      `json:"generated_at"`
	Contest     contestExportMeta                              `json:"contest"`
	Scoreboard  []assessmentdomain.ContestExportScoreboardItem `json:"scoreboard"`
	Challenges  []assessmentdomain.ContestExportChallengeItem  `json:"challenges"`
	Teams       []assessmentdomain.ContestExportTeamItem       `json:"teams"`
}

type contestExportMeta struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Mode        string     `json:"mode"`
	Status      string     `json:"status"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	FreezeTime  *time.Time `json:"freeze_time,omitempty"`
}

type ReviewArchiveData struct {
	GeneratedAt         time.Time                                         `json:"generated_at"`
	Student             ReviewArchiveStudent                              `json:"student"`
	Summary             assessmentdomain.ReviewArchiveSummary             `json:"summary"`
	SkillProfile        []*assessmentcontracts.SkillDimension             `json:"skill_profile,omitempty"`
	Timeline            []assessmentdomain.ReviewArchiveTimelineEvent     `json:"timeline"`
	Evidence            []assessmentdomain.ReviewArchiveEvidenceEvent     `json:"evidence"`
	Writeups            []assessmentdomain.ReviewArchiveWriteupItem       `json:"writeups"`
	ManualReviews       []assessmentdomain.ReviewArchiveManualReviewItem  `json:"manual_reviews"`
	TeacherObservations assessmentdomain.ReviewArchiveTeacherObservations `json:"teacher_observations"`
}

type ReviewArchiveStudent struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	ClassName string `json:"class_name,omitempty"`
}

var reportNow = func() time.Time {
	return time.Now().UTC()
}

func NewReportService(
	lifecycleRepo assessmentports.AssessmentReportLifecycleRepository,
	userRepo assessmentports.AssessmentReportUserLookupRepository,
	contestRepo assessmentports.AssessmentReportContestLookupRepository,
	personalRepo assessmentports.AssessmentPersonalReportRepository,
	classRepo assessmentports.AssessmentClassReportRepository,
	classInsightRepo assessmentports.AssessmentClassInsightRepository,
	contestExportRepo assessmentports.AssessmentContestExportRepository,
	reviewArchiveRepo assessmentports.AssessmentReviewArchiveRepository,
	assessmentService assessmentports.AssessmentProfileReader,
	cfg config.ReportConfig,
	logger *zap.Logger,
) *ReportService {
	if logger == nil {
		logger = zap.NewNop()
	}

	cfg = assessmentconfig.NormalizeReportConfig(cfg)
	return &ReportService{
		lifecycleRepo:     lifecycleRepo,
		userRepo:          userRepo,
		contestRepo:       contestRepo,
		personalRepo:      personalRepo,
		classRepo:         classRepo,
		classInsightRepo:  classInsightRepo,
		contestExportRepo: contestExportRepo,
		reviewArchiveRepo: reviewArchiveRepo,
		assessmentService: assessmentService,
		config:            cfg,
		logger:            logger,
		workerPool:        make(chan struct{}, cfg.MaxWorkers),
	}
}

func (s *ReportService) StartBackgroundTasks(ctx context.Context) {
	if s == nil || ctx == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.baseCtx, s.cancel = context.WithCancel(ctx)
}

func (s *ReportService) SetAWDReviewExportBuilder(builder AWDReviewExportBuilder) {
	if s == nil {
		return
	}
	s.awdReviewBuilder = builder
}

func (s *ReportService) CreatePersonalReport(ctx context.Context, userID int64, req CreatePersonalReportInput) (*ReportExportData, error) {
	if ctx == nil {
		return nil, errors.New("create personal report requires context")
	}

	format := s.normalizeFormat(req.Format)
	report := &assessmententity.Report{
		Type:   assessmententity.ReportTypePersonal,
		Format: format,
		UserID: &userID,
		Status: assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	reportCtx, cancel := s.withPersonalTimeout(ctx)
	defer cancel()

	filePath, expiresAt, err := s.generatePersonalReport(reportCtx, report.ID, userID, format)
	if err != nil {
		s.markFailed(reportCtx, report.ID, err)
		return nil, err
	}
	if err := s.lifecycleRepo.MarkReady(reportCtx, report.ID, filePath, expiresAt); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return buildReportExportData(report.ID, assessmententity.ReportStatusReady, expiresAt), nil
}

func (s *ReportService) withPersonalTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.config.PersonalTimeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, s.config.PersonalTimeout)
}

func (s *ReportService) CreateClassReport(ctx context.Context, requesterID int64, req CreateClassReportInput) (*ReportExportData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	className := strings.TrimSpace(req.ClassName)
	if className == "" {
		className = strings.TrimSpace(requester.ClassName)
	}
	if className == "" {
		return nil, apperror.ErrInvalidParams.WithMessage("class_name 不能为空")
	}
	if err := validateClassReportAccess(requester, className); err != nil {
		return nil, err
	}
	window, err := s.parseClassWindow(req)
	if err != nil {
		return nil, err
	}

	format := s.normalizeFormat(req.Format)
	report := &assessmententity.Report{
		Type:      assessmententity.ReportTypeClass,
		Format:    format,
		UserID:    &requesterID,
		ClassName: &className,
		Status:    assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateClassReport(runCtx, report.ID, className, format, window)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, assessmententity.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateContestExport(ctx context.Context, requesterID, contestID int64, req CreateContestExportInput) (*ReportExportData, error) {
	if _, err := s.contestRepo.FindContestByID(ctx, contestID); err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	format := s.normalizeArchiveFormat(req.Format)
	report := &assessmententity.Report{
		Type:   assessmententity.ReportTypeContest,
		Format: format,
		UserID: &requesterID,
		Status: assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateContestExport(runCtx, report.ID, contestID, format)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, assessmententity.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateStudentReviewArchive(ctx context.Context, requesterID, studentID int64, req CreateStudentReviewArchiveInput) (*ReportExportData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}
	if err := validateStudentReviewArchiveAccess(requester, student); err != nil {
		return nil, err
	}

	format := s.normalizeArchiveFormat(req.Format)
	report := &assessmententity.Report{
		Type:      assessmententity.ReportTypeReview,
		Format:    format,
		UserID:    &requesterID,
		ClassName: &student.ClassName,
		Status:    assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateStudentReviewArchive(runCtx, report.ID, studentID, format)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, assessmententity.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateTeacherAWDReviewArchive(ctx context.Context, requesterID, contestID int64, req CreateTeacherAWDReviewExportInput) (*ReportExportData, error) {
	if _, err := s.findAWDContestForExport(ctx, contestID); err != nil {
		return nil, err
	}
	if s.awdReviewBuilder == nil {
		return nil, apperror.ErrServiceUnavailable.WithMessage("教师 AWD 复盘归档导出暂不可用")
	}

	report := &assessmententity.Report{
		Type:   assessmententity.ReportTypeAWDReviewArchive,
		Format: assessmententity.ReportFormatZIP,
		UserID: &requesterID,
		Status: assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	roundNumber := req.RoundNumber
	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		archive, err := s.awdReviewBuilder.BuildArchive(runCtx, requesterID, contestID, roundNumber)
		if err != nil {
			return err
		}
		filePath, expiresAt, err := s.generateTeacherAWDReviewArchive(report.ID, archive)
		if err != nil {
			return err
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, assessmententity.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateTeacherAWDReviewReport(ctx context.Context, requesterID, contestID int64, req CreateTeacherAWDReviewExportInput) (*ReportExportData, error) {
	contest, err := s.findAWDContestForExport(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if contest.Status != contestcontracts.ContestStatusEnded {
		return nil, apperror.ErrInvalidParams.WithMessage("教师复盘报告仅支持赛后导出")
	}
	if s.awdReviewBuilder == nil {
		return nil, apperror.ErrServiceUnavailable.WithMessage("教师 AWD 复盘报告导出暂不可用")
	}

	report := &assessmententity.Report{
		Type:   assessmententity.ReportTypeAWDReviewReport,
		Format: assessmententity.ReportFormatPDF,
		UserID: &requesterID,
		Status: assessmententity.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	roundNumber := req.RoundNumber
	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		archive, err := s.awdReviewBuilder.BuildArchive(runCtx, requesterID, contestID, roundNumber)
		if err != nil {
			return err
		}
		filePath, expiresAt, err := s.generateTeacherAWDReviewReport(report.ID, archive)
		if err != nil {
			return err
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, assessmententity.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) GetStudentReviewArchive(ctx context.Context, requesterID, studentID int64) (*ReviewArchiveData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}
	if err := validateStudentReviewArchiveAccess(requester, student); err != nil {
		return nil, err
	}
	return s.buildStudentReviewArchiveData(ctx, studentID)
}

func validateClassReportAccess(requester *assessmentdomain.ReportUser, className string) error {
	if requester == nil || requester.ID <= 0 {
		return apperror.ErrUnauthorized
	}
	if requester.Role == identitycontracts.RoleAdmin {
		return nil
	}
	if strings.TrimSpace(requester.ClassName) == "" || strings.TrimSpace(requester.ClassName) != className {
		return apperror.ErrForbidden
	}
	return nil
}

func validateStudentReviewArchiveAccess(requester, student *assessmentdomain.ReportUser) error {
	if requester == nil || requester.ID <= 0 {
		return apperror.ErrUnauthorized
	}
	if student == nil || student.ID <= 0 {
		return apperror.ErrNotFound
	}
	if student.Role != identitycontracts.RoleStudent {
		return apperror.ErrInvalidParams.WithMessage("目标用户不是学生")
	}
	if requester.Role == identitycontracts.RoleAdmin {
		return nil
	}
	if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != student.ClassName {
		return apperror.ErrForbidden
	}
	return nil
}

func (s *ReportService) findAWDContestForExport(ctx context.Context, contestID int64) (*contestcontracts.Contest, error) {
	contest, err := s.contestRepo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if contest.Mode != contestcontracts.ContestModeAWD {
		return nil, contestcontracts.ErrContestNotFound
	}
	return contest, nil
}

func (s *ReportService) GetDownload(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error) {
	report, err := s.lifecycleRepo.FindByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentReportNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := assessmentdomain.ValidateReportAccess(report, requesterID, role); err != nil {
		return nil, err
	}
	if report.Status == assessmententity.ReportStatusProcessing {
		return nil, apperror.ErrConflict.WithMessage("报告仍在生成中")
	}
	if report.Status == assessmententity.ReportStatusFailed {
		message := "报告生成失败"
		if report.ErrorMsg != nil && strings.TrimSpace(*report.ErrorMsg) != "" {
			message = *report.ErrorMsg
		}
		return nil, apperror.ErrConflict.WithMessage(message)
	}

	filePath, err := s.safeReportPath(report.FilePath)
	if err != nil {
		return nil, apperror.ErrForbidden
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(statErr)
	}

	fileName := reportDownloadFileName(report)
	format := reportOutputFormat(report)
	contentType := reportContentType(format)
	if contentType == "" {
		switch format {
		case assessmententity.ReportFormatJSON:
			contentType = "application/json"
		case assessmententity.ReportFormatPDF:
			contentType = "application/pdf"
		case assessmententity.ReportFormatExcel:
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case assessmententity.ReportFormatZIP:
			contentType = "application/zip"
		default:
			contentType = "application/octet-stream"
		}
	}

	return &assessmentdomain.ReportDownload{
		Path:        filePath,
		FileName:    fileName,
		ContentType: contentType,
	}, nil
}

func (s *ReportService) GetStatus(ctx context.Context, reportID, requesterID int64, role string) (*ReportExportData, error) {
	report, err := s.lifecycleRepo.FindByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentReportNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := assessmentdomain.ValidateReportAccess(report, requesterID, role); err != nil {
		return nil, err
	}
	return buildReportExportDataFromModel(report), nil
}

func (s *ReportService) runAsyncReport(reportID int64, fn func(context.Context) error) {
	if s.baseCtx == nil {
		s.logger.Error("报告异步任务未启动", zap.Int64("report_id", reportID))
		return
	}
	s.tasks.Add(1)
	go func() {
		defer s.tasks.Done()
		taskCtx := s.baseCtx

		select {
		case s.workerPool <- struct{}{}:
		case <-s.baseCtx.Done():
			s.markFailed(taskCtx, reportID, s.baseCtx.Err())
			return
		}
		defer func() {
			<-s.workerPool
			if recovered := recover(); recovered != nil {
				s.markFailed(taskCtx, reportID, fmt.Errorf("报告任务崩溃: %v", recovered))
			}
		}()

		ctx, cancel := context.WithTimeout(s.baseCtx, s.config.ClassTimeout)
		taskCtx = ctx
		defer cancel()

		if err := fn(ctx); err != nil {
			s.markFailed(ctx, reportID, err)
		}
	}()
}

func (s *ReportService) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("report service close requires context")
	}
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	go func() {
		s.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *ReportService) generatePersonalReport(ctx context.Context, reportID, userID int64, format string) (string, time.Time, error) {
	data, err := s.buildPersonalReportData(ctx, userID)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypePersonal, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateClassReport(ctx context.Context, reportID int64, className, format string, window classwindow.Range) (string, time.Time, error) {
	data, err := s.buildClassReportData(ctx, className, window)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypeClass, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateContestExport(ctx context.Context, reportID, contestID int64, format string) (string, time.Time, error) {
	data, err := s.buildContestExportData(ctx, contestID)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypeContest, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateStudentReviewArchive(ctx context.Context, reportID, studentID int64, format string) (string, time.Time, error) {
	data, err := s.buildStudentReviewArchiveData(ctx, studentID)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypeReview, format)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewArchive(reportID int64, archive *assessmentqry.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypeAWDReviewArchive, assessmententity.ReportFormatZIP)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := RenderAWDReviewArchiveZip(filePath, archive); err != nil {
		return "", time.Time{}, err
	}
	return filePath, reportNow().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewReport(reportID int64, archive *assessmentqry.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	filePath, err := s.reportFilePath(reportID, assessmententity.ReportTypeAWDReviewReport, assessmententity.ReportFormatPDF)
	if err != nil {
		return "", time.Time{}, apperror.ErrInternal.WithCause(err)
	}
	if err := RenderAWDReviewReportPDF(filePath, archive); err != nil {
		return "", time.Time{}, err
	}
	return filePath, reportNow().Add(s.config.FileTTL), nil
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

	var summaryResp *teachingquerycontracts.TeacherClassSummary
	var trendResp *teachingquerycontracts.TeacherClassTrend
	var reviewResp *teachingquerycontracts.TeacherClassReview
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

func (s *ReportService) buildStudentReviewArchiveData(ctx context.Context, studentID int64) (*ReviewArchiveData, error) {
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrNotFound
	}

	stats, err := s.personalRepo.GetPersonalStats(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	totalChallenges, err := s.reviewArchiveRepo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	timeline, err := s.reviewArchiveRepo.GetStudentTimeline(ctx, studentID, 200, 0)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	evidence, err := s.reviewArchiveRepo.GetStudentEvidence(ctx, studentID, teachingevidence.Query{})
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	writeups, err := s.reviewArchiveRepo.ListStudentWriteups(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	manualReviews, err := s.reviewArchiveRepo.ListStudentManualReviews(ctx, studentID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var skillProfile []*assessmentcontracts.SkillDimension
	if s.assessmentService != nil {
		skillProfileResp, skillErr := s.assessmentService.GetSkillProfile(ctx, studentID)
		if skillErr != nil {
			return nil, apperror.ErrInternal.WithCause(skillErr)
		}
		skillProfile = skillProfileResp.Dimensions
	}

	summary := buildReviewArchiveSummary(int(totalChallenges), stats, timeline, evidence, writeups, manualReviews)

	return &ReviewArchiveData{
		GeneratedAt: time.Now().UTC(),
		Student: ReviewArchiveStudent{
			ID:        student.ID,
			Username:  student.Username,
			Name:      student.Name,
			ClassName: student.ClassName,
		},
		Summary:             summary,
		SkillProfile:        skillProfile,
		Timeline:            timeline,
		Evidence:            evidence,
		Writeups:            writeups,
		ManualReviews:       manualReviews,
		TeacherObservations: buildReviewArchiveObservations(summary, skillProfile, timeline, evidence, writeups, manualReviews),
	}, nil
}

func buildReviewArchiveSummary(
	totalChallenges int,
	stats *assessmentdomain.PersonalReportStats,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) assessmentdomain.ReviewArchiveSummary {
	summary := assessmentdomain.ReviewArchiveSummary{
		TotalChallenges:        totalChallenges,
		TimelineEventCount:     len(timeline),
		EvidenceEventCount:     len(evidence),
		WriteupCount:           len(writeups),
		ManualReviewCount:      len(manualReviews),
		CorrectSubmissionCount: countCorrectSubmissions(timeline, evidence),
		LastActivityAt:         latestReviewArchiveActivity(timeline, evidence, writeups, manualReviews),
	}
	if stats != nil {
		summary.TotalSolved = stats.TotalSolved
		summary.TotalScore = stats.TotalScore
		summary.Rank = stats.Rank
		summary.TotalAttempts = stats.TotalAttempts
	}
	return summary
}

func countCorrectSubmissions(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	if stats, ok := reviewArchiveSubmissionStatsFromEvidence(evidence); ok {
		if !stats.HasChallengeEvidence {
			challengeSuccessCount := countCorrectTimelineChallengeSubmissions(timeline)
			stats.ChallengeSuccessCount += challengeSuccessCount
			stats.SuccessCount += challengeSuccessCount
		}
		if !stats.HasAWDEvidence {
			awdSuccessCount := countCorrectTimelineAWDSubmissions(timeline)
			stats.AWDSuccessCount += awdSuccessCount
			stats.SuccessCount += awdSuccessCount
		}
		return stats.SuccessCount
	}
	return countCorrectTimelineChallengeSubmissions(timeline) + countCorrectTimelineAWDSubmissions(timeline)
}

func latestReviewArchiveActivity(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) *time.Time {
	var latest *time.Time
	record := func(candidate *time.Time) {
		if candidate == nil || candidate.IsZero() {
			return
		}
		if latest == nil || candidate.After(*latest) {
			copyValue := *candidate
			latest = &copyValue
		}
	}

	for _, item := range timeline {
		record(&item.Timestamp)
	}
	for _, item := range evidence {
		if !includeEvidenceInPersonalActivity(item) {
			continue
		}
		record(&item.Timestamp)
	}
	for _, item := range writeups {
		if item.PublishedAt != nil {
			record(item.PublishedAt)
			continue
		}
		record(&item.UpdatedAt)
	}
	for _, item := range manualReviews {
		record(&item.SubmittedAt)
	}
	return latest
}

func buildReviewArchiveObservations(
	summary assessmentdomain.ReviewArchiveSummary,
	skillProfile []*assessmentcontracts.SkillDimension,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) assessmentdomain.ReviewArchiveTeacherObservations {
	snapshot := buildReviewArchiveTeachingFactSnapshot(summary, skillProfile, timeline, evidence, writeups, manualReviews)
	evaluation := teachingadvice.EvaluateStudent(snapshot)
	adviceItems := teachingadvice.BuildReviewArchiveObservations(snapshot, evaluation)

	items := make([]assessmentdomain.ReviewArchiveObservation, 0, len(adviceItems))
	for _, item := range adviceItems {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Code:      item.Code,
			Label:     item.Label,
			Severity:  string(item.Severity),
			Dimension: item.Dimension,
			Summary:   item.Summary,
			Evidence:  item.Evidence,
			Action:    item.Action,
		})
	}
	return assessmentdomain.ReviewArchiveTeacherObservations{Items: items}
}

func buildReviewArchiveTeachingFactSnapshot(
	summary assessmentdomain.ReviewArchiveSummary,
	skillProfile []*assessmentcontracts.SkillDimension,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) teachingadvice.StudentFactSnapshot {
	recentEventCount, activeDays := recentReviewArchiveActivityStats(time.Now().UTC(), timeline, evidence, writeups, manualReviews)
	submissionStats := buildReviewArchiveSubmissionStats(summary, timeline, evidence)
	snapshot := teachingadvice.StudentFactSnapshot{
		ActiveDays7d:           activeDays,
		RecentEventCount7d:     recentEventCount,
		LastActivityAt:         summary.LastActivityAt,
		CorrectSubmissionCount: submissionStats.SuccessCount,
		WrongSubmissionCount:   submissionStats.FailureCount,
		ChallengeSuccessCount:  submissionStats.ChallengeSuccessCount,
		SubmissionSuccessCount: submissionStats.SuccessCount,
		SubmissionFailureCount: submissionStats.FailureCount,
		WriteupCount:           summary.WriteupCount,
		ApprovedReviewCount:    countApprovedManualReviews(manualReviews),
		Dimensions:             make([]teachingadvice.DimensionFact, 0, len(taxonomy.AllDimensions)),
	}

	factMap := make(map[string]*teachingadvice.DimensionFact, len(taxonomy.AllDimensions))
	for _, dimension := range taxonomy.AllDimensions {
		dimensionCopy := dimension
		factMap[dimension] = &teachingadvice.DimensionFact{Dimension: dimensionCopy}
	}

	snapshot.MaxWrongStreak = submissionStats.MaxWrongStreak
	snapshot.HandsOnEventCount = countReviewArchiveAWDHandsOnEvidence(timeline, evidence)
	snapshot.AWDAttemptCount = countReviewArchiveAWDAttempts(timeline, evidence)
	snapshot.AWDSuccessCount = submissionStats.AWDSuccessCount

	for _, dimension := range skillProfile {
		if dimension == nil {
			continue
		}
		fact := ensureReviewArchiveDimensionFact(factMap, dimension.Dimension)
		if fact == nil {
			continue
		}
		fact.ProfileScore = dimension.Score
	}

	for _, item := range evidence {
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		switch item.Type {
		case teachingevidence.EventTypeChallengeSubmission, teachingevidence.EventTypeAWDAttackSubmission:
			if item.Type == teachingevidence.EventTypeAWDAttackSubmission && !isStudentScopedAWDAttackEvidence(item) {
				continue
			}
			fact.AttemptCount++
			if success, tracked := extractEvidenceSubmissionResult(item); tracked && success {
				fact.SuccessCount++
			}
			fact.EvidenceCount++
		case teachingevidence.EventTypeInstanceAccess, teachingevidence.EventTypeInstanceProxy, teachingevidence.EventTypeAWDTraffic:
			fact.EvidenceCount++
		}
	}

	for _, item := range writeups {
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		fact.EvidenceCount++
	}

	for _, item := range manualReviews {
		if item.ReviewStatus != "approved" {
			continue
		}
		fact := ensureReviewArchiveDimensionFact(factMap, item.Category)
		if fact == nil {
			continue
		}
		fact.EvidenceCount++
	}

	for _, dimension := range taxonomy.AllDimensions {
		fact := ensureReviewArchiveDimensionFact(factMap, dimension)
		if fact == nil {
			continue
		}
		snapshot.Dimensions = append(snapshot.Dimensions, *fact)
	}

	return snapshot
}

func recentReviewArchiveActivityStats(
	referenceTime time.Time,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) (int, int) {
	if referenceTime.IsZero() {
		referenceTime = time.Now().UTC()
	}
	cutoff := referenceTime.AddDate(0, 0, -7)
	activeDays := make(map[string]struct{})
	recentEventCount := 0

	record := func(timestamp time.Time) {
		if timestamp.IsZero() || timestamp.Before(cutoff) {
			return
		}
		recentEventCount++
		activeDays[timestamp.UTC().Format("2006-01-02")] = struct{}{}
	}

	if len(evidence) > 0 {
		for _, item := range evidence {
			if !includeEvidenceInPersonalActivity(item) {
				continue
			}
			record(item.Timestamp)
		}
	} else {
		for _, item := range timeline {
			record(item.Timestamp)
		}
	}

	for _, item := range writeups {
		if item.PublishedAt != nil {
			record(item.PublishedAt.UTC())
			continue
		}
		record(item.UpdatedAt)
	}

	for _, item := range manualReviews {
		record(item.SubmittedAt)
	}

	return recentEventCount, len(activeDays)
}

func ensureReviewArchiveDimensionFact(
	facts map[string]*teachingadvice.DimensionFact,
	dimension string,
) *teachingadvice.DimensionFact {
	normalized := strings.ToLower(strings.TrimSpace(dimension))
	if normalized == "" {
		return nil
	}
	if fact, ok := facts[normalized]; ok {
		return fact
	}
	fact := &teachingadvice.DimensionFact{Dimension: normalized}
	facts[normalized] = fact
	return fact
}

func hasSubmittedWriteup(writeups []assessmentdomain.ReviewArchiveWriteupItem) bool {
	for _, item := range writeups {
		if item.SubmissionStatus == "published" || item.SubmissionStatus == "submitted" {
			return true
		}
	}
	return false
}

func countApprovedManualReviews(items []assessmentdomain.ReviewArchiveManualReviewItem) int {
	count := 0
	for _, item := range items {
		if item.ReviewStatus == "approved" {
			count++
		}
	}
	return count
}

func hasApprovedManualReview(items []assessmentdomain.ReviewArchiveManualReviewItem) bool {
	return countApprovedManualReviews(items) > 0
}

func hasRepeatedWrongSubmissions(evidence []assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	streak := 0
	for _, item := range evidence {
		isCorrect, tracked := extractEvidenceSubmissionResult(item)
		if !tracked {
			continue
		}
		if isCorrect {
			streak = 0
			continue
		}
		streak++
		if streak >= 2 {
			return true
		}
	}
	return false
}

func hasHandsOnExploit(evidence []assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	for _, item := range evidence {
		if item.Type == teachingevidence.EventTypeInstanceAccess ||
			item.Type == teachingevidence.EventTypeInstanceProxy ||
			item.Type == teachingevidence.EventTypeAWDTraffic {
			return true
		}
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission && isStudentScopedAWDAttackEvidence(item) {
			return true
		}
	}
	return false
}

func isCorrectTimelineSubmission(item assessmentdomain.ReviewArchiveTimelineEvent) bool {
	if item.IsCorrect == nil || !*item.IsCorrect {
		return false
	}
	return item.Type == "flag_submit" || item.Type == "awd_attack_submit"
}

func isCorrectEvidenceSubmission(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	isCorrect, tracked := extractEvidenceSubmissionResult(item)
	return tracked && isCorrect
}

func extractEvidenceSubmissionResult(item assessmentdomain.ReviewArchiveEvidenceEvent) (bool, bool) {
	if item.Meta == nil {
		return false, false
	}

	switch item.Type {
	case teachingevidence.EventTypeChallengeSubmission:
		isCorrect, ok := item.Meta["is_correct"].(bool)
		return isCorrect, ok
	case teachingevidence.EventTypeAWDAttackSubmission:
		if !isStudentScopedAWDAttackEvidence(item) {
			return false, false
		}
		isCorrect, ok := item.Meta["is_success"].(bool)
		return isCorrect, ok
	default:
		return false, false
	}
}

type reviewArchiveSubmissionStats struct {
	ChallengeSuccessCount int
	SuccessCount          int
	FailureCount          int
	AWDSuccessCount       int
	MaxWrongStreak        int
	HasChallengeEvidence  bool
	HasAWDEvidence        bool
}

func buildReviewArchiveSubmissionStats(
	summary assessmentdomain.ReviewArchiveSummary,
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) reviewArchiveSubmissionStats {
	if stats, ok := reviewArchiveSubmissionStatsFromEvidence(evidence); ok {
		if !stats.HasChallengeEvidence {
			challengeSuccessCount := countCorrectTimelineChallengeSubmissions(timeline)
			stats.ChallengeSuccessCount += challengeSuccessCount
			stats.SuccessCount += challengeSuccessCount
			stats.FailureCount += max(summary.TotalAttempts-stats.ChallengeSuccessCount, 0)
		}
		if !stats.HasAWDEvidence {
			awdSuccessCount := countCorrectTimelineAWDSubmissions(timeline)
			stats.AWDSuccessCount += awdSuccessCount
			stats.SuccessCount += awdSuccessCount
		}
		return stats
	}

	stats := reviewArchiveSubmissionStats{}
	stats.ChallengeSuccessCount = countCorrectTimelineChallengeSubmissions(timeline)
	stats.AWDSuccessCount = countCorrectTimelineAWDSubmissions(timeline)
	stats.SuccessCount = stats.ChallengeSuccessCount + stats.AWDSuccessCount
	stats.FailureCount = max(summary.TotalAttempts-stats.ChallengeSuccessCount, 0)
	return stats
}

func reviewArchiveSubmissionStatsFromEvidence(
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) (reviewArchiveSubmissionStats, bool) {
	type trackedEvent struct {
		timestamp time.Time
		success   bool
	}

	stats := reviewArchiveSubmissionStats{}
	trackedEvents := make([]trackedEvent, 0, len(evidence))
	trackedCount := 0

	for _, item := range evidence {
		isCorrect, tracked := extractEvidenceSubmissionResult(item)
		if !tracked {
			continue
		}
		trackedCount++
		trackedEvents = append(trackedEvents, trackedEvent{timestamp: item.Timestamp, success: isCorrect})
		if item.Type == teachingevidence.EventTypeChallengeSubmission {
			stats.HasChallengeEvidence = true
		}
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission {
			stats.HasAWDEvidence = true
		}
		if isCorrect {
			stats.SuccessCount++
			if item.Type == teachingevidence.EventTypeChallengeSubmission {
				stats.ChallengeSuccessCount++
			}
			if item.Type == teachingevidence.EventTypeAWDAttackSubmission {
				stats.AWDSuccessCount++
			}
			continue
		}
		stats.FailureCount++
	}

	if trackedCount == 0 {
		return reviewArchiveSubmissionStats{}, false
	}

	sort.Slice(trackedEvents, func(i, j int) bool {
		return trackedEvents[i].timestamp.Before(trackedEvents[j].timestamp)
	})

	currentWrongStreak := 0
	for _, event := range trackedEvents {
		if event.success {
			currentWrongStreak = 0
			continue
		}
		currentWrongStreak++
		if currentWrongStreak > stats.MaxWrongStreak {
			stats.MaxWrongStreak = currentWrongStreak
		}
	}

	return stats, true
}

func countCorrectTimelineChallengeSubmissions(timeline []assessmentdomain.ReviewArchiveTimelineEvent) int {
	count := 0
	for _, item := range timeline {
		if item.Type == "flag_submit" && item.IsCorrect != nil && *item.IsCorrect {
			count++
		}
	}
	return count
}

func countReviewArchiveAWDHandsOnEvidence(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	handsOnCount := 0
	for _, item := range evidence {
		switch item.Type {
		case teachingevidence.EventTypeAWDTraffic:
			handsOnCount++
		case teachingevidence.EventTypeAWDAttackSubmission:
			if isStudentScopedAWDAttackEvidence(item) {
				handsOnCount++
			}
		}
	}
	if handsOnCount > 0 {
		return handsOnCount
	}
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" {
			handsOnCount++
		}
	}
	return handsOnCount
}

func countReviewArchiveAWDAttempts(
	timeline []assessmentdomain.ReviewArchiveTimelineEvent,
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
) int {
	attemptCount := 0
	for _, item := range evidence {
		if item.Type == teachingevidence.EventTypeAWDAttackSubmission && isStudentScopedAWDAttackEvidence(item) {
			attemptCount++
		}
	}
	if attemptCount > 0 {
		return attemptCount
	}
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" {
			attemptCount++
		}
	}
	return attemptCount
}

func includeEvidenceInPersonalActivity(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	switch item.Type {
	case teachingevidence.EventTypeAWDTraffic:
		return false
	case teachingevidence.EventTypeAWDAttackSubmission:
		return isStudentScopedAWDAttackEvidence(item)
	default:
		return true
	}
}

func isStudentScopedAWDAttackEvidence(item assessmentdomain.ReviewArchiveEvidenceEvent) bool {
	if item.Type != teachingevidence.EventTypeAWDAttackSubmission {
		return false
	}
	if item.Meta == nil {
		return true
	}
	scope, ok := item.Meta["scope"].(string)
	if !ok {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(scope), "student")
}

func countCorrectTimelineAWDSubmissions(timeline []assessmentdomain.ReviewArchiveTimelineEvent) int {
	count := 0
	for _, item := range timeline {
		if item.Type == "awd_attack_submit" && item.IsCorrect != nil && *item.IsCorrect {
			count++
		}
	}
	return count
}

func (s *ReportService) renderReport(filePath, format string, data any) error {
	switch format {
	case assessmententity.ReportFormatJSON:
		return writeJSONReport(filePath, data)
	case assessmententity.ReportFormatExcel:
		switch payload := data.(type) {
		case *personalReportData:
			return writePersonalExcel(filePath, payload)
		case *classReportData:
			return writeClassExcel(filePath, payload)
		}
	default:
		switch payload := data.(type) {
		case *personalReportData:
			return writePersonalPDF(filePath, payload)
		case *classReportData:
			return writeClassPDF(filePath, payload)
		}
	}
	return apperror.ErrInternal.WithCause(fmt.Errorf("unsupported report payload"))
}

func (s *ReportService) reportFilePath(reportID int64, reportType, format string) (string, error) {
	storageDir := filepath.Clean(s.config.StorageDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return "", err
	}
	extension := reportFileExtension(format)
	fileName := fmt.Sprintf("%s-%d-%d.%s", reportType, reportID, time.Now().Unix(), extension)
	return filepath.Join(storageDir, fileName), nil
}

func (s *ReportService) safeReportPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absStorage, err := filepath.Abs(s.config.StorageDir)
	if err != nil {
		return "", err
	}
	prefix := absStorage + string(os.PathSeparator)
	if absPath != absStorage && !strings.HasPrefix(absPath, prefix) {
		return "", fmt.Errorf("unsafe path")
	}
	return absPath, nil
}

func (s *ReportService) normalizeFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case assessmententity.ReportFormatJSON:
		return assessmententity.ReportFormatJSON
	case assessmententity.ReportFormatExcel:
		return assessmententity.ReportFormatExcel
	case assessmententity.ReportFormatPDF:
		return assessmententity.ReportFormatPDF
	default:
		return s.config.DefaultFormat
	}
}

func (s *ReportService) normalizeArchiveFormat(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatJSON) {
		return assessmententity.ReportFormatJSON
	}
	return assessmententity.ReportFormatJSON
}

func (s *ReportService) parseClassWindow(req CreateClassReportInput) (classwindow.Range, error) {
	window, err := classwindow.Parse(reportNow(), req.FromDate, req.ToDate)
	if err != nil {
		return classwindow.Range{}, apperror.ErrInvalidParams.WithMessage(err.Error())
	}
	return window, nil
}

func mapClassSummary(summary *queryports.ClassSummary) *teachingquerycontracts.TeacherClassSummary {
	if summary == nil {
		return nil
	}
	return &teachingquerycontracts.TeacherClassSummary{
		ClassName:          summary.ClassName,
		StudentCount:       summary.StudentCount,
		AverageSolved:      summary.AverageSolved,
		ActiveStudentCount: summary.ActiveStudentCount,
		ActiveRate:         summary.ActiveRate,
		RecentEventCount:   summary.RecentEventCount,
	}
}

func mapClassTrend(trend *queryports.ClassTrend) *teachingquerycontracts.TeacherClassTrend {
	if trend == nil {
		return nil
	}
	points := make([]teachingquerycontracts.TeacherClassTrendPoint, 0, len(trend.Points))
	for _, point := range trend.Points {
		points = append(points, teachingquerycontracts.TeacherClassTrendPoint{
			Date:               point.Date,
			ActiveStudentCount: point.ActiveStudentCount,
			EventCount:         point.EventCount,
			SolveCount:         point.SolveCount,
		})
	}
	return &teachingquerycontracts.TeacherClassTrend{
		ClassName: trend.ClassName,
		Points:    points,
	}
}

func mapClassReview(review *classreview.Response) *teachingquerycontracts.TeacherClassReview {
	if review == nil {
		return nil
	}
	items := make([]teachingquerycontracts.TeacherClassReviewItem, 0, len(review.Items))
	for _, item := range review.Items {
		items = append(items, teachingquerycontracts.TeacherClassReviewItem{
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
	return &teachingquerycontracts.TeacherClassReview{
		ClassName: review.ClassName,
		Items:     items,
	}
}

func mapReviewStudents(students []classreview.ReviewStudentRef) []teachingquerycontracts.TeacherReviewStudentRef {
	items := make([]teachingquerycontracts.TeacherReviewStudentRef, 0, len(students))
	for _, student := range students {
		items = append(items, teachingquerycontracts.TeacherReviewStudentRef{
			ID:       student.ID,
			Username: student.Username,
			Name:     student.Name,
		})
	}
	return items
}

func mapReviewRecommendation(item *classreview.RecommendationItem) *teachingquerycontracts.TeacherRecommendationItem {
	if item == nil {
		return nil
	}
	return &teachingquerycontracts.TeacherRecommendationItem{
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

func reportFileExtension(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatJSON) {
		return "json"
	}
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatZIP) {
		return "zip"
	}
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatExcel) {
		return "xlsx"
	}
	return "pdf"
}

func reportOutputFormat(report *assessmententity.Report) string {
	if report == nil {
		return assessmententity.ReportFormatPDF
	}
	switch report.Type {
	case assessmententity.ReportTypeAWDReviewArchive:
		return assessmententity.ReportFormatZIP
	case assessmententity.ReportTypeAWDReviewReport:
		return assessmententity.ReportFormatPDF
	default:
		return report.Format
	}
}

func reportDownloadFileName(report *assessmententity.Report) string {
	return fmt.Sprintf("%s-report-%d.%s", report.Type, report.ID, reportFileExtension(reportOutputFormat(report)))
}

func reportContentType(format string) string {
	return mime.TypeByExtension("." + reportFileExtension(format))
}

func (s *ReportService) markFailed(ctx context.Context, reportID int64, err error) {
	if s.lifecycleRepo == nil {
		return
	}
	if ctx == nil {
		s.logger.Error("report_mark_failed_missing_context", zap.Int64("report_id", reportID))
		return
	}

	message := "报告生成失败"
	if err != nil {
		message = err.Error()
	}
	markCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if updateErr := s.lifecycleRepo.MarkFailed(markCtx, reportID, message); updateErr != nil {
		s.logger.Error("report_mark_failed_error", zap.Int64("report_id", reportID), zap.Error(updateErr))
	}
}

func buildReportExportData(reportID int64, status string, expiresAt time.Time) *ReportExportData {
	report := &assessmententity.Report{
		ID:        reportID,
		Status:    status,
		ExpiresAt: nil,
	}
	if !expiresAt.IsZero() {
		report.ExpiresAt = &expiresAt
	}
	return buildReportExportDataFromModel(report)
}

func buildReportExportDataFromModel(report *assessmententity.Report) *ReportExportData {
	resp := assessmentCommandResponseMapperInst.ToReportExportDataBasePtr(report)
	if report.Status == assessmententity.ReportStatusReady {
		downloadURL := fmt.Sprintf("/api/v1/reports/%d/download", report.ID)
		resp.DownloadURL = &downloadURL
		if report.ExpiresAt != nil && !report.ExpiresAt.IsZero() {
			expires := report.ExpiresAt.Format(time.RFC3339)
			resp.ExpiresAt = &expires
		}
	}
	if report.Status == assessmententity.ReportStatusFailed {
		resp.ErrorMessage = normalizeOptionalReportErrorMessage(report.ErrorMsg)
	}
	return resp
}

func normalizeOptionalReportErrorMessage(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func writePersonalPDF(filePath string, data *personalReportData) error {
	pdf, err := newReportPDF()
	if err != nil {
		return err
	}
	addReportTitle(pdf, "个人训练报告")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "用户名", Value: sanitizePDFText(data.User.Username)},
		{Label: "班级", Value: sanitizePDFText(data.User.ClassName)},
		{Label: "总分", Value: fmt.Sprintf("%d", data.Stats.TotalScore)},
		{Label: "排名", Value: fmt.Sprintf("%d", data.Stats.Rank)},
		{Label: "解出题数", Value: fmt.Sprintf("%d", data.Stats.TotalSolved)},
		{Label: "尝试次数", Value: fmt.Sprintf("%d", data.Stats.TotalAttempts)},
	})
	addDimensionChart(pdf, "能力画像", skillProfileChartRows(data.SkillProfile))
	addDimensionStatsTable(pdf, "维度明细", data.DimensionStats)
	return pdf.OutputFileAndClose(filePath)
}

func writeClassPDF(filePath string, data *classReportData) error {
	pdf, err := newReportPDF()
	if err != nil {
		return err
	}
	addReportTitle(pdf, "班级训练报告")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "班级", Value: sanitizePDFText(data.ClassName)},
		{Label: "统计窗口", Value: fmt.Sprintf("%s 至 %s（%d 天）", data.Window.FromDate, data.Window.ToDate, data.Window.Days)},
		{Label: "学生总数", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "平均分", Value: fmt.Sprintf("%.2f", data.AverageScore)},
		{Label: "活跃率", Value: fmt.Sprintf("%.0f%%", reviewSummaryActiveRate(data.Summary))},
		{Label: "近期事件数", Value: fmt.Sprintf("%d", reviewSummaryRecentEvents(data.Summary))},
	})
	addAverageChart(pdf, "维度平均分", data.DimensionAverages)
	addClassTrendTable(pdf, "趋势快照", data.Trend)
	addDistributionTable(pdf, "分类分布", data.CategoryDistribution)
	addDistributionTable(pdf, "难度分布", data.DifficultyDistribution)
	addTopStudentsTable(pdf, "班级前列学生", data.TopStudents)
	addContestMigrationSection(pdf, data.ContestMigration)
	addClassReviewOutlineTable(pdf, data.Review)
	return pdf.OutputFileAndClose(filePath)
}

func writePersonalExcel(filePath string, data *personalReportData) error {
	file := excelize.NewFile()
	defer file.Close()

	summarySheet := "Summary"
	file.SetSheetName("Sheet1", summarySheet)
	detailsSheet := "Dimensions"
	file.NewSheet(detailsSheet)

	headerStyle := mustNewExcelStyle(file, &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E2F3"}},
	})

	writePairs(file, summarySheet, []summaryLine{
		{Label: "Username", Value: data.User.Username},
		{Label: "Class", Value: data.User.ClassName},
		{Label: "Total Score", Value: fmt.Sprintf("%d", data.Stats.TotalScore)},
		{Label: "Rank", Value: fmt.Sprintf("%d", data.Stats.Rank)},
		{Label: "Solved", Value: fmt.Sprintf("%d", data.Stats.TotalSolved)},
		{Label: "Attempts", Value: fmt.Sprintf("%d", data.Stats.TotalAttempts)},
	}, headerStyle)

	file.SetCellValue(summarySheet, "A10", "Dimension")
	file.SetCellValue(summarySheet, "B10", "Score")
	file.SetCellStyle(summarySheet, "A10", "B10", headerStyle)
	for idx, dimension := range data.SkillProfile {
		row := idx + 11
		file.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), dimension.Dimension)
		file.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), dimension.Score)
	}

	file.SetCellValue(detailsSheet, "A1", "Dimension")
	file.SetCellValue(detailsSheet, "B1", "Solved")
	file.SetCellValue(detailsSheet, "C1", "Total")
	file.SetCellStyle(detailsSheet, "A1", "C1", headerStyle)
	for idx, stat := range data.DimensionStats {
		row := idx + 2
		file.SetCellValue(detailsSheet, fmt.Sprintf("A%d", row), stat.Dimension)
		file.SetCellValue(detailsSheet, fmt.Sprintf("B%d", row), stat.Solved)
		file.SetCellValue(detailsSheet, fmt.Sprintf("C%d", row), stat.Total)
	}

	if len(data.SkillProfile) > 0 {
		_ = file.AddChart(summarySheet, "D2", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{{
				Name:       fmt.Sprintf("%s!$B$10", summarySheet),
				Categories: fmt.Sprintf("%s!$A$11:$A$%d", summarySheet, len(data.SkillProfile)+10),
				Values:     fmt.Sprintf("%s!$B$11:$B$%d", summarySheet, len(data.SkillProfile)+10),
			}},
			Title:  []excelize.RichTextRun{{Text: "Skill Profile"}},
			Legend: excelize.ChartLegend{Position: "bottom"},
		})
	}

	setReportSheetLayout(file, summarySheet)
	setReportSheetLayout(file, detailsSheet)
	return file.SaveAs(filePath)
}

func writeClassExcel(filePath string, data *classReportData) error {
	file := excelize.NewFile()
	defer file.Close()

	summarySheet := "Summary"
	file.SetSheetName("Sheet1", summarySheet)
	trendSheet := "Trend"
	reviewSheet := "Review"
	categorySheet := "Category"
	difficultySheet := "Difficulty"
	migrationSheet := "Migration"
	topSheet := "TopStudents"
	file.NewSheet(trendSheet)
	file.NewSheet(reviewSheet)
	file.NewSheet(categorySheet)
	file.NewSheet(difficultySheet)
	file.NewSheet(migrationSheet)
	file.NewSheet(topSheet)

	headerStyle := mustNewExcelStyle(file, &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FCE4D6"}},
	})

	writePairs(file, summarySheet, []summaryLine{
		{Label: "Class", Value: data.ClassName},
		{Label: "Window", Value: fmt.Sprintf("%s ~ %s (%d days)", data.Window.FromDate, data.Window.ToDate, data.Window.Days)},
		{Label: "Total Students", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "Average Score", Value: fmt.Sprintf("%.2f", data.AverageScore)},
		{Label: "Active Rate", Value: fmt.Sprintf("%.0f%%", reviewSummaryActiveRate(data.Summary))},
		{Label: "Recent Events", Value: fmt.Sprintf("%d", reviewSummaryRecentEvents(data.Summary))},
	}, headerStyle)

	file.SetCellValue(summarySheet, "A9", "Dimension")
	file.SetCellValue(summarySheet, "B9", "Average Score")
	file.SetCellStyle(summarySheet, "A9", "B9", headerStyle)
	for idx, dimension := range data.DimensionAverages {
		row := idx + 10
		file.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), dimension.Dimension)
		file.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), dimension.AvgScore)
	}

	file.SetCellValue(trendSheet, "A1", "Date")
	file.SetCellValue(trendSheet, "B1", "Active Students")
	file.SetCellValue(trendSheet, "C1", "Events")
	file.SetCellValue(trendSheet, "D1", "Solves")
	file.SetCellStyle(trendSheet, "A1", "D1", headerStyle)
	for idx, point := range safeTrendPoints(data.Trend) {
		row := idx + 2
		file.SetCellValue(trendSheet, fmt.Sprintf("A%d", row), point.Date)
		file.SetCellValue(trendSheet, fmt.Sprintf("B%d", row), point.ActiveStudentCount)
		file.SetCellValue(trendSheet, fmt.Sprintf("C%d", row), point.EventCount)
		file.SetCellValue(trendSheet, fmt.Sprintf("D%d", row), point.SolveCount)
	}

	writeDistributionSheet(file, categorySheet, headerStyle, data.CategoryDistribution)
	writeDistributionSheet(file, difficultySheet, headerStyle, data.DifficultyDistribution)
	writeReviewSheet(file, reviewSheet, headerStyle, data.Review)
	writeContestMigrationSheet(file, migrationSheet, headerStyle, data.ContestMigration)

	file.SetCellValue(topSheet, "A1", "Rank")
	file.SetCellValue(topSheet, "B1", "Username")
	file.SetCellValue(topSheet, "C1", "Total Score")
	file.SetCellStyle(topSheet, "A1", "C1", headerStyle)
	for idx, student := range data.TopStudents {
		row := idx + 2
		file.SetCellValue(topSheet, fmt.Sprintf("A%d", row), student.Rank)
		file.SetCellValue(topSheet, fmt.Sprintf("B%d", row), student.Username)
		file.SetCellValue(topSheet, fmt.Sprintf("C%d", row), student.TotalScore)
	}

	if len(data.DimensionAverages) > 0 {
		_ = file.AddChart(summarySheet, "D2", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{{
				Name:       fmt.Sprintf("%s!$B$9", summarySheet),
				Categories: fmt.Sprintf("%s!$A$10:$A$%d", summarySheet, len(data.DimensionAverages)+9),
				Values:     fmt.Sprintf("%s!$B$10:$B$%d", summarySheet, len(data.DimensionAverages)+9),
			}},
			Title:  []excelize.RichTextRun{{Text: "Dimension Average"}},
			Legend: excelize.ChartLegend{Position: "bottom"},
		})
	}

	setReportSheetLayout(file, summarySheet)
	setReportSheetLayout(file, trendSheet)
	setReportSheetLayout(file, reviewSheet)
	setReportSheetLayout(file, categorySheet)
	setReportSheetLayout(file, difficultySheet)
	setReportSheetLayout(file, migrationSheet)
	setReportSheetLayout(file, topSheet)
	return file.SaveAs(filePath)
}

func writeJSONReport(filePath string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

type summaryLine struct {
	Label string
	Value string
}

func newReportPDF() (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	if err := registerReportPDFFonts(pdf); err != nil {
		return nil, err
	}
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddPage()
	setReportPDFFont(pdf, "", 12)
	return pdf, nil
}

func addReportTitle(pdf *gofpdf.Fpdf, title string) {
	setReportPDFFont(pdf, "B", 20)
	pdf.SetTextColor(24, 39, 75)
	pdf.CellFormat(0, 13, sanitizePDFText(title), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(16, pdf.GetY(), 194, pdf.GetY())
	pdf.Ln(6)
}

func addSummaryBlock(pdf *gofpdf.Fpdf, lines []summaryLine) {
	addReportSectionTitle(pdf, "摘要")
	for _, line := range lines {
		setReportPDFFont(pdf, "B", 11)
		pdf.SetTextColor(45, 57, 84)
		pdf.CellFormat(45, 7, sanitizePDFText(line.Label), "0", 0, "L", false, 0, "")
		setReportPDFFont(pdf, "", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 7, sanitizePDFText(line.Value), "0", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addReportBulletSection(pdf *gofpdf.Fpdf, title string, items []string) {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		filtered = append(filtered, sanitizePDFText(item))
	}
	if len(filtered) == 0 {
		return
	}

	ensurePDFSpace(pdf, 16+float64(len(filtered))*8)
	addReportSectionTitle(pdf, title)
	setReportPDFFont(pdf, "", 11)
	for _, item := range filtered {
		ensurePDFSpace(pdf, 10)
		pdf.MultiCell(0, 7, sanitizePDFText("• "+item), "", "L", false)
	}
	pdf.Ln(2)
}

func addReportSectionTitle(pdf *gofpdf.Fpdf, title string) {
	ensurePDFSpace(pdf, 12)
	setReportPDFFont(pdf, "B", 15)
	pdf.SetFillColor(235, 240, 250)
	pdf.SetTextColor(24, 39, 75)
	pdf.CellFormat(0, 9, sanitizePDFText(title), "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(1.5)
}

func addDimensionChart(pdf *gofpdf.Fpdf, title string, rows []chartRow) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	for _, row := range rows {
		ensurePDFSpace(pdf, 12)
		setReportPDFFont(pdf, "", 10)
		pdf.CellFormat(28, 7, sanitizePDFText(localizeReportTerm(row.Label)), "", 0, "L", false, 0, "")
		x := pdf.GetX()
		y := pdf.GetY() + 2
		pdf.SetFillColor(232, 236, 241)
		pdf.Rect(x, y, 100, 4, "F")
		pdf.SetFillColor(79, 129, 189)
		pdf.Rect(x, y, 100*row.Value, 4, "F")
		pdf.SetX(x + 104)
		pdf.CellFormat(0, 7, fmt.Sprintf("%.0f%%", row.Value*100), "", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addAverageChart(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassDimensionAverage) {
	chartRows := make([]chartRow, 0, len(rows))
	for _, row := range rows {
		chartRows = append(chartRows, chartRow{Label: row.Dimension, Value: row.AvgScore})
	}
	addDimensionChart(pdf, title, chartRows)
}

func addDimensionStatsTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ReportDimensionStat) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"维度", "解出题数", "总题数"})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(70, 7, sanitizePDFText(localizeReportTerm(row.Dimension)), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Solved), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Total), "1", 1, "C", false, 0, "")
	}
}

func addTopStudentsTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassTopStudent) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"排名", "用户名", "分数"})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(90, 7, sanitizePDFText(row.Username), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.TotalScore), "1", 1, "C", false, 0, "")
	}
}

func addClassTrendTable(pdf *gofpdf.Fpdf, title string, trend *teachingquerycontracts.TeacherClassTrend) {
	points := safeTrendPoints(trend)
	if len(points) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"日期", "活跃学生", "事件数", "解题数"}, []float64{42, 36, 36, 36})
	setReportPDFFont(pdf, "", 10)
	for _, point := range points {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{42, 36, 36, 36}, []string{
			point.Date,
			fmt.Sprintf("%d", point.ActiveStudentCount),
			fmt.Sprintf("%d", point.EventCount),
			fmt.Sprintf("%d", point.SolveCount),
		})
	}
}

func addDistributionTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassDistributionStat) {
	if len(rows) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"项目", "解出学生数", "已覆盖", "总数"}, []float64{46, 44, 36, 36})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{46, 44, 36, 36}, []string{
			localizeReportTerm(row.Key),
			fmt.Sprintf("%d", row.SolvedStudents),
			fmt.Sprintf("%d", row.CoveredChallenges),
			fmt.Sprintf("%d", row.TotalChallenges),
		})
	}
}

func addContestMigrationSection(pdf *gofpdf.Fpdf, summary assessmentdomain.ClassContestMigrationSummary) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "竞赛迁移情况", "", 1, "L", false, 0, "")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "参赛学生数", Value: fmt.Sprintf("%d", summary.ParticipatingStudents)},
		{Label: "成功学生数", Value: fmt.Sprintf("%d", summary.SuccessfulStudents)},
		{Label: "攻击事件数", Value: fmt.Sprintf("%d", summary.AttackCount)},
		{Label: "成功事件数", Value: fmt.Sprintf("%d", summary.SuccessCount)},
		{Label: "成功维度", Value: strings.Join(localizeReportTerms(summary.SuccessDimensions), "、")},
	})
}

func addClassReviewOutlineTable(pdf *gofpdf.Fpdf, review *teachingquerycontracts.TeacherClassReview) {
	items := safeReviewItems(review)
	if len(items) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "复盘提要", "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"编码", "级别", "维度", "学生"}, []float64{42, 32, 36, 68})
	setReportPDFFont(pdf, "", 10)
	for _, item := range items {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{42, 32, 36, 68}, []string{
			item.Code,
			localizeReportTerm(item.Severity),
			localizeReportTerm(item.Dimension),
			reviewStudentNames(item.Students),
		})
	}
}

func writePDFTableHeader(pdf *gofpdf.Fpdf, headers []string) {
	pdf.SetFillColor(220, 230, 241)
	setReportPDFFont(pdf, "B", 10)
	widths := []float64{70, 40, 40}
	if len(headers) == 3 && headers[0] == "排名" {
		widths = []float64{30, 90, 30}
	}
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

func writePDFCustomTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFillColor(220, 230, 241)
	setReportPDFFont(pdf, "B", 10)
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

func writePDFTableRow(pdf *gofpdf.Fpdf, widths []float64, values []string) {
	for idx, value := range values {
		align := "L"
		if idx > 0 {
			align = "C"
		}
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(value), "1", 0, align, false, 0, "")
	}
	pdf.Ln(-1)
}

type chartRow struct {
	Label string
	Value float64
}

func skillProfileChartRows(dimensions []*assessmentcontracts.SkillDimension) []chartRow {
	rows := make([]chartRow, 0, len(dimensions))
	for _, dimension := range dimensions {
		rows = append(rows, chartRow{
			Label: dimension.Dimension,
			Value: dimension.Score,
		})
	}
	return rows
}

func ensurePDFSpace(pdf *gofpdf.Fpdf, needed float64) {
	if pdf.GetY()+needed <= 280 {
		return
	}
	pdf.AddPage()
}

func sanitizePDFText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}

func localizeReportTerms(values []string) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, localizeReportTerm(value))
	}
	return items
}

func localizeReportTerm(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "":
		return ""
	case "final":
		return "最终快照"
	case contestcontracts.ContestStatusDraft:
		return "草稿"
	case contestcontracts.ContestStatusRegistration:
		return "报名中"
	case contestcontracts.ContestStatusFrozen:
		return "冻结中"
	case contestcontracts.ContestStatusEnded:
		return "已结束"
	case contestcontracts.AWDRoundStatusPending:
		return "待开始"
	case contestcontracts.ContestStatusRunning:
		return "进行中"
	case contestcontracts.AWDRoundStatusFinished:
		return "已完成"
	case contestcontracts.AWDServiceStatusUp:
		return "正常"
	case contestcontracts.AWDServiceStatusDown:
		return "下线"
	case contestcontracts.AWDServiceStatusCompromised:
		return "失陷"
	case contestcontracts.AWDAttackTypeFlagCapture:
		return "夺旗"
	case contestcontracts.AWDAttackTypeServiceExploit:
		return "服务利用"
	case contestcontracts.AWDAttackSourceManual:
		return "手工记录"
	case contestcontracts.AWDAttackSourceSubmission:
		return "提交记录"
	case contestcontracts.AWDTrafficSourceRuntimeProxy:
		return "代理流量"
	case "critical":
		return "高"
	case "warning":
		return "中"
	case "good":
		return "低"
	case taxonomy.DimensionWeb:
		return "Web"
	case taxonomy.DimensionPwn:
		return "Pwn"
	case taxonomy.DimensionReverse:
		return "逆向"
	case taxonomy.DimensionCrypto:
		return "密码"
	case taxonomy.DimensionMisc:
		return "综合"
	case taxonomy.DimensionForensics:
		return "取证"
	case taxonomy.DifficultyBeginner:
		return "入门"
	case taxonomy.DifficultyEasy:
		return "简单"
	case taxonomy.DifficultyMedium:
		return "中等"
	case taxonomy.DifficultyHard:
		return "困难"
	case taxonomy.DifficultyInsane:
		return "极难"
	default:
		return value
	}
}

func mustNewExcelStyle(file *excelize.File, style *excelize.Style) int {
	styleID, _ := file.NewStyle(style)
	return styleID
}

func writePairs(file *excelize.File, sheet string, rows []summaryLine, headerStyle int) {
	for idx, row := range rows {
		line := idx + 1
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), row.Label)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), row.Value)
		file.SetCellStyle(sheet, fmt.Sprintf("A%d", line), fmt.Sprintf("A%d", line), headerStyle)
	}
}

func setReportSheetLayout(file *excelize.File, sheet string) {
	file.SetColWidth(sheet, "A", "A", 22)
	file.SetColWidth(sheet, "B", "E", 18)
}

func writeDistributionSheet(file *excelize.File, sheet string, headerStyle int, rows []assessmentdomain.ClassDistributionStat) {
	file.SetCellValue(sheet, "A1", "Key")
	file.SetCellValue(sheet, "B1", "Solved Students")
	file.SetCellValue(sheet, "C1", "Covered Challenges")
	file.SetCellValue(sheet, "D1", "Total Challenges")
	file.SetCellStyle(sheet, "A1", "D1", headerStyle)
	for idx, row := range rows {
		line := idx + 2
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), row.Key)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), row.SolvedStudents)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", line), row.CoveredChallenges)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", line), row.TotalChallenges)
	}
}

func writeReviewSheet(file *excelize.File, sheet string, headerStyle int, review *teachingquerycontracts.TeacherClassReview) {
	file.SetCellValue(sheet, "A1", "Code")
	file.SetCellValue(sheet, "B1", "Severity")
	file.SetCellValue(sheet, "C1", "Dimension")
	file.SetCellValue(sheet, "D1", "Students")
	file.SetCellValue(sheet, "E1", "Summary")
	file.SetCellValue(sheet, "F1", "Evidence")
	file.SetCellValue(sheet, "G1", "Action")
	file.SetCellStyle(sheet, "A1", "G1", headerStyle)
	for idx, item := range safeReviewItems(review) {
		line := idx + 2
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), item.Code)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), item.Severity)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", line), item.Dimension)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", line), reviewStudentNames(item.Students))
		file.SetCellValue(sheet, fmt.Sprintf("E%d", line), item.Summary)
		file.SetCellValue(sheet, fmt.Sprintf("F%d", line), item.Evidence)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", line), item.Action)
	}
}

func writeContestMigrationSheet(file *excelize.File, sheet string, headerStyle int, summary assessmentdomain.ClassContestMigrationSummary) {
	writePairs(file, sheet, []summaryLine{
		{Label: "Participating Students", Value: fmt.Sprintf("%d", summary.ParticipatingStudents)},
		{Label: "Successful Students", Value: fmt.Sprintf("%d", summary.SuccessfulStudents)},
		{Label: "Attack Events", Value: fmt.Sprintf("%d", summary.AttackCount)},
		{Label: "Success Events", Value: fmt.Sprintf("%d", summary.SuccessCount)},
		{Label: "Success Dimensions", Value: strings.Join(summary.SuccessDimensions, ", ")},
	}, headerStyle)
}

func safeTrendPoints(trend *teachingquerycontracts.TeacherClassTrend) []teachingquerycontracts.TeacherClassTrendPoint {
	if trend == nil {
		return []teachingquerycontracts.TeacherClassTrendPoint{}
	}
	return trend.Points
}

func safeReviewItems(review *teachingquerycontracts.TeacherClassReview) []teachingquerycontracts.TeacherClassReviewItem {
	if review == nil {
		return []teachingquerycontracts.TeacherClassReviewItem{}
	}
	return review.Items
}

func reviewStudentNames(students []teachingquerycontracts.TeacherReviewStudentRef) string {
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

func reviewSummaryActiveRate(summary *teachingquerycontracts.TeacherClassSummary) float64 {
	if summary == nil {
		return 0
	}
	return summary.ActiveRate
}

func reviewSummaryRecentEvents(summary *teachingquerycontracts.TeacherClassSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.RecentEventCount
}
