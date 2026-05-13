package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type TeamFinderRepository struct {
	source contestports.ContestTeamFinder
}

func NewTeamFinderRepository(source contestports.ContestTeamFinder) *TeamFinderRepository {
	if source == nil {
		return nil
	}
	return &TeamFinderRepository{source: source}
}

func (r *TeamFinderRepository) FindUserTeamInContest(ctx context.Context, userID, contestID int64) (*model.Team, error) {
	team, err := r.source.FindUserTeamInContest(ctx, userID, contestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, contestports.ErrContestUserTeamNotFound
	}
	return team, err
}

var _ contestports.ContestTeamFinder = (*TeamFinderRepository)(nil)
