package infrastructure

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *TeamRepository) CreateWithMember(team *model.Team, captainID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		member := &model.TeamMember{
			ContestID: team.ContestID,
			TeamID:    team.ID,
			UserID:    captainID,
			JoinedAt:  time.Now(),
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, team.ContestID, captainID, &team.ID)
	})
}

func (r *TeamRepository) DeleteWithMembers(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Where("id = ?", id).First(&team).Error; err != nil {
			return err
		}

		var userIDs []int64
		if err := tx.Model(&model.TeamMember{}).Where("team_id = ?", id).Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ?", id).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		if len(userIDs) > 0 {
			if err := tx.Model(&model.ContestRegistration{}).
				Where("contest_id = ? AND user_id IN ?", team.ContestID, userIDs).
				Updates(map[string]any{
					"team_id":    nil,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&model.Team{}, id).Error
	})
}

func (r *TeamRepository) AddMemberWithLock(contestID, teamID, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND contest_id = ?", teamID, contestID).
			First(&team).Error; err != nil {
			return err
		}

		var existingCount int64
		if err := tx.Model(&model.TeamMember{}).
			Where("contest_id = ? AND user_id = ?", contestID, userID).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return contestdomain.ErrAlreadyJoinedContest
		}

		var memberCount int64
		if err := tx.Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&memberCount).Error; err != nil {
			return err
		}
		if memberCount >= int64(team.MaxMembers) {
			return contestdomain.ErrTeamFull
		}

		member := &model.TeamMember{
			ContestID: contestID,
			TeamID:    teamID,
			UserID:    userID,
			JoinedAt:  time.Now(),
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, contestID, userID, &teamID)
	})
}

func (r *TeamRepository) RemoveMember(teamID, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var team model.Team
		if err := tx.Where("id = ?", teamID).First(&team).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&model.TeamMember{}).Error; err != nil {
			return err
		}
		return bindContestRegistrationTeam(tx, team.ContestID, userID, nil)
	})
}

func bindContestRegistrationTeam(tx *gorm.DB, contestID, userID int64, teamID *int64) error {
	result := tx.Model(&model.ContestRegistration{}).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Updates(map[string]any{
			"team_id":    teamID,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
