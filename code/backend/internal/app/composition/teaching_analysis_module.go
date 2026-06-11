package composition

import (
	"context"
	"errors"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	teachinganalysisruntime "ctf-platform/internal/module/teaching_analysis/runtime"
)

type TeachingAnalysisModule = teachinganalysisruntime.Module

type teachingAnalysisUserLookupAdapter struct {
	users identitycontracts.UserLookupRepository
}

func (a teachingAnalysisUserLookupAdapter) FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
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

func BuildTeachingAnalysisModule(root *Root, assessment *AssessmentModule, identity *IdentityModule) *TeachingAnalysisModule {
	return teachinganalysisruntime.Build(teachinganalysisruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		DB:              root.DB(),
		Users:           teachingAnalysisUserLookupAdapter{users: identity.Users},
		Recommendations: assessment.Recommendations,
	})
}
