package ports

import "ctf-platform/internal/model"

type ContestTeamFinder interface {
	FindUserTeamInContest(userID, contestID int64) (*model.Team, error)
}

type ContestTeamRepository interface {
	ContestTeamFinder
	CreateWithMember(team *model.Team, captainID int64) error
	FindByID(id int64) (*model.Team, error)
	DeleteWithMembers(id int64) error
	AddMemberWithLock(contestID, teamID, userID int64) error
	RemoveMember(teamID, userID int64) error
	FindContestRegistration(contestID, userID int64) (*model.ContestRegistration, error)
	GetMembers(teamID int64) ([]*model.TeamMember, error)
	GetMemberCount(teamID int64) (int64, error)
	ListByContest(contestID int64) ([]*model.Team, error)
	GetMemberCountBatch(teamIDs []int64) (map[int64]int, error)
	FindUsersByIDs(ids []int64) ([]*model.User, error)
	IsUniqueViolation(err error, constraint string) bool
}
