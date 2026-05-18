package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type contestScopeLookupSource interface {
	practiceports.PracticeContestLookupRepository
	practiceports.PracticeContestChallengeLookupRepository
	practiceports.PracticeContestAWDServiceRepository
	practiceports.PracticeContestAWDServiceRuntimeSubjectRepository
	practiceports.PracticeContestTeamRepository
	practiceports.PracticeContestRegistrationRepository
}

type ContestScopeRepository struct {
	source contestScopeLookupSource
}

func NewContestScopeRepository(source contestScopeLookupSource) *ContestScopeRepository {
	if source == nil {
		return nil
	}
	return &ContestScopeRepository{source: source}
}

func (r *ContestScopeRepository) FindContestByID(ctx context.Context, contestID int64) (*contestentity.Contest, error) {
	contest, err := r.source.FindContestByID(ctx, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestNotFound
	}
	return contest, err
}

func (r *ContestScopeRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error) {
	item, err := r.source.FindContestChallenge(ctx, contestID, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestChallengeNotFound
	}
	return item, err
}

func (r *ContestScopeRepository) FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*contestentity.ContestAWDService, error) {
	service, err := r.source.FindContestAWDService(ctx, contestID, serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestAWDServiceNotFound
	}
	return service, err
}

func (r *ContestScopeRepository) ListContestAWDServices(ctx context.Context, contestID int64) ([]*contestentity.ContestAWDService, error) {
	return r.source.ListContestAWDServices(ctx, contestID)
}

func (r *ContestScopeRepository) FindContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	subject, err := r.source.FindContestAWDServiceRuntimeSubject(ctx, contestID, serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestAWDServiceNotFound
	}
	return subject, err
}

func (r *ContestScopeRepository) FindContestTeam(ctx context.Context, contestID, teamID int64) (*contestentity.Team, error) {
	team, err := r.source.FindContestTeam(ctx, contestID, teamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestTeamNotFound
	}
	return team, err
}

func (r *ContestScopeRepository) ListContestTeams(ctx context.Context, contestID int64) ([]*contestentity.Team, error) {
	return r.source.ListContestTeams(ctx, contestID)
}

func (r *ContestScopeRepository) FindContestRegistration(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
	registration, err := r.source.FindContestRegistration(ctx, contestID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeContestRegistrationNotFound
	}
	return registration, err
}

var _ practiceports.PracticeContestScopeRepository = (*ContestScopeRepository)(nil)
