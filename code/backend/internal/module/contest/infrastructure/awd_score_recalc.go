package infrastructure

import (
	"context"

	"gorm.io/gorm"
)

func RecalculateAWDContestTeamScores(ctx context.Context, db *gorm.DB, contestID int64) error {
	if db == nil || contestID <= 0 {
		return nil
	}

	teams, err := loadAWDContestTeams(ctx, db, contestID)
	if err != nil {
		return err
	}
	if len(teams) == 0 {
		return nil
	}

	serviceRows, err := loadAWDServiceScoreRows(ctx, db, contestID)
	if err != nil {
		return err
	}

	attackRows, err := loadAWDAttackScoreRows(ctx, db, contestID)
	if err != nil {
		return err
	}

	defenseMap := accumulateAWDDefenseScores(serviceRows)
	attackMap := accumulateAWDAttackScores(attackRows)

	return applyAWDContestTeamScores(ctx, db, teams, defenseMap, attackMap)
}

func SyncAWDContestScores(ctx context.Context, db *gorm.DB, redis redisScoreboardCache, contestID int64) error {
	if err := RecalculateAWDContestTeamScores(ctx, db, contestID); err != nil {
		return err
	}
	return RebuildContestScoreboardCache(ctx, db, redis, contestID)
}
