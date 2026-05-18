package queries

import (
	"context"
	"strings"

	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
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

func (s *ProfileService) GetSkillProfile(ctx context.Context, userID int64) (*assessmentcontracts.SkillProfile, error) {
	profiles, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return assessmentdomain.BuildEmptyProfileContract(userID), nil
	}
	return assessmentdomain.BuildSkillProfileContract(userID, profiles), nil
}

func (s *ProfileService) GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*assessmentcontracts.SkillProfile, error) {
	student, err := s.repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != identitycontracts.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if requesterRole != identitycontracts.RoleAdmin {
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
