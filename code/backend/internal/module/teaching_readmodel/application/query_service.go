package application

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
	"ctf-platform/pkg/errcode"
)

type RecommendationProvider interface {
	Recommend(userID int64, limit int) (*dto.RecommendationResp, error)
	RecommendWithContext(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error)
}

type QueryService struct {
	repo                  *readmodelinfra.Repository
	recommendationService RecommendationProvider
	logger                *zap.Logger
}

var _ teachingreadmodel.TeachingQuery = (*QueryService)(nil)

func NewQueryService(repo *readmodelinfra.Repository, recommendationService RecommendationProvider, logger *zap.Logger) *QueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QueryService{
		repo:                  repo,
		recommendationService: recommendationService,
		logger:                logger,
	}
}

func (s *QueryService) ListClasses(ctx context.Context, requesterID int64, requesterRole string) ([]dto.TeacherClassItem, error) {
	requester, err := s.repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return nil, errcode.ErrUnauthorized
	}

	if requesterRole == model.RoleAdmin {
		items, err := s.repo.ListClasses(ctx)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return items, nil
	}

	className := strings.TrimSpace(requester.ClassName)
	if className == "" {
		return []dto.TeacherClassItem{}, nil
	}

	count, err := s.repo.CountStudentsByClass(ctx, className)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return []dto.TeacherClassItem{{
		Name:         className,
		StudentCount: count,
	}}, nil
}

func (s *QueryService) ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string, query *dto.TeacherStudentQuery) ([]dto.TeacherStudentItem, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}

	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	studentNo := ""
	keyword := ""
	if query != nil {
		studentNo = strings.TrimSpace(query.StudentNo)
		keyword = strings.TrimSpace(query.Keyword)
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	items, err := s.repo.ListStudentsByClass(ctx, normalized, keyword, studentNo, startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return items, nil
}

func (s *QueryService) GetClassSummary(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassSummaryResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return summary, nil
}

func (s *QueryService) GetClassTrend(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassTrendResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return trend, nil
}

func (s *QueryService) GetClassReview(ctx context.Context, requesterID int64, requesterRole, className string) (*dto.TeacherClassReviewResp, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}
	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	since := time.Now().AddDate(0, 0, -6)
	startOfDay := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())

	students, err := s.repo.ListStudentsByClass(ctx, normalized, "", "", startOfDay)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	summary, err := s.repo.GetClassSummary(ctx, normalized, time.Now().AddDate(0, 0, -7))
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	trend, err := s.repo.GetClassTrend(ctx, normalized, startOfDay, 7)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.TeacherClassReviewItem, 0, 4)
	riskStudents := selectRiskStudents(students, 3)
	activeRate := summary.ActiveRate

	switch {
	case activeRate < 50:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级活跃度偏低",
			Detail: fmt.Sprintf("%s 近 7 天活跃率只有 %.0f%%，建议优先跟进 %d 名低活跃学生。", normalized, activeRate, len(riskStudents)),
			Accent: "danger",
		})
	case activeRate < 75:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级活跃度需要补强",
			Detail: fmt.Sprintf("%s 近 7 天活跃率为 %.0f%%，适合通过定向训练把低活跃学生重新拉回训练节奏。", normalized, activeRate),
			Accent: "warning",
		})
	default:
		items = append(items, dto.TeacherClassReviewItem{
			Key:    "activity",
			Title:  "班级训练节奏整体稳定",
			Detail: fmt.Sprintf("%s 近 7 天活跃率达到 %.0f%%，当前更适合做薄弱维度补强而不是全面催学。", normalized, activeRate),
			Accent: "success",
		})
	}

	if weakDimension, weakStudents := selectWeakDimensionStudents(students); weakDimension != "" {
		item := dto.TeacherClassReviewItem{
			Key:      "weak_dimension",
			Title:    "优先补薄弱维度",
			Detail:   fmt.Sprintf("%s 是当前最集中的薄弱项，涉及 %d 名学生，建议本周统一布置该维度基础题。", weakDimension, len(weakStudents)),
			Accent:   "primary",
			Students: toReviewStudentRefs(limitStudents(weakStudents, 3)),
		}
		if len(weakStudents) >= 3 {
			item.Accent = "warning"
		}
		if recommendation := s.firstStudentRecommendation(ctx, weakStudents, 1); recommendation != nil {
			item.Recommendation = recommendation
		}
		items = append(items, item)
	}

	if len(riskStudents) > 0 {
		item := dto.TeacherClassReviewItem{
			Key:      "focus_students",
			Title:    "先跟进重点学生",
			Detail:   fmt.Sprintf("建议教师先跟进 %s，并优先布置推荐题做补强训练。", joinStudentNames(riskStudents)),
			Accent:   "primary",
			Students: toReviewStudentRefs(riskStudents),
		}
		if recommendation := s.firstStudentRecommendation(ctx, riskStudents, 1); recommendation != nil {
			item.Recommendation = recommendation
		}
		items = append(items, item)
	}

	if trendItem := buildTrendReviewItem(trend); trendItem != nil {
		items = append(items, *trendItem)
	}

	return &dto.TeacherClassReviewResp{
		ClassName: normalized,
		Items:     items,
	}, nil
}

