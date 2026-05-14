package queries

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	queryports "ctf-platform/internal/module/teaching_query/ports"
	"ctf-platform/internal/teaching/classreview"
	"ctf-platform/internal/teaching/classwindow"
	"ctf-platform/pkg/errcode"
)

type classInsightQueryRepository interface {
	queryports.TeachingClassInsightRepository
}

type ClassInsightQueryService struct {
	users                 queryports.TeachingUserLookupRepository
	repo                  classInsightQueryRepository
	recommendationService assessmentcontracts.RecommendationProvider
	logger                *zap.Logger
}

var _ ClassInsightService = (*ClassInsightQueryService)(nil)

func NewClassInsightService(
	users queryports.TeachingUserLookupRepository,
	repo classInsightQueryRepository,
	recommendationService assessmentcontracts.RecommendationProvider,
	logger *zap.Logger,
) *ClassInsightQueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ClassInsightQueryService{
		users:                 users,
		repo:                  repo,
		recommendationService: recommendationService,
		logger:                logger,
	}
}

func (s *ClassInsightQueryService) GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherClassInsightQuery) (*dto.TeacherClassSummaryResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.users, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	window, err := s.parseWindow(query)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.GetClassSummary(ctx, normalized, window.Since)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return teachingQueryMapper.ToClassSummaryPtr(summary), nil
}

func (s *ClassInsightQueryService) GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherClassInsightQuery) (*dto.TeacherClassTrendResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.users, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	window, err := s.parseWindow(query)
	if err != nil {
		return nil, err
	}

	trend, err := s.repo.GetClassTrend(ctx, normalized, window.StartOfDay, window.Days)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return teachingQueryMapper.ToClassTrendRespPtr(trend), nil
}

func (s *ClassInsightQueryService) GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherClassInsightQuery) (*dto.TeacherClassReviewResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := ensureClassAccess(ctx, s.users, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	window, err := s.parseWindow(query)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.GetClassSummary(ctx, normalized, window.Since)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	trend, err := s.repo.GetClassTrend(ctx, normalized, window.StartOfDay, window.Days)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	snapshots, err := s.repo.ListClassTeachingFactSnapshots(ctx, normalized, window.Since)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
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

	return classreview.BuildResponse(ctx, classreview.Input{
		ClassName:        normalized,
		ActiveRate:       summary.ActiveRate,
		RecentEventCount: summary.RecentEventCount,
		HasTrend:         hasTrend,
		TrendEventDelta:  trendEventDelta,
		TrendSolveDelta:  trendSolveDelta,
		Snapshots:        snapshots,
	}, classreview.RecommendationResolverFunc(s.matchingStudentRecommendation)), nil
}

func (s *ClassInsightQueryService) parseWindow(query *dto.TeacherClassInsightQuery) (classwindow.Range, error) {
	if query == nil {
		return classwindow.Parse(queryNow(), "", "")
	}
	window, err := classwindow.Parse(queryNow(), query.FromDate, query.ToDate)
	if err != nil {
		return classwindow.Range{}, errcode.New(errcode.ErrInvalidParams.Code, err.Error(), errcode.ErrInvalidParams.HTTPStatus)
	}
	return window, nil
}

func (s *ClassInsightQueryService) matchingStudentRecommendation(
	ctx context.Context,
	studentIDs []int64,
	dimension string,
	limit int,
) *dto.TeacherRecommendationItem {
	if s.recommendationService == nil {
		return nil
	}
	targetDimension := strings.ToLower(strings.TrimSpace(dimension))
	if targetDimension == "" {
		return nil
	}
	for _, studentID := range studentIDs {
		result, err := s.recommendationService.Recommend(ctx, studentID, limit)
		if err != nil {
			s.logger.Warn("recommend_student_for_class_review_failed", zap.Int64("student_id", studentID), zap.Error(err))
			continue
		}
		if result == nil || len(result.Challenges) == 0 {
			continue
		}
		for _, challenge := range result.Challenges {
			if challenge == nil || !recommendationMatchesDimension(challenge, targetDimension) {
				continue
			}
			return teachingQueryMapper.ToTeacherRecommendationItemPtr(challenge)
		}
	}
	return nil
}

var queryNow = func() time.Time {
	return time.Now().UTC()
}

func recommendationMatchesDimension(challenge *dto.ChallengeRecommendation, dimension string) bool {
	if challenge == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(challenge.Dimension), dimension) {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(challenge.Category), dimension)
}
