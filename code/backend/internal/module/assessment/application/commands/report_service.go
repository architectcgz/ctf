package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	assessmentconfig "ctf-platform/internal/module/assessment/config"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
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
	outputStore       assessmentports.ReportOutputStore
	awdReviewBuilder  AWDReviewExportBuilder
	config            config.ReportConfig
	logger            *zap.Logger
	workerPool        chan struct{}
	baseCtx           context.Context
	cancel            context.CancelFunc
	tasks             sync.WaitGroup
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
	outputStore assessmentports.ReportOutputStore,
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
		outputStore:       outputStore,
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
		filePath, expiresAt, err := s.generateTeacherAWDReviewArchive(runCtx, report.ID, archive)
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
		filePath, expiresAt, err := s.generateTeacherAWDReviewReport(runCtx, report.ID, archive)
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

	if s.outputStore == nil {
		return nil, apperror.ErrServiceUnavailable.WithMessage("报告输出存储暂不可用")
	}
	downloadStream, err := s.outputStore.OpenReportDownload(ctx, report.FilePath)
	if err != nil {
		if errors.Is(err, assessmentports.ErrAssessmentReportOutputUnsafePath) {
			return nil, apperror.ErrForbidden
		}
		if errors.Is(err, assessmentports.ErrAssessmentReportOutputNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
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
		Reader:      downloadStream.Reader,
		Size:        downloadStream.Size,
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
