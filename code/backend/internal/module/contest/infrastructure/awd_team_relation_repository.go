package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (r *AWDRepository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Find(&teams).Error
	return teams, err
}

func (r *AWDRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *AWDRepository) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*model.Team, error) {
	var team model.Team
	if err := r.dbWithContext(ctx).
		Table("teams AS t").
		Select("t.*").
		Joins("JOIN team_members AS tm ON tm.team_id = t.id").
		Where("t.contest_id = ? AND tm.user_id = ?", contestID, userID).
		First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}
