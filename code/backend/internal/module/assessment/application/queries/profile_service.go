package queries

import (
	"context"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	"ctf-platform/pkg/errcode"
)

type ProfileService struct {
	repo profileQueryRepository
}

type profileQueryRepository interface {
	assessmentports.AssessmentProfileLookupRepository
	assessmentports.AssessmentProfileReadRepository
}

func NewProfileService(repo profileQueryRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetSkillProfile(ctx context.Context, userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return assessmentdomain.BuildEmptyProfile(userID), nil
	}
	return assessmentdomain.BuildSkillProfile(userID, profiles), nil
}

func (s *ProfileService) GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.SkillProfileResp, error) {
	student, err := s.repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, findErr := s.repo.FindUserByID(ctx, requesterID)
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

	return s.GetSkillProfile(ctx, studentID)
}
