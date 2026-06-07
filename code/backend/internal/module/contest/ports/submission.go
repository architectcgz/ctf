package ports

import (
	"context"
	"errors"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
)

var ErrContestSubmissionChallengeNotFound = errors.New("contest submission challenge not found")
var ErrContestSubmissionChallengeEntityNotFound = errors.New("contest submission challenge entity not found")

type ContestSubmissionScoringTxRepository interface {
	LockContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error)
	CountCorrectSubmissions(ctx context.Context, contestID, challengeID int64, teamID *int64, userID int64) (int64, error)
	UpdateFirstBlood(ctx context.Context, contestID, challengeID, teamID int64) error
	ListCorrectSubmissions(ctx context.Context, contestID, challengeID int64) ([]contestentity.Submission, error)
	UpdateSubmissionScore(ctx context.Context, submissionID int64, score int) error
	AddTeamScore(ctx context.Context, teamID int64, delta int, lastSolveAt *time.Time) error
	CreateSubmission(ctx context.Context, submission *contestentity.Submission) error
	EnqueueRealtimeRelay(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error
}

type ContestSubmissionScoringTxRunner interface {
	WithinScoringTransaction(ctx context.Context, fn func(repo ContestSubmissionScoringTxRepository) error) error
}

type ContestSubmissionRegistrationLookupRepository interface {
	FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error)
}

type ContestSubmissionChallengeLookupRepository interface {
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*contestentity.Challenge, error)
}

type ContestSubmissionWriteRepository interface {
	CreateSubmission(ctx context.Context, submission *contestentity.Submission) error
}

type ContestSubmissionRateLimitStore interface {
	HasIncorrectSubmissionRateLimit(ctx context.Context, userID, contestID, challengeID int64) (bool, error)
	SetIncorrectSubmissionRateLimit(ctx context.Context, userID, contestID, challengeID int64, ttl time.Duration) error
}
