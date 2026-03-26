package infrastructure

import "ctf-platform/internal/model"

func (r *TeamRepository) FindByID(id int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Where("id = ?", id).First(&team).Error
	return &team, err
}

func (r *TeamRepository) FindContestRegistration(contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	err := r.db.Where("contest_id = ? AND user_id = ?", contestID, userID).First(&registration).Error
	return &registration, err
}

func (r *TeamRepository) GetMembers(teamID int64) ([]*model.TeamMember, error) {
	var members []*model.TeamMember
	err := r.db.Where("team_id = ?", teamID).Order("joined_at ASC").Find(&members).Error
	return members, err
}

func (r *TeamRepository) GetMemberCount(teamID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

func (r *TeamRepository) FindUserTeamInContest(userID, contestID int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.contest_id = ? AND teams.deleted_at IS NULL", userID, contestID).
		First(&team).Error
	return &team, err
}

func (r *TeamRepository) ListByContest(contestID int64) ([]*model.Team, error) {
	var teams []*model.Team
	err := r.db.Where("contest_id = ?", contestID).Order("created_at DESC").Find(&teams).Error
	return teams, err
}

func (r *TeamRepository) GetMemberCountBatch(teamIDs []int64) (map[int64]int, error) {
	type result struct {
		TeamID int64
		Count  int
	}
	var results []result
	err := r.db.Model(&model.TeamMember{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Scan(&results).Error

	countMap := make(map[int64]int)
	for _, item := range results {
		countMap[item.TeamID] = item.Count
	}
	return countMap, err
}

func (r *TeamRepository) FindUsersByIDs(ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}

	var users []*model.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}
