package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdJobRepositorySource interface {
	contestports.AWDRoundReconcileTxRunner
	contestports.AWDRoundServiceWritebackTxRunner
	contestports.AWDRoundStore
	contestports.AWDContestScheduleQuery
	contestports.AWDTeamLookup
	contestports.AWDServiceDefinitionQuery
	contestports.AWDServiceInstanceQuery
	contestports.AWDServiceOperationQuery
}

type AWDJobRepository struct {
	awdJobRepositorySource
}

func NewAWDJobRepository(source awdJobRepositorySource) *AWDJobRepository {
	if source == nil {
		return nil
	}
	return &AWDJobRepository{awdJobRepositorySource: source}
}

func (r *AWDJobRepository) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	round, err := r.awdJobRepositorySource.FindRoundByNumber(ctx, contestID, roundNumber)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDJobRepository) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	round, err := r.awdJobRepositorySource.FindRunningRound(ctx, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

var _ contestports.AWDRoundReconcileTxRunner = (*AWDJobRepository)(nil)
var _ contestports.AWDRoundServiceWritebackTxRunner = (*AWDJobRepository)(nil)
var _ contestports.AWDRoundStore = (*AWDJobRepository)(nil)
var _ contestports.AWDContestScheduleQuery = (*AWDJobRepository)(nil)
var _ contestports.AWDTeamLookup = (*AWDJobRepository)(nil)
var _ contestports.AWDServiceDefinitionQuery = (*AWDJobRepository)(nil)
var _ contestports.AWDServiceInstanceQuery = (*AWDJobRepository)(nil)
var _ contestports.AWDServiceOperationQuery = (*AWDJobRepository)(nil)
