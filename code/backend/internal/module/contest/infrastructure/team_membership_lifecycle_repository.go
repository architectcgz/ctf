package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *TeamRepository) CreateWithMember(ctx context.Context, team *contestentity.Team, captainID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		member := &contestentity.TeamMember{
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
		var team contestentity.Team
		if err := tx.Where("id = ?", id).First(&team).Error; err != nil {
			return err
		}

		var userIDs []int64
		if err := tx.Model(&contestentity.TeamMember{}).Where("team_id = ?", id).Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := tx.Where("team_id = ?", id).Delete(&contestentity.TeamMember{}).Error; err != nil {
			return err
		}
		if len(userIDs) > 0 {
			if err := tx.Model(&contestentity.ContestRegistration{}).
				Where("contest_id = ? AND user_id IN ?", team.ContestID, userIDs).
				Updates(map[string]any{
					"team_id":    nil,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&contestentity.Team{}, id).Error
	})
}
