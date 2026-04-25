package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func (r *TeamRepository) CreateWithMember(ctx context.Context, team *model.Team, captainID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *TeamRepository) DeleteWithMembers(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
