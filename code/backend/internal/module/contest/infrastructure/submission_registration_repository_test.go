package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type submissionRegistrationSourceStub struct {
	findRegistrationFn func(context.Context, int64, int64) (*model.ContestRegistration, error)
}

func (s submissionRegistrationSourceStub) WithinScoringTransaction(context.Context, func(contestports.ContestSubmissionScoringTxRepository) error) error {
	return nil
}

func (s submissionRegistrationSourceStub) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return &model.ContestRegistration{}, nil
}

func (s submissionRegistrationSourceStub) FindContestChallenge(context.Context, int64, int64) (*model.ContestChallenge, error) {
	return &model.ContestChallenge{}, nil
}

func (s submissionRegistrationSourceStub) FindChallengeByID(context.Context, int64) (*model.Challenge, error) {
	return &model.Challenge{}, nil
}

func (s submissionRegistrationSourceStub) CreateSubmission(context.Context, *model.Submission) error {
	return nil
}

func TestSubmissionRegistrationRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewSubmissionRegistrationRepository(submissionRegistrationSourceStub{
		findRegistrationFn: func(context.Context, int64, int64) (*model.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRegistration(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}
