package queries

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
	teachingadvice "ctf-platform/internal/teaching/advice"
	"ctf-platform/pkg/errcode"
)

type classInsightQueryRepository interface {
	classAccessRepository
	readmodelports.TeachingClassInsightRepository
}

type ClassInsightQueryService struct {
	repo                  classInsightQueryRepository
	recommendationService assessmentcontracts.RecommendationProvider
	logger                *zap.Logger
}

var _ ClassInsightService = (*ClassInsightQueryService)(nil)

func NewClassInsightService(
	repo classInsightQueryRepository,
	recommendationService assessmentcontracts.RecommendationProvider,
	logger *zap.Logger,
) *ClassInsightQueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ClassInsightQueryService{
		repo:                  repo,
		recommendationService: recommendationService,
		logger:                logger,
	}
}

func (s *ClassInsightQueryService) GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassSummaryResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.repo, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return teachingReadmodelMapper.ToClassSummaryPtr(summary), nil
}

func (s *ClassInsightQueryService) GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassTrendResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.repo, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return teachingReadmodelMapper.ToClassTrendRespPtr(trend), nil
}

func (s *ClassInsightQueryService) GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassReviewResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.repo, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())

	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	snapshots, err := s.repo.ListClassTeachingFactSnapshots(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	summaryDTO := teachingReadmodelMapper.ToClassSummaryPtr(summary)
	classTrend := buildClassTrendSnapshot(trend)
	evaluations := make(map[int64]teachingadvice.StudentEvaluation, len(snapshots))
	studentRefs := make(map[int64]dto.TeacherReviewStudentRef, len(snapshots))
	for _, snapshot := range snapshots {
		evaluations[snapshot.UserID] = teachingadvice.EvaluateStudent(snapshot)
		studentRefs[snapshot.UserID] = dto.TeacherReviewStudentRef{
			ID:       snapshot.UserID,
			Username: snapshot.Username,
			Name:     snapshot.Name,
		}
	}

	adviceItems := teachingadvice.BuildClassReview(
		normalized,
		teachingadvice.ClassSummarySnapshot{
			ClassName:        normalized,
			StudentCount:     len(snapshots),
			ActiveRate:       summaryDTO.ActiveRate,
			RecentEventCount: summaryDTO.RecentEventCount,
		},
		classTrend,
		snapshots,
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
		if adviceItem.RecommendationStudentID != nil {
			if recommendation := s.firstStudentRecommendation(ctx, []int64{*adviceItem.RecommendationStudentID}, 1); recommendation != nil {
				item.Recommendation = recommendation
			}
		}
		items = append(items, item)
	}

	return &dto.TeacherClassReviewResp{
		ClassName: normalized,
		Items:     items,
	}, nil
}

func buildClassTrendSnapshot(trend *readmodelports.ClassTrend) *teachingadvice.ClassTrendSnapshot {
	if trend == nil || len(trend.Points) < 2 {
		return nil
	}
	first := trend.Points[0]
	last := trend.Points[len(trend.Points)-1]
	return &teachingadvice.ClassTrendSnapshot{
		EventDelta: last.EventCount - first.EventCount,
		SolveDelta: last.SolveCount - first.SolveCount,
	}
}

func (s *ClassInsightQueryService) firstStudentRecommendation(ctx context.Context, studentIDs []int64, limit int) *dto.TeacherRecommendationItem {
	if s.recommendationService == nil {
		return nil
	}
	for _, studentID := range studentIDs {
		result, err := s.recommendationService.Recommend(ctx, studentID, limit)
		if err != nil {
			s.logger.Warn("recommend_student_for_class_review_failed", zap.Int64("student_id", studentID), zap.Error(err))
			continue
		}
		if result == nil || len(result.Challenges) == 0 || result.Challenges[0] == nil {
			continue
		}
		return teachingReadmodelMapper.ToTeacherRecommendationItemPtr(result.Challenges[0])
	}
	return nil
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
