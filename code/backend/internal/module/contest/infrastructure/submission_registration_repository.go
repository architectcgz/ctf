package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type SubmissionRegistrationRepository struct {
	source interface {
		contestports.ContestSubmissionScoringTxRunner
		contestports.ContestSubmissionRegistrationLookupRepository
		contestports.ContestSubmissionChallengeLookupRepository
		contestports.ContestSubmissionWriteRepository
	}
}

func NewSubmissionRegistrationRepository(source interface {
	contestports.ContestSubmissionScoringTxRunner
	contestports.ContestSubmissionRegistrationLookupRepository
	contestports.ContestSubmissionChallengeLookupRepository
	contestports.ContestSubmissionWriteRepository
}) *SubmissionRegistrationRepository {
	if source == nil {
		return nil
	}
	return &SubmissionRegistrationRepository{source: source}
}

func (r *SubmissionRegistrationRepository) WithinScoringTransaction(ctx context.Context, fn func(repo contestports.ContestSubmissionScoringTxRepository) error) error {
	return r.source.WithinScoringTransaction(ctx, fn)
}

func (r *SubmissionRegistrationRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	registration, err := r.source.FindRegistration(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestParticipationRegistrationNotFound
	}
	return registration, err
}

func (r *SubmissionRegistrationRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	return r.source.FindContestChallenge(ctx, contestID, challengeID)
}

func (r *SubmissionRegistrationRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	return r.source.FindChallengeByID(ctx, challengeID)
}

func (r *SubmissionRegistrationRepository) CreateSubmission(ctx context.Context, submission *model.Submission) error {
	return r.source.CreateSubmission(ctx, submission)
}

var _ contestports.ContestSubmissionScoringTxRunner = (*SubmissionRegistrationRepository)(nil)
var _ contestports.ContestSubmissionRegistrationLookupRepository = (*SubmissionRegistrationRepository)(nil)
var _ contestports.ContestSubmissionChallengeLookupRepository = (*SubmissionRegistrationRepository)(nil)
var _ contestports.ContestSubmissionWriteRepository = (*SubmissionRegistrationRepository)(nil)
