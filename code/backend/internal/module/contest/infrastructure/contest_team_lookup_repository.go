package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (r *Repository) FindTeamsByIDs(ctx context.Context, ids []int64) ([]*model.Team, error) {
	if len(ids) == 0 {
		return []*model.Team{}, nil
	}

	var teams []*model.Team
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&teams).Error
	return teams, err
}

func (r *Repository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error
	return teams, err
}
