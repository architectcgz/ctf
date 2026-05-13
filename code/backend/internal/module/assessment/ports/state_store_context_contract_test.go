package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	assessmentports "ctf-platform/internal/module/assessment/ports"
)

type ctxOnlyAssessmentProfileLockLease struct{}

func (ctxOnlyAssessmentProfileLockLease) Release(context.Context) (bool, error) {
	return true, nil
}

type ctxOnlyAssessmentProfileLockStore struct{}

func (ctxOnlyAssessmentProfileLockStore) AcquireDimensionUpdateLock(context.Context, int64, string, time.Duration) (assessmentports.AssessmentProfileLockLease, bool, error) {
	return ctxOnlyAssessmentProfileLockLease{}, true, nil
}

func (ctxOnlyAssessmentProfileLockStore) AcquireFullProfileRebuildLock(context.Context, int64, time.Duration) (assessmentports.AssessmentProfileLockLease, bool, error) {
	return ctxOnlyAssessmentProfileLockLease{}, true, nil
}

type ctxOnlyAssessmentRecommendationCacheStore struct{}

func (ctxOnlyAssessmentRecommendationCacheStore) LoadRecommendations(context.Context, int64) ([]*dto.ChallengeRecommendation, bool, error) {
	return nil, false, nil
}

func (ctxOnlyAssessmentRecommendationCacheStore) StoreRecommendations(context.Context, int64, []*dto.ChallengeRecommendation, time.Duration) error {
	return nil
}

func (ctxOnlyAssessmentRecommendationCacheStore) DeleteRecommendations(context.Context, int64) error {
	return nil
}

var _ assessmentports.AssessmentProfileLockLease = (*ctxOnlyAssessmentProfileLockLease)(nil)
var _ assessmentports.AssessmentProfileLockStore = (*ctxOnlyAssessmentProfileLockStore)(nil)
var _ assessmentports.AssessmentRecommendationCacheStore = (*ctxOnlyAssessmentRecommendationCacheStore)(nil)
