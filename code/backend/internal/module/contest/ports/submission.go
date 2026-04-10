package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type ContestSubmissionRepository interface {
	WithinTransaction(ctx context.Context, fn func(repo ContestSubmissionRepository) error) error
	FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error)
	FindActiveSharedProofByHash(ctx context.Context, proofHash string) (*model.SharedProof, error)
	CreateSubmission(ctx context.Context, submission *model.Submission) error
	ConsumeSharedProof(ctx context.Context, sharedProofID int64, consumedAt time.Time) (bool, error)
	LockContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	CountCorrectSubmissions(ctx context.Context, contestID, challengeID int64, teamID *int64, userID int64) (int64, error)
	UpdateFirstBlood(ctx context.Context, contestID, challengeID, teamID int64) error
	ListCorrectSubmissions(ctx context.Context, contestID, challengeID int64) ([]model.Submission, error)
	UpdateSubmissionScore(ctx context.Context, submissionID int64, score int) error
	AddTeamScore(ctx context.Context, teamID int64, delta int, lastSolveAt *time.Time) error
}
