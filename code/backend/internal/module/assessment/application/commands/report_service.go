package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	"ctf-platform/internal/shared/mapperutil"
	"ctf-platform/internal/teaching/evidence"
	"ctf-platform/pkg/errcode"
)

type ReportService struct {
	lifecycleRepo     assessmentports.AssessmentReportLifecycleRepository
	userRepo          assessmentports.AssessmentReportUserLookupRepository
	contestRepo       assessmentports.AssessmentReportContestLookupRepository
	personalRepo      assessmentports.AssessmentPersonalReportRepository
	classRepo         assessmentports.AssessmentClassReportRepository
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
	SkillProfile   []*dto.SkillDimension
	Stats          *assessmentdomain.PersonalReportStats
	DimensionStats []assessmentdomain.ReportDimensionStat
}

type classReportData struct {
	ClassName         string
	TotalStudents     int
	AverageScore      float64
	DimensionAverages []assessmentdomain.ClassDimensionAverage
	TopStudents       []assessmentdomain.ClassTopStudent
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
	SkillProfile        []*dto.SkillDimension                             `json:"skill_profile,omitempty"`
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

func NewReportService(
	lifecycleRepo assessmentports.AssessmentReportLifecycleRepository,
	userRepo assessmentports.AssessmentReportUserLookupRepository,
	contestRepo assessmentports.AssessmentReportContestLookupRepository,
	personalRepo assessmentports.AssessmentPersonalReportRepository,
	classRepo assessmentports.AssessmentClassReportRepository,
	contestExportRepo assessmentports.AssessmentContestExportRepository,
	reviewArchiveRepo assessmentports.AssessmentReviewArchiveRepository,
	assessmentService assessmentports.AssessmentProfileReader,
	cfg config.ReportConfig,
	logger *zap.Logger,
) *ReportService {
	if logger == nil {
		logger = zap.NewNop()
	}

	cfg = assessmentdomain.NormalizeReportConfig(cfg)
	return &ReportService{
		lifecycleRepo:     lifecycleRepo,
		userRepo:          userRepo,
		contestRepo:       contestRepo,
		personalRepo:      personalRepo,
		classRepo:         classRepo,
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

func (s *ReportService) CreatePersonalReport(ctx context.Context, userID int64, req CreatePersonalReportInput) (*dto.ReportExportData, error) {
	if ctx == nil {
		return nil, errors.New("create personal report requires context")
	}

	format := s.normalizeFormat(req.Format)
	report := &model.Report{
		Type:   model.ReportTypePersonal,
		Format: format,
		UserID: &userID,
		Status: model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	reportCtx, cancel := s.withPersonalTimeout(ctx)
	defer cancel()

	filePath, expiresAt, err := s.generatePersonalReport(reportCtx, report.ID, userID, format)
	if err != nil {
		s.markFailed(reportCtx, report.ID, err)
		return nil, err
	}
	if err := s.lifecycleRepo.MarkReady(reportCtx, report.ID, filePath, expiresAt); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return buildReportExportData(report.ID, model.ReportStatusReady, expiresAt), nil
}

func (s *ReportService) withPersonalTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.config.PersonalTimeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, s.config.PersonalTimeout)
}

func (s *ReportService) CreateClassReport(ctx context.Context, requesterID int64, req CreateClassReportInput) (*dto.ReportExportData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}

	className := strings.TrimSpace(req.ClassName)
	if className == "" {
		className = strings.TrimSpace(requester.ClassName)
	}
	if className == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := validateClassReportAccess(requester, className); err != nil {
		return nil, err
	}

	format := s.normalizeFormat(req.Format)
	report := &model.Report{
		Type:      model.ReportTypeClass,
		Format:    format,
		UserID:    &requesterID,
		ClassName: &className,
		Status:    model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateClassReport(runCtx, report.ID, className, format)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, model.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateContestExport(ctx context.Context, requesterID, contestID int64, req CreateContestExportInput) (*dto.ReportExportData, error) {
	if _, err := s.contestRepo.FindContestByID(ctx, contestID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	format := s.normalizeArchiveFormat(req.Format)
	report := &model.Report{
		Type:   model.ReportTypeContest,
		Format: format,
		UserID: &requesterID,
		Status: model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateContestExport(runCtx, report.ID, contestID, format)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, model.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateStudentReviewArchive(ctx context.Context, requesterID, studentID int64, req CreateStudentReviewArchiveInput) (*dto.ReportExportData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	if err := validateStudentReviewArchiveAccess(requester, student); err != nil {
		return nil, err
	}

	format := s.normalizeArchiveFormat(req.Format)
	report := &model.Report{
		Type:      model.ReportTypeReview,
		Format:    format,
		UserID:    &requesterID,
		ClassName: &student.ClassName,
		Status:    model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.runAsyncReport(report.ID, func(runCtx context.Context) error {
		filePath, expiresAt, genErr := s.generateStudentReviewArchive(runCtx, report.ID, studentID, format)
		if genErr != nil {
			return genErr
		}
		return s.lifecycleRepo.MarkReady(runCtx, report.ID, filePath, expiresAt)
	})

	return buildReportExportData(report.ID, model.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateTeacherAWDReviewArchive(ctx context.Context, requesterID, contestID int64, req CreateTeacherAWDReviewExportInput) (*dto.ReportExportData, error) {
	if _, err := s.findAWDContestForExport(ctx, contestID); err != nil {
		return nil, err
	}
	if s.awdReviewBuilder == nil {
		return nil, errcode.New(errcode.ErrServiceUnavailable.Code, "教师 AWD 复盘归档导出暂不可用", errcode.ErrServiceUnavailable.HTTPStatus)
	}

	report := &model.Report{
		Type:   model.ReportTypeAWDReviewArchive,
		Format: model.ReportFormatZIP,
		UserID: &requesterID,
		Status: model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
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

	return buildReportExportData(report.ID, model.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) CreateTeacherAWDReviewReport(ctx context.Context, requesterID, contestID int64, req CreateTeacherAWDReviewExportInput) (*dto.ReportExportData, error) {
	contest, err := s.findAWDContestForExport(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if contest.Status != model.ContestStatusEnded {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "教师复盘报告仅支持赛后导出", errcode.ErrInvalidParams.HTTPStatus)
	}
	if s.awdReviewBuilder == nil {
		return nil, errcode.New(errcode.ErrServiceUnavailable.Code, "教师 AWD 复盘报告导出暂不可用", errcode.ErrServiceUnavailable.HTTPStatus)
	}

	report := &model.Report{
		Type:   model.ReportTypeAWDReviewReport,
		Format: model.ReportFormatPDF,
		UserID: &requesterID,
		Status: model.ReportStatusProcessing,
	}
	if err := s.lifecycleRepo.Create(ctx, report); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
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

	return buildReportExportData(report.ID, model.ReportStatusProcessing, time.Time{}), nil
}

func (s *ReportService) GetStudentReviewArchive(ctx context.Context, requesterID, studentID int64) (*ReviewArchiveData, error) {
	requester, err := s.userRepo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}
	student, err := s.userRepo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	if err := validateStudentReviewArchiveAccess(requester, student); err != nil {
		return nil, err
	}
	return s.buildStudentReviewArchiveData(ctx, studentID)
}

func validateClassReportAccess(requester *assessmentdomain.ReportUser, className string) error {
	if requester == nil || requester.ID <= 0 {
		return errcode.ErrUnauthorized
	}
	if requester.Role == model.RoleAdmin {
		return nil
	}
	if strings.TrimSpace(requester.ClassName) == "" || strings.TrimSpace(requester.ClassName) != className {
		return errcode.ErrForbidden
	}
	return nil
}

func validateStudentReviewArchiveAccess(requester, student *assessmentdomain.ReportUser) error {
	if requester == nil || requester.ID <= 0 {
		return errcode.ErrUnauthorized
	}
	if student == nil || student.ID <= 0 {
		return errcode.ErrNotFound
	}
	if student.Role != model.RoleStudent {
		return errcode.New(errcode.ErrInvalidParams.Code, "目标用户不是学生", errcode.ErrInvalidParams.HTTPStatus)
	}
	if requester.Role == model.RoleAdmin {
		return nil
	}
	if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != student.ClassName {
		return errcode.ErrForbidden
	}
	return nil
}

func (s *ReportService) findAWDContestForExport(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrContestNotFound
	}
	return contest, nil
}

func (s *ReportService) GetDownload(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error) {
	report, err := s.lifecycleRepo.FindByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := assessmentdomain.ValidateReportAccess(report, requesterID, role); err != nil {
		return nil, err
	}
	if report.Status == model.ReportStatusProcessing {
		return nil, errcode.New(errcode.ErrConflict.Code, "报告仍在生成中", errcode.ErrConflict.HTTPStatus)
	}
	if report.Status == model.ReportStatusFailed {
		message := "报告生成失败"
		if report.ErrorMsg != nil && strings.TrimSpace(*report.ErrorMsg) != "" {
			message = *report.ErrorMsg
		}
		return nil, errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}

	filePath, err := s.safeReportPath(report.FilePath)
	if err != nil {
		return nil, errcode.ErrForbidden
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(statErr)
	}

	fileName := reportDownloadFileName(report)
	format := reportOutputFormat(report)
	contentType := reportContentType(format)
	if contentType == "" {
		switch format {
		case model.ReportFormatJSON:
			contentType = "application/json"
		case model.ReportFormatPDF:
			contentType = "application/pdf"
		case model.ReportFormatExcel:
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case model.ReportFormatZIP:
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

func (s *ReportService) GetStatus(ctx context.Context, reportID, requesterID int64, role string) (*dto.ReportExportData, error) {
	report, err := s.lifecycleRepo.FindByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
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

	filePath, err := s.reportFilePath(reportID, model.ReportTypePersonal, format)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateClassReport(ctx context.Context, reportID int64, className, format string) (string, time.Time, error) {
	data, err := s.buildClassReportData(ctx, className)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, model.ReportTypeClass, format)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateContestExport(ctx context.Context, reportID, contestID int64, format string) (string, time.Time, error) {
	data, err := s.buildContestExportData(ctx, contestID)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, model.ReportTypeContest, format)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateStudentReviewArchive(ctx context.Context, reportID, studentID int64, format string) (string, time.Time, error) {
	data, err := s.buildStudentReviewArchiveData(ctx, studentID)
	if err != nil {
		return "", time.Time{}, err
	}

	filePath, err := s.reportFilePath(reportID, model.ReportTypeReview, format)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := s.renderReport(filePath, format, data); err != nil {
		return "", time.Time{}, err
	}

	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewArchive(reportID int64, archive *dto.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	filePath, err := s.reportFilePath(reportID, model.ReportTypeAWDReviewArchive, model.ReportFormatZIP)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := RenderAWDReviewArchiveZip(filePath, archive); err != nil {
		return "", time.Time{}, err
	}
	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) generateTeacherAWDReviewReport(reportID int64, archive *dto.TeacherAWDReviewArchiveResp) (string, time.Time, error) {
	filePath, err := s.reportFilePath(reportID, model.ReportTypeAWDReviewReport, model.ReportFormatPDF)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}
	if err := RenderAWDReviewReportPDF(filePath, archive); err != nil {
		return "", time.Time{}, err
	}
	return filePath, time.Now().Add(s.config.FileTTL), nil
}

func (s *ReportService) buildPersonalReportData(ctx context.Context, userID int64) (*personalReportData, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}

	skillProfileResp, err := s.assessmentService.GetSkillProfile(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	stats, err := s.personalRepo.GetPersonalStats(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	dimensionStats, err := s.personalRepo.ListPersonalDimensionStats(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &personalReportData{
		User:           user,
		SkillProfile:   skillProfileResp.Dimensions,
		Stats:          stats,
		DimensionStats: dimensionStats,
	}, nil
}

func (s *ReportService) buildClassReportData(ctx context.Context, className string) (*classReportData, error) {
	totalStudents, err := s.classRepo.CountClassStudents(ctx, className)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	avgScore, err := s.classRepo.GetClassAverageScore(ctx, className)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	dimensionAverages, err := s.classRepo.ListClassDimensionAverages(ctx, className)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	topStudents, err := s.classRepo.ListClassTopStudents(ctx, className, 10)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &classReportData{
		ClassName:         className,
		TotalStudents:     totalStudents,
		AverageScore:      avgScore,
		DimensionAverages: assessmentdomain.FillMissingDimensionAverages(dimensionAverages),
		TopStudents:       topStudents,
	}, nil
}

func (s *ReportService) buildContestExportData(ctx context.Context, contestID int64) (*contestExportData, error) {
	contest, err := s.contestRepo.FindContestByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	scoreboard, err := s.contestExportRepo.ListContestScoreboard(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	challenges, err := s.contestExportRepo.ListContestChallenges(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	teams, err := s.contestExportRepo.ListContestTeams(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &contestExportData{
		GeneratedAt: time.Now(),
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
		return nil, errcode.ErrNotFound
	}

	stats, err := s.personalRepo.GetPersonalStats(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	totalChallenges, err := s.reviewArchiveRepo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	timeline, err := s.reviewArchiveRepo.GetStudentTimeline(ctx, studentID, 200, 0)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	evidence, err := s.reviewArchiveRepo.GetStudentEvidence(ctx, studentID, evidence.Query{})
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	writeups, err := s.reviewArchiveRepo.ListStudentWriteups(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	manualReviews, err := s.reviewArchiveRepo.ListStudentManualReviews(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var skillProfile []*dto.SkillDimension
	if s.assessmentService != nil {
		skillProfileResp, skillErr := s.assessmentService.GetSkillProfile(ctx, studentID)
		if skillErr != nil {
			return nil, errcode.ErrInternal.WithCause(skillErr)
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
		TeacherObservations: buildReviewArchiveObservations(summary, evidence, writeups, manualReviews),
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
	count := 0
	for _, item := range timeline {
		if isCorrectTimelineSubmission(item) {
			count++
		}
	}
	if count > 0 {
		return count
	}
	for _, item := range evidence {
		if isCorrectEvidenceSubmission(item) {
			count++
		}
	}
	return count
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
	evidence []assessmentdomain.ReviewArchiveEvidenceEvent,
	writeups []assessmentdomain.ReviewArchiveWriteupItem,
	manualReviews []assessmentdomain.ReviewArchiveManualReviewItem,
) assessmentdomain.ReviewArchiveTeacherObservations {
	items := make([]assessmentdomain.ReviewArchiveObservation, 0, 4)

	if summary.CorrectSubmissionCount > 0 && (hasSubmittedWriteup(writeups) || hasApprovedManualReview(manualReviews)) {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Key:      "training_closure",
			Label:    "训练闭环",
			Level:    "good",
			Summary:  "已形成从利用到复盘输出的有效闭环。",
			Evidence: fmt.Sprintf("命中正确提交 %d 次，复盘材料 %d 份，人工审核记录 %d 条。", summary.CorrectSubmissionCount, summary.WriteupCount, summary.ManualReviewCount),
		})
	} else if summary.CorrectSubmissionCount > 0 {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Key:      "training_closure",
			Label:    "训练闭环",
			Level:    "attention",
			Summary:  "已完成解题，但复盘输出仍偏弱。",
			Evidence: fmt.Sprintf("命中正确提交 %d 次，但当前复盘材料 %d 份。", summary.CorrectSubmissionCount, summary.WriteupCount),
		})
	}

	items = append(items, assessmentdomain.ReviewArchiveObservation{
		Key:      "hint_usage",
		Label:    "提示依赖",
		Level:    "good",
		Summary:  "当前归档不再统计提示解锁行为，训练记录聚焦实操与提交表现。",
		Evidence: "未纳入提示解锁事件。",
	})

	if hasRepeatedWrongSubmissions(evidence) {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Key:      "submission_stability",
			Label:    "提交稳定性",
			Level:    "attention",
			Summary:  "存在连续错误提交，课堂复盘可重点回看试错节奏。",
			Evidence: "证据链中存在至少 2 次连续未命中提交。",
		})
	}

	if hasHandsOnExploit(evidence) {
		items = append(items, assessmentdomain.ReviewArchiveObservation{
			Key:      "hands_on_activity",
			Label:    "实操参与",
			Level:    "good",
			Summary:  "实操交互记录充分，具备课堂演示价值。",
			Evidence: "证据链中包含实例访问、平台代理请求或 AWD 攻击日志。",
		})
	}

	return assessmentdomain.ReviewArchiveTeacherObservations{Items: items}
}

func hasSubmittedWriteup(writeups []assessmentdomain.ReviewArchiveWriteupItem) bool {
	for _, item := range writeups {
		if item.SubmissionStatus == "published" || item.SubmissionStatus == "submitted" {
			return true
		}
	}
	return false
}

func hasApprovedManualReview(items []assessmentdomain.ReviewArchiveManualReviewItem) bool {
	for _, item := range items {
		if item.ReviewStatus == "approved" {
			return true
		}
	}
	return false
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
		if item.Type == "instance_access" || item.Type == "instance_proxy_request" || item.Type == "awd_attack_submission" {
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
	case "challenge_submission":
		isCorrect, ok := item.Meta["is_correct"].(bool)
		return isCorrect, ok
	case "awd_attack_submission":
		isCorrect, ok := item.Meta["is_success"].(bool)
		return isCorrect, ok
	default:
		return false, false
	}
}

func (s *ReportService) renderReport(filePath, format string, data any) error {
	switch format {
	case model.ReportFormatJSON:
		return writeJSONReport(filePath, data)
	case model.ReportFormatExcel:
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
	return errcode.ErrInternal.WithCause(fmt.Errorf("unsupported report payload"))
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
	case model.ReportFormatJSON:
		return model.ReportFormatJSON
	case model.ReportFormatExcel:
		return model.ReportFormatExcel
	case model.ReportFormatPDF:
		return model.ReportFormatPDF
	default:
		return s.config.DefaultFormat
	}
}

func (s *ReportService) normalizeArchiveFormat(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), model.ReportFormatJSON) {
		return model.ReportFormatJSON
	}
	return model.ReportFormatJSON
}

func reportFileExtension(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), model.ReportFormatJSON) {
		return "json"
	}
	if strings.EqualFold(strings.TrimSpace(format), model.ReportFormatZIP) {
		return "zip"
	}
	if strings.EqualFold(strings.TrimSpace(format), model.ReportFormatExcel) {
		return "xlsx"
	}
	return "pdf"
}

func reportOutputFormat(report *model.Report) string {
	if report == nil {
		return model.ReportFormatPDF
	}
	switch report.Type {
	case model.ReportTypeAWDReviewArchive:
		return model.ReportFormatZIP
	case model.ReportTypeAWDReviewReport:
		return model.ReportFormatPDF
	default:
		return report.Format
	}
}

func reportDownloadFileName(report *model.Report) string {
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

func buildReportExportData(reportID int64, status string, expiresAt time.Time) *dto.ReportExportData {
	report := &model.Report{
		ID:        reportID,
		Status:    status,
		ExpiresAt: nil,
	}
	if !expiresAt.IsZero() {
		report.ExpiresAt = &expiresAt
	}
	return buildReportExportDataFromModel(report)
}

func buildReportExportDataFromModel(report *model.Report) *dto.ReportExportData {
	resp := assessmentCommandResponseMapperInst.ToReportExportDataBasePtr(report)
	if report.Status == model.ReportStatusReady {
		downloadURL := fmt.Sprintf("/api/v1/reports/%d/download", report.ID)
		resp.DownloadURL = &downloadURL
		if report.ExpiresAt != nil && !report.ExpiresAt.IsZero() {
			expires := report.ExpiresAt.Format(time.RFC3339)
			resp.ExpiresAt = &expires
		}
	}
	if report.Status == model.ReportStatusFailed {
		resp.ErrorMessage = mapperutil.NormalizeOptionalString(report.ErrorMsg)
	}
	return resp
}

func writePersonalPDF(filePath string, data *personalReportData) error {
	pdf := newReportPDF()
	addReportTitle(pdf, "Personal Training Report")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "Username", Value: sanitizePDFText(data.User.Username)},
		{Label: "Class", Value: sanitizePDFText(data.User.ClassName)},
		{Label: "Total Score", Value: fmt.Sprintf("%d", data.Stats.TotalScore)},
		{Label: "Rank", Value: fmt.Sprintf("%d", data.Stats.Rank)},
		{Label: "Solved", Value: fmt.Sprintf("%d", data.Stats.TotalSolved)},
		{Label: "Attempts", Value: fmt.Sprintf("%d", data.Stats.TotalAttempts)},
	})
	addDimensionChart(pdf, "Skill Profile", skillProfileChartRows(data.SkillProfile))
	addDimensionStatsTable(pdf, "Dimension Details", data.DimensionStats)
	return pdf.OutputFileAndClose(filePath)
}

func writeClassPDF(filePath string, data *classReportData) error {
	pdf := newReportPDF()
	addReportTitle(pdf, "Class Training Report")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "Class", Value: sanitizePDFText(data.ClassName)},
		{Label: "Total Students", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "Average Score", Value: fmt.Sprintf("%.2f", data.AverageScore)},
	})
	addAverageChart(pdf, "Dimension Average", data.DimensionAverages)
	addTopStudentsTable(pdf, "Top Students", data.TopStudents)
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
	topSheet := "TopStudents"
	file.NewSheet(topSheet)

	headerStyle := mustNewExcelStyle(file, &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FCE4D6"}},
	})

	writePairs(file, summarySheet, []summaryLine{
		{Label: "Class", Value: data.ClassName},
		{Label: "Total Students", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "Average Score", Value: fmt.Sprintf("%.2f", data.AverageScore)},
	}, headerStyle)

	file.SetCellValue(summarySheet, "A7", "Dimension")
	file.SetCellValue(summarySheet, "B7", "Average Score")
	file.SetCellStyle(summarySheet, "A7", "B7", headerStyle)
	for idx, dimension := range data.DimensionAverages {
		row := idx + 8
		file.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), dimension.Dimension)
		file.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), dimension.AvgScore)
	}

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
				Name:       fmt.Sprintf("%s!$B$7", summarySheet),
				Categories: fmt.Sprintf("%s!$A$8:$A$%d", summarySheet, len(data.DimensionAverages)+7),
				Values:     fmt.Sprintf("%s!$B$8:$B$%d", summarySheet, len(data.DimensionAverages)+7),
			}},
			Title:  []excelize.RichTextRun{{Text: "Dimension Average"}},
			Legend: excelize.ChartLegend{Position: "bottom"},
		})
	}

	setReportSheetLayout(file, summarySheet)
	setReportSheetLayout(file, topSheet)
	return file.SaveAs(filePath)
}

