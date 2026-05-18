package infrastructure

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *AWDRepository) FindTeamsByContest(ctx context.Context, contestID int64) ([]*contestentity.Team, error) {
	var teams []*contestentity.Team
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Find(&teams).Error
	return teams, err
}

func (r *AWDRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error) {
	var registration contestentity.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *AWDRepository) FindContestTeamByMember(ctx context.Context, contestID, userID int64) (*contestentity.Team, error) {
	var team contestentity.Team
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
