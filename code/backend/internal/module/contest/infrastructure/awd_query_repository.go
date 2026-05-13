package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdQueryRepositorySource interface {
	contestports.AWDRoundStore
	contestports.AWDTeamLookup
	contestports.AWDServiceDefinitionQuery
	contestports.AWDReadinessQuery
	contestports.AWDServiceInstanceQuery
	contestports.AWDDefenseWorkspaceSummaryQuery
	contestports.AWDServiceOperationQuery
	contestports.AWDTeamServiceStore
	contestports.AWDAttackLogStore
	contestports.AWDTrafficEventQuery
}

type AWDQueryRepository struct {
	awdQueryRepositorySource
}

func NewAWDQueryRepository(source awdQueryRepositorySource) *AWDQueryRepository {
	if source == nil {
		return nil
	}
	return &AWDQueryRepository{awdQueryRepositorySource: source}
}

func (r *AWDQueryRepository) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	round, err := r.awdQueryRepositorySource.FindRoundByContestAndID(ctx, contestID, roundID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDQueryRepository) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	round, err := r.awdQueryRepositorySource.FindRunningRound(ctx, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDQueryRepository) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	team, err := r.awdQueryRepositorySource.FindContestTeamByMember(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestUserTeamNotFound
	}
	return team, err
}

var _ contestports.AWDRoundStore = (*AWDQueryRepository)(nil)
var _ contestports.AWDTeamLookup = (*AWDQueryRepository)(nil)
var _ contestports.AWDServiceDefinitionQuery = (*AWDQueryRepository)(nil)
var _ contestports.AWDReadinessQuery = (*AWDQueryRepository)(nil)
var _ contestports.AWDServiceInstanceQuery = (*AWDQueryRepository)(nil)
var _ contestports.AWDDefenseWorkspaceSummaryQuery = (*AWDQueryRepository)(nil)
var _ contestports.AWDServiceOperationQuery = (*AWDQueryRepository)(nil)
var _ contestports.AWDTeamServiceStore = (*AWDQueryRepository)(nil)
var _ contestports.AWDAttackLogStore = (*AWDQueryRepository)(nil)
var _ contestports.AWDTrafficEventQuery = (*AWDQueryRepository)(nil)
