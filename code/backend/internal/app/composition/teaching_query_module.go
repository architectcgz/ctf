package composition

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachingqueryruntime "ctf-platform/internal/module/teaching_query/runtime"
)

type TeachingQueryModule = teachingqueryruntime.Module

type teachingQueryUserLookupAdapter struct {
	users identitycontracts.UserLookupRepository
}

func (a teachingQueryUserLookupAdapter) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	if a.users == nil {
		return nil, nil
	}

	user, err := a.users.FindByID(ctx, userID)
	if errors.Is(err, identitycontracts.ErrUserNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func BuildTeachingQueryModule(root *Root, assessment *AssessmentModule, identity *IdentityModule) *TeachingQueryModule {
	return teachingqueryruntime.Build(teachingqueryruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		DB:              root.DB(),
		Users:           teachingQueryUserLookupAdapter{users: identity.Users},
		Recommendations: assessment.Recommendations,
	})
}
