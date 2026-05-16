package infrastructure

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func (r *Repository) AddPausedDurationToActiveAWDContests(ctx context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*model.Contest, error) {
	if targetPausedSeconds <= 0 || strings.TrimSpace(recoveryKey) == "" {
		return nil, nil
	}

	var updatedContests []*model.Contest
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var contests []*model.Contest
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"mode = ? AND status IN ? AND deleted_at IS NULL",
				model.ContestModeAWD,
				[]string{model.ContestStatusRunning, model.ContestStatusFrozen},
			).
			Order("id ASC").
			Find(&contests).Error; err != nil {
			return err
		}
		if len(contests) == 0 {
			return nil
		}

		for _, contest := range contests {
			if contest == nil || contestdomain.ContestHasEndedAt(contest, activeAt) {
				continue
			}

			delta := targetPausedSeconds
			if strings.TrimSpace(contest.RuntimeRecoveryKey) == strings.TrimSpace(recoveryKey) {
				delta = targetPausedSeconds - contest.RuntimeRecoveryAppliedSeconds
			}
			if delta > 0 {
				if err := tx.Model(&model.Contest{}).
					Where("id = ?", contest.ID).
					Updates(map[string]any{
						"paused_seconds":                   gorm.Expr("paused_seconds + ?", delta),
						"runtime_recovery_key":             recoveryKey,
						"runtime_recovery_applied_seconds": targetPausedSeconds,
						"updated_at":                       updatedAt.UTC(),
					}).Error; err != nil {
					return err
				}

				contest.PausedSeconds += delta
				contest.UpdatedAt = updatedAt.UTC()
			}
			contest.RuntimeRecoveryKey = recoveryKey
			contest.RuntimeRecoveryAppliedSeconds = targetPausedSeconds
			updatedContests = append(updatedContests, contest)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedContests, nil
}
