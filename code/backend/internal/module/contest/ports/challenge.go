package ports

import (
	"context"
	"errors"

	contestentity "ctf-platform/internal/module/contest/entity"
)

var (
	ErrContestChallengeEntityNotFound = errors.New("contest challenge entity not found")
)

type ContestChallengeWriteRepository interface {
	AddChallenge(ctx context.Context, cc *contestentity.ContestChallenge) error
	RemoveChallenge(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error
	Exists(ctx context.Context, contestID, challengeID int64) (bool, error)
	HasSubmissions(ctx context.Context, contestID, challengeID int64) (bool, error)
}

type ContestChallengeReadRepository interface {
	FindChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error)
	ListChallenges(ctx context.Context, contestID int64, visibleOnly bool) ([]*contestentity.ContestChallenge, error)
}

type ContestChallengeAWDServiceListRepository interface {
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]contestentity.ContestAWDService, error)
}
