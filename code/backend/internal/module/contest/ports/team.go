package ports

import (
	"context"

	"ctf-platform/internal/model"
)

type ContestTeamFinder interface {
	FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error)
}

type ContestTeamRepository interface {
	ContestTeamFinder
	CreateWithMember(ctx context.Context, team *model.Team, captainID int64) error
	FindByID(ctx context.Context, id int64) (*model.Team, error)
	DeleteWithMembers(ctx context.Context, id int64) error
	AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error
	RemoveMember(ctx context.Context, teamID, userID int64) error
	FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error)
	GetMembers(ctx context.Context, teamID int64) ([]*model.TeamMember, error)
	GetMemberCount(ctx context.Context, teamID int64) (int64, error)
	ListByContest(ctx context.Context, contestID int64) ([]*model.Team, error)
	GetMemberCountBatch(ctx context.Context, teamIDs []int64) (map[int64]int, error)
	FindUsersByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
	IsUniqueViolation(err error, constraint string) bool
}
