package ports

import (
	"context"
	"errors"

	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

var ErrContestUserTeamNotFound = errors.New("contest user team not found")
var ErrContestTeamNotFound = errors.New("contest team not found")

type ContestTeamFinder interface {
	FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*contestentity.Team, error)
}

type ContestTeamWriteRepository interface {
	CreateWithMember(ctx context.Context, team *contestentity.Team, captainID int64) error
	DeleteWithMembers(ctx context.Context, id int64) error
	IsUniqueViolation(err error, constraint string) bool
}

type ContestTeamLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*contestentity.Team, error)
}

type ContestTeamMembershipRepository interface {
	AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error
	RemoveMember(ctx context.Context, teamID, userID int64) error
	GetMembers(ctx context.Context, teamID int64) ([]*contestentity.TeamMember, error)
	GetMemberCount(ctx context.Context, teamID int64) (int64, error)
	GetMemberCountBatch(ctx context.Context, teamIDs []int64) (map[int64]int, error)
}

type ContestTeamRegistrationLookupRepository interface {
	FindContestRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error)
}

type ContestTeamListRepository interface {
	ListByContest(ctx context.Context, contestID int64) ([]*contestentity.Team, error)
}

type ContestTeamUserLookupRepository interface {
	FindUsersByIDs(ctx context.Context, ids []int64) ([]*identitycontracts.User, error)
}