func (s *QueryService) GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.TeacherProgressResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	totalChallenges, err := s.repo.CountPublishedChallenges(ctx)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	solvedChallenges, err := s.repo.CountSolvedChallenges(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	categoryRows, err := s.repo.GetCategoryProgress(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	difficultyRows, err := s.repo.GetDifficultyProgress(ctx, student.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TeacherProgressResp{
		TotalChallenges:  int(totalChallenges),
		SolvedChallenges: int(solvedChallenges),
		ByCategory:       toProgressBreakdownMap(categoryRows),
		ByDifficulty:     toProgressBreakdownMap(difficultyRows),
	}, nil
}

func (s *QueryService) GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) ([]dto.TeacherRecommendationItem, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	result, err := s.recommendationService.RecommendWithContext(ctx, student.ID, limit)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]dto.TeacherRecommendationItem, 0, len(result.Challenges))
	for _, challenge := range result.Challenges {
		items = append(items, dto.TeacherRecommendationItem{
			ChallengeID: challenge.ID,
			Title:       challenge.Title,
			Category:    challenge.Category,
			Difficulty:  challenge.Difficulty,
			Reason:      challenge.Reason,
		})
	}

	return items, nil
}

func (s *QueryService) GetStudentTimeline(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit, offset int) (*dto.TimelineResp, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.GetStudentTimeline(ctx, student.ID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TimelineResp{Events: events}, nil
}

func (s *QueryService) getAccessibleStudent(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*model.User, error) {
	student, err := s.repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, student.ClassName); err != nil {
		return nil, err
	}
	return student, nil
}

func (s *QueryService) ensureClassAccess(ctx context.Context, requesterID int64, requesterRole, className string) error {
	if requesterRole == model.RoleAdmin {
		return nil
	}

	requester, err := s.repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if requester == nil {
		return errcode.ErrUnauthorized
	}

	if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != className {
		return errcode.ErrForbidden
	}
	return nil
}

func toProgressBreakdownMap(rows []readmodelinfra.ProgressRow) map[string]dto.ProgressBreakdown {
	if len(rows) == 0 {
		return map[string]dto.ProgressBreakdown{}
	}

	result := make(map[string]dto.ProgressBreakdown, len(rows))
	for _, row := range rows {
		result[row.Key] = dto.ProgressBreakdown{
			Total:  row.Total,
			Solved: row.Solved,
		}
	}
	return result
}

func selectRiskStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
	filtered := make([]dto.TeacherStudentItem, 0, len(students))
	for _, student := range students {
		if student.RecentEventCount <= 1 || student.SolvedCount <= 1 {
			filtered = append(filtered, student)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].RecentEventCount != filtered[j].RecentEventCount {
			return filtered[i].RecentEventCount < filtered[j].RecentEventCount
		}
		if filtered[i].SolvedCount != filtered[j].SolvedCount {
			return filtered[i].SolvedCount < filtered[j].SolvedCount
		}
		return filtered[i].Username < filtered[j].Username
	})

	return limitStudents(filtered, limit)
}

