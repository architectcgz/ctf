package infrastructure

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *Repository) FindTeamsByIDs(ctx context.Context, ids []int64) ([]*contestentity.Team, error) {
	if len(ids) == 0 {
		return []*contestentity.Team{}, nil
	}

	var teams []*contestentity.Team
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&teams).Error
	return teams, err
}

func (r *Repository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*contestentity.Team, error) {
	var teams []*contestentity.Team
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error
	return teams, err
}
