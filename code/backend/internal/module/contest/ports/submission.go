package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type ContestSubmissionScoringTxRepository interface {
	LockContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	CountCorrectSubmissions(ctx context.Context, contestID, challengeID int64, teamID *int64, userID int64) (int64, error)
	UpdateFirstBlood(ctx context.Context, contestID, challengeID, teamID int64) error
	ListCorrectSubmissions(ctx context.Context, contestID, challengeID int64) ([]model.Submission, error)
	UpdateSubmissionScore(ctx context.Context, submissionID int64, score int) error
	AddTeamScore(ctx context.Context, teamID int64, delta int, lastSolveAt *time.Time) error
	CreateSubmission(ctx context.Context, submission *model.Submission) error
}

type ContestSubmissionScoringTxRunner interface {
	WithinScoringTransaction(ctx context.Context, fn func(repo ContestSubmissionScoringTxRepository) error) error
}

type ContestSubmissionRegistrationLookupRepository interface {
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
}

type ContestSubmissionChallengeLookupRepository interface {
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error)
}

type ContestSubmissionWriteRepository interface {
	CreateSubmission(ctx context.Context, submission *model.Submission) error
}
