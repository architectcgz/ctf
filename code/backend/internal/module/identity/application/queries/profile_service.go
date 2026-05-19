package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/apperror"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

type ProfileService struct {
	users identitycontracts.UserLookupRepository
}

var _ identitycontracts.ProfileQueryService = (*ProfileService)(nil)

func NewProfileService(users identitycontracts.UserLookupRepository) *ProfileService {
	return &ProfileService{users: users}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*identitycontracts.ProfileUser, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitycontracts.ErrUserNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	profile := buildProfileUser(user)
	return &profile, nil
}

func buildProfileUser(user *identitycontracts.User) identitycontracts.ProfileUser {
	resp := adminUserMapper.ToProfileUserBasePtr(user)
	resp.Name = commonmapper.NormalizeOptionalTrimmedString(user.Name)
	resp.ClassName = commonmapper.NormalizeOptionalTrimmedString(user.ClassName)
	return *resp
}