func selectWeakDimensionStudents(students []dto.TeacherStudentItem) (string, []dto.TeacherStudentItem) {
	counter := make(map[string]int)
	grouped := make(map[string][]dto.TeacherStudentItem)
	for _, student := range students {
		if student.WeakDimension == nil {
			continue
		}
		key := strings.TrimSpace(*student.WeakDimension)
		if key == "" {
			continue
		}
		counter[key]++
		grouped[key] = append(grouped[key], student)
	}

	bestDimension := ""
	bestCount := 0
	for dimension, count := range counter {
		if count > bestCount || (count == bestCount && (bestDimension == "" || dimension < bestDimension)) {
			bestDimension = dimension
			bestCount = count
		}
	}
	if bestDimension == "" {
		return "", nil
	}

	studentsInDimension := grouped[bestDimension]
	sort.Slice(studentsInDimension, func(i, j int) bool {
		if studentsInDimension[i].SolvedCount != studentsInDimension[j].SolvedCount {
			return studentsInDimension[i].SolvedCount < studentsInDimension[j].SolvedCount
		}
		if studentsInDimension[i].RecentEventCount != studentsInDimension[j].RecentEventCount {
			return studentsInDimension[i].RecentEventCount < studentsInDimension[j].RecentEventCount
		}
		return studentsInDimension[i].Username < studentsInDimension[j].Username
	})
	return bestDimension, studentsInDimension
}

func limitStudents(students []dto.TeacherStudentItem, limit int) []dto.TeacherStudentItem {
	if limit <= 0 || len(students) <= limit {
		return students
	}
	return students[:limit]
}

func toReviewStudentRefs(students []dto.TeacherStudentItem) []dto.TeacherReviewStudentRef {
	items := make([]dto.TeacherReviewStudentRef, 0, len(students))
	for _, student := range students {
		items = append(items, dto.TeacherReviewStudentRef{
			ID:       student.ID,
			Username: student.Username,
			Name:     student.Name,
		})
	}
	return items
}

func joinStudentNames(students []dto.TeacherStudentItem) string {
	names := make([]string, 0, len(students))
	for _, student := range students {
		if student.Name != nil && strings.TrimSpace(*student.Name) != "" {
			names = append(names, strings.TrimSpace(*student.Name))
			continue
		}
		names = append(names, student.Username)
	}
	return strings.Join(names, "、")
}

func buildTrendReviewItem(trend *dto.TeacherClassTrendResp) *dto.TeacherClassReviewItem {
	if trend == nil || len(trend.Points) < 2 {
		return nil
	}
	first := trend.Points[0]
	last := trend.Points[len(trend.Points)-1]
	eventDelta := last.EventCount - first.EventCount
	solveDelta := last.SolveCount - first.SolveCount

	item := &dto.TeacherClassReviewItem{
		Key:    "trend",
		Title:  "观察最近一周走势",
		Accent: "success",
		Detail: fmt.Sprintf("最近一周训练事件变化 %d，成功解题变化 %d，说明班级训练投入仍在向前推进。", eventDelta, solveDelta),
	}
	if solveDelta < 0 || eventDelta < 0 {
		item.Accent = "warning"
		item.Detail = fmt.Sprintf("最近一周训练事件变化 %d，成功解题变化 %d，需要关注训练投入是否正在下滑。", eventDelta, solveDelta)
	}
	return item
}

func (s *QueryService) firstStudentRecommendation(ctx context.Context, students []dto.TeacherStudentItem, limit int) *dto.TeacherRecommendationItem {
	if s.recommendationService == nil {
		return nil
	}
	for _, student := range students {
		result, err := s.recommendationService.RecommendWithContext(ctx, student.ID, limit)
		if err != nil {
			s.logger.Warn("recommend_student_for_class_review_failed", zap.Int64("student_id", student.ID), zap.Error(err))
			continue
		}
		if result == nil || len(result.Challenges) == 0 || result.Challenges[0] == nil {
			continue
		}
		recommendation := result.Challenges[0]
		return &dto.TeacherRecommendationItem{
			ChallengeID: recommendation.ID,
			Title:       recommendation.Title,
			Category:    recommendation.Category,
			Difficulty:  recommendation.Difficulty,
			Reason:      recommendation.Reason,
		}
	}
	return nil
}
