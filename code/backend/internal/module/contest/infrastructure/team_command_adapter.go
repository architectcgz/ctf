package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type TeamCommandAdapter struct {
	source interface {
		contestports.ContestTeamFinder
		contestports.ContestTeamWriteRepository
		contestports.ContestTeamLookupRepository
		contestports.ContestTeamMembershipRepository
		contestports.ContestTeamRegistrationLookupRepository
	}
}

func NewTeamCommandAdapter(source interface {
	contestports.ContestTeamFinder
	contestports.ContestTeamWriteRepository
	contestports.ContestTeamLookupRepository
	contestports.ContestTeamMembershipRepository
	contestports.ContestTeamRegistrationLookupRepository
}) *TeamCommandAdapter {
	if source == nil {
		return nil
	}
	return &TeamCommandAdapter{source: source}
}

func (r *TeamCommandAdapter) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	team, err := r.source.FindUserTeamInContest(ctx, userID, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestUserTeamNotFound
	}
	return team, err
}

func (r *TeamCommandAdapter) CreateWithMember(ctx context.Context, team *model.Team, captainID int64) error {
	return mapCommandRegistrationNotFound(r.source.CreateWithMember(ctx, team, captainID))
}

func (r *TeamCommandAdapter) DeleteWithMembers(ctx context.Context, id int64) error {
	return r.source.DeleteWithMembers(ctx, id)
}

func (r *TeamCommandAdapter) IsUniqueViolation(err error, constraint string) bool {
	return r.source.IsUniqueViolation(err, constraint)
}

func (r *TeamCommandAdapter) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	team, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestTeamNotFound
	}
	return team, err
}

func (r *TeamCommandAdapter) AddMemberWithLock(ctx context.Context, contestID, teamID, userID int64) error {
	return mapCommandRegistrationNotFound(r.source.AddMemberWithLock(ctx, contestID, teamID, userID))
}

func (r *TeamCommandAdapter) RemoveMember(ctx context.Context, teamID, userID int64) error {
	return r.source.RemoveMember(ctx, teamID, userID)
}

func (r *TeamCommandAdapter) GetMembers(ctx context.Context, teamID int64) ([]*model.TeamMember, error) {
	return r.source.GetMembers(ctx, teamID)
}

func (r *TeamCommandAdapter) GetMemberCount(ctx context.Context, teamID int64) (int64, error) {
	return r.source.GetMemberCount(ctx, teamID)
}

func (r *TeamCommandAdapter) GetMemberCountBatch(ctx context.Context, teamIDs []int64) (map[int64]int, error) {
	return r.source.GetMemberCountBatch(ctx, teamIDs)
}

func (r *TeamCommandAdapter) FindContestRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	registration, err := r.source.FindContestRegistration(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestParticipationRegistrationNotFound
	}
	return registration, err
}

func mapCommandRegistrationNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contestports.ErrContestParticipationRegistrationNotFound
	}
	return err
}

var _ contestports.ContestTeamFinder = (*TeamCommandAdapter)(nil)
var _ contestports.ContestTeamWriteRepository = (*TeamCommandAdapter)(nil)
var _ contestports.ContestTeamLookupRepository = (*TeamCommandAdapter)(nil)
var _ contestports.ContestTeamMembershipRepository = (*TeamCommandAdapter)(nil)
var _ contestports.ContestTeamRegistrationLookupRepository = (*TeamCommandAdapter)(nil)
