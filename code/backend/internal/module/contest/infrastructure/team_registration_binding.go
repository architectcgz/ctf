package infrastructure

import (
	"time"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func bindContestRegistrationTeam(tx *gorm.DB, contestID, userID int64, teamID *int64) error {
	result := tx.Model(&contestentity.ContestRegistration{}).
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
