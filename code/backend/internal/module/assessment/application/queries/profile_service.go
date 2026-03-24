package queries

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	"ctf-platform/pkg/errcode"
)

type ProfileService struct {
	repo assessmentports.ProfileRepository
}

func NewProfileService(repo assessmentports.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetSkillProfile(userID int64) (*dto.SkillProfileResp, error) {
	return s.GetSkillProfileWithContext(context.Background(), userID)
}

func (s *ProfileService) GetSkillProfileWithContext(ctx context.Context, userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserIDWithContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return assessmentdomain.BuildEmptyProfile(userID), nil
	}

	dimensionMap := make(map[string]float64)
	var latestUpdate time.Time
	for _, profile := range profiles {
		dimensionMap[profile.Dimension] = profile.Score
		if profile.UpdatedAt.After(latestUpdate) {
			latestUpdate = profile.UpdatedAt
		}
	}

	dimensions := make([]*dto.SkillDimension, 0, len(model.AllDimensions))
	for _, dim := range model.AllDimensions {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: dim,
			Score:     dimensionMap[dim],
		})
	}

	return &dto.SkillProfileResp{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  latestUpdate.Format(time.RFC3339),
	}, nil
}

func (s *ProfileService) GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.SkillProfileResp, error) {
	student, err := s.repo.FindUserByIDWithContext(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, findErr := s.repo.FindUserByIDWithContext(ctx, requesterID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		if requester == nil {
			return nil, errcode.ErrUnauthorized
		}
		if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != student.ClassName {
			return nil, errcode.ErrForbidden
		}
	}

	return s.GetSkillProfileWithContext(ctx, studentID)
}
