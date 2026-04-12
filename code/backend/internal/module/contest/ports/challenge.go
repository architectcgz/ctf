package ports

import (
	"context"

	"ctf-platform/internal/model"
)

type ContestChallengeRepository interface {
	AddChallenge(ctx context.Context, cc *model.ContestChallenge) error
	FindChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error)
	RemoveChallenge(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, updates map[string]any) error
	ListChallenges(ctx context.Context, contestID int64, visibleOnly bool) ([]*model.ContestChallenge, error)
	Exists(ctx context.Context, contestID, challengeID int64) (bool, error)
	HasSubmissions(ctx context.Context, contestID, challengeID int64) (bool, error)
}
