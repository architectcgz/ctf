package ports

import (
	"context"

	"ctf-platform/internal/model"
)

type ContestChallengeWriteRepository interface {
	AddChallenge(ctx context.Context, cc *model.ContestChallenge) error
	RemoveChallenge(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error
	Exists(ctx context.Context, contestID, challengeID int64) (bool, error)
	HasSubmissions(ctx context.Context, contestID, challengeID int64) (bool, error)
}

type ContestChallengeReadRepository interface {
	FindChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	ListChallenges(ctx context.Context, contestID int64, visibleOnly bool) ([]*model.ContestChallenge, error)
}

type ContestChallengeAWDServiceListRepository interface {
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error)
}