func writeJSONReport(filePath string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

type summaryLine struct {
	Label string
	Value string
}

func newReportPDF() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)
	return pdf
}

func addReportTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 12, sanitizePDFText(title), "", 1, "L", false, 0, "")
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(16, pdf.GetY(), 194, pdf.GetY())
	pdf.Ln(6)
}

func addSummaryBlock(pdf *gofpdf.Fpdf, lines []summaryLine) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "Summary", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	for _, line := range lines {
		pdf.CellFormat(45, 7, sanitizePDFText(line.Label), "0", 0, "L", false, 0, "")
		pdf.CellFormat(0, 7, sanitizePDFText(line.Value), "0", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addDimensionChart(pdf *gofpdf.Fpdf, title string, rows []chartRow) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	for _, row := range rows {
		ensurePDFSpace(pdf, 12)
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(28, 7, sanitizePDFText(row.Label), "", 0, "L", false, 0, "")
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
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"Dimension", "Solved", "Total"})
	pdf.SetFont("Helvetica", "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(70, 7, sanitizePDFText(row.Dimension), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Solved), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Total), "1", 1, "C", false, 0, "")
	}
}

func addTopStudentsTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassTopStudent) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"Rank", "Username", "Score"})
	pdf.SetFont("Helvetica", "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(90, 7, sanitizePDFText(row.Username), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.TotalScore), "1", 1, "C", false, 0, "")
	}
}

func writePDFTableHeader(pdf *gofpdf.Fpdf, headers []string) {
	pdf.SetFillColor(220, 230, 241)
	pdf.SetFont("Helvetica", "B", 10)
	widths := []float64{70, 40, 40}
	if len(headers) == 3 && headers[0] == "Rank" {
		widths = []float64{30, 90, 30}
	}
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

type chartRow struct {
	Label string
	Value float64
}

func skillProfileChartRows(dimensions []*dto.SkillDimension) []chartRow {
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
	buf := strings.Builder{}
	for _, r := range value {
		if r >= 32 && r <= 126 {
			buf.WriteRune(r)
			continue
		}
		buf.WriteRune('?')
	}
	return buf.String()
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
