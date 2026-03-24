package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type Repository interface {
	Create(ctx context.Context, contest *model.Contest) error
	FindByID(ctx context.Context, id int64) (*model.Contest, error)
	Update(ctx context.Context, contest *model.Contest) error
	List(ctx context.Context, status *string, offset, limit int) ([]*model.Contest, int64, error)
	ListByStatusesAndTimeRange(ctx context.Context, statuses []string, now time.Time, offset, limit int) ([]*model.Contest, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error)
	FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	FindScoreboardTeamStats(ctx context.Context, contestID int64, contestMode string, teamIDs []int64) (map[int64]ScoreboardTeamStats, error)
}
