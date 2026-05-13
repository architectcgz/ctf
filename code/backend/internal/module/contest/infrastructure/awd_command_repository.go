package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdCommandRepositorySource interface {
	contestports.AWDServiceCheckTxRunner
	contestports.AWDAttackLogTxRunner
	contestports.AWDServiceStore
	contestports.AWDRoundStore
	contestports.AWDTeamLookup
	contestports.AWDChallengeLookup
	contestports.AWDReadinessQuery
	contestports.AWDTeamServiceStore
	contestports.AWDAttackLogStore
}

type AWDCommandRepository struct {
	awdCommandRepositorySource
}

func NewAWDCommandRepository(source awdCommandRepositorySource) *AWDCommandRepository {
	if source == nil {
		return nil
	}
	return &AWDCommandRepository{awdCommandRepositorySource: source}
}

func (r *AWDCommandRepository) FindRoundByContestAndID(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	round, err := r.awdCommandRepositorySource.FindRoundByContestAndID(ctx, contestID, roundID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDCommandRepository) FindRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	round, err := r.awdCommandRepositorySource.FindRoundByNumber(ctx, contestID, roundNumber)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDCommandRepository) FindRunningRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	round, err := r.awdCommandRepositorySource.FindRunningRound(ctx, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDRoundNotFound
	}
	return round, err
}

func (r *AWDCommandRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	registration, err := r.awdCommandRepositorySource.FindRegistration(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestParticipationRegistrationNotFound
	}
	return registration, err
}

func (r *AWDCommandRepository) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	team, err := r.awdCommandRepositorySource.FindContestTeamByMember(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestUserTeamNotFound
	}
	return team, err
}

func (r *AWDCommandRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	challenge, err := r.awdCommandRepositorySource.FindChallengeByID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDChallengeNotFound
	}
	return challenge, err
}

func (r *AWDCommandRepository) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	service, err := r.awdCommandRepositorySource.FindContestAWDServiceByContestAndID(ctx, contestID, serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestAWDServiceNotFound
	}
	return service, err
}

func (r *AWDCommandRepository) WithinAttackLogTransaction(ctx context.Context, fn func(txRepo contestports.AWDAttackLogTxRepository) error) error {
	err := r.awdCommandRepositorySource.WithinAttackLogTransaction(ctx, func(txRepo contestports.AWDAttackLogTxRepository) error {
		return fn(txRepo)
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contestports.ErrContestAWDAttackLogTransactionNotFound
	}
	return err
}

var _ contestports.AWDServiceCheckTxRunner = (*AWDCommandRepository)(nil)
var _ contestports.AWDAttackLogTxRunner = (*AWDCommandRepository)(nil)
var _ contestports.AWDServiceStore = (*AWDCommandRepository)(nil)
var _ contestports.AWDRoundStore = (*AWDCommandRepository)(nil)
var _ contestports.AWDTeamLookup = (*AWDCommandRepository)(nil)
var _ contestports.AWDChallengeLookup = (*AWDCommandRepository)(nil)
var _ contestports.AWDReadinessQuery = (*AWDCommandRepository)(nil)
var _ contestports.AWDTeamServiceStore = (*AWDCommandRepository)(nil)
var _ contestports.AWDAttackLogStore = (*AWDCommandRepository)(nil)
