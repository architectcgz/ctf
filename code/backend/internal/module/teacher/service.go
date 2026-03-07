package teacher

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type RecommendationProvider interface {
	Recommend(userID int64, limit int) (*dto.RecommendationResp, error)
}

type Service struct {
	repo                  *Repository
	recommendationService RecommendationProvider
	logger                *zap.Logger
}

func NewService(repo *Repository, recommendationService RecommendationProvider, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repo:                  repo,
		recommendationService: recommendationService,
		logger:                logger,
	}
}

func (s *Service) ListClasses(ctx context.Context, requesterID int64, requesterRole string) ([]dto.TeacherClassItem, error) {
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

func (s *Service) ListClassStudents(ctx context.Context, requesterID int64, requesterRole, className string) ([]dto.TeacherStudentItem, error) {
	normalized := strings.TrimSpace(className)
	if normalized == "" {
		return nil, errcode.New(errcode.ErrInvalidParams.Code, "class_name 不能为空", errcode.ErrInvalidParams.HTTPStatus)
	}

	if err := s.ensureClassAccess(ctx, requesterID, requesterRole, normalized); err != nil {
		return nil, err
	}

	items, err := s.repo.ListStudentsByClass(ctx, normalized)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return items, nil
}

func (s *Service) GetStudentProgress(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.TeacherProgressResp, error) {
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

func (s *Service) GetStudentRecommendations(ctx context.Context, requesterID int64, requesterRole string, studentID int64, limit int) ([]dto.TeacherRecommendationItem, error) {
	student, err := s.getAccessibleStudent(ctx, requesterID, requesterRole, studentID)
	if err != nil {
		return nil, err
	}

	result, err := s.recommendationService.Recommend(student.ID, limit)
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

func (s *Service) getAccessibleStudent(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*model.User, error) {
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

func (s *Service) ensureClassAccess(ctx context.Context, requesterID int64, requesterRole, className string) error {
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

func toProgressBreakdownMap(rows []progressRow) map[string]dto.ProgressBreakdown {
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
