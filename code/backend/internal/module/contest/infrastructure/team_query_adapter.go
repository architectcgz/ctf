package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type TeamQueryAdapter struct {
	source interface {
		contestports.ContestTeamFinder
		contestports.ContestTeamLookupRepository
		contestports.ContestTeamMembershipRepository
		contestports.ContestTeamListRepository
		contestports.ContestTeamUserLookupRepository
	}
}

func NewTeamQueryAdapter(source interface {
	contestports.ContestTeamFinder
	contestports.ContestTeamLookupRepository
	contestports.ContestTeamMembershipRepository
	contestports.ContestTeamListRepository
	contestports.ContestTeamUserLookupRepository
}) *TeamQueryAdapter {
	if source == nil {
		return nil
	}
	return &TeamQueryAdapter{source: source}
}

func (r *TeamQueryAdapter) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	team, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestTeamNotFound
	}
	return team, err
}

func (r *TeamQueryAdapter) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	team, err := r.source.FindUserTeamInContest(ctx, userID, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestUserTeamNotFound
	}
	return team, err
}

func (r *TeamQueryAdapter) GetMembers(ctx context.Context, teamID int64) ([]*model.TeamMember, error) {
	return r.source.GetMembers(ctx, teamID)
}

func (r *TeamQueryAdapter) AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error {
	return r.source.AddMemberWithLock(ctx, contestID, teamID, userID)
}

func (r *TeamQueryAdapter) RemoveMember(ctx context.Context, teamID, userID int64) error {
	return r.source.RemoveMember(ctx, teamID, userID)
}

func (r *TeamQueryAdapter) GetMemberCount(ctx context.Context, teamID int64) (int64, error) {
	return r.source.GetMemberCount(ctx, teamID)
}

func (r *TeamQueryAdapter) GetMemberCountBatch(ctx context.Context, teamIDs []int64) (map[int64]int, error) {
	return r.source.GetMemberCountBatch(ctx, teamIDs)
}

func (r *TeamQueryAdapter) ListByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	return r.source.ListByContest(ctx, contestID)
}

func (r *TeamQueryAdapter) FindUsersByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	return r.source.FindUsersByIDs(ctx, ids)
}

var _ contestports.ContestTeamFinder = (*TeamQueryAdapter)(nil)
var _ contestports.ContestTeamLookupRepository = (*TeamQueryAdapter)(nil)
var _ contestports.ContestTeamMembershipRepository = (*TeamQueryAdapter)(nil)
var _ contestports.ContestTeamListRepository = (*TeamQueryAdapter)(nil)
var _ contestports.ContestTeamUserLookupRepository = (*TeamQueryAdapter)(nil)
