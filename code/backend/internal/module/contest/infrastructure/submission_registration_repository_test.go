package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

type submissionRegistrationSourceStub struct {
	findRegistrationFn     func(context.Context, int64, int64) (*contestentity.ContestRegistration, error)
	findContestChallengeFn func(context.Context, int64, int64) (*contestentity.ContestChallenge, error)
	findChallengeByIDFn    func(context.Context, int64) (*model.Challenge, error)
}

func (s submissionRegistrationSourceStub) WithinScoringTransaction(context.Context, func(contestports.ContestSubmissionScoringTxRepository) error) error {
	return nil
}

func (s submissionRegistrationSourceStub) FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error) {
	if s.findRegistrationFn != nil {
		return s.findRegistrationFn(ctx, contestID, userID)
	}
	return &contestentity.ContestRegistration{}, nil
}

func (s submissionRegistrationSourceStub) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error) {
	if s.findContestChallengeFn != nil {
		return s.findContestChallengeFn(ctx, contestID, challengeID)
	}
	return &contestentity.ContestChallenge{}, nil
}

func (s submissionRegistrationSourceStub) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	if s.findChallengeByIDFn != nil {
		return s.findChallengeByIDFn(ctx, challengeID)
	}
	return &model.Challenge{}, nil
}

func (s submissionRegistrationSourceStub) CreateSubmission(context.Context, *contestentity.Submission) error {
	return nil
}

func TestSubmissionRegistrationRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewSubmissionRegistrationRepository(submissionRegistrationSourceStub{
		findRegistrationFn: func(context.Context, int64, int64) (*contestentity.ContestRegistration, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindRegistration(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestParticipationRegistrationNotFound)
	}
}

func TestSubmissionRegistrationRepositoryMapsContestChallengeNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewSubmissionRegistrationRepository(submissionRegistrationSourceStub{
		findContestChallengeFn: func(context.Context, int64, int64) (*contestentity.ContestChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindContestChallenge(context.Background(), 1, 2); !errors.Is(err, contestports.ErrContestSubmissionChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestSubmissionChallengeNotFound)
	}
}

func TestSubmissionRegistrationRepositoryMapsChallengeEntityNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewSubmissionRegistrationRepository(submissionRegistrationSourceStub{
		findChallengeByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindChallengeByID(context.Background(), 2); !errors.Is(err, contestports.ErrContestSubmissionChallengeEntityNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestSubmissionChallengeEntityNotFound)
	}
}
