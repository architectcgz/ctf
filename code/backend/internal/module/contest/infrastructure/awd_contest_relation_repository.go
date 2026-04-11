package infrastructure

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (r *AWDRepository) ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error) {
	var contests []model.Contest
	query := r.dbWithContext(ctx).
		Where("mode = ?", model.ContestModeAWD).
		Where("status IN ?", []string{
			model.ContestStatusRegistration,
			model.ContestStatusRunning,
			model.ContestStatusFrozen,
			model.ContestStatusEnded,
		}).
		Where("start_time <= ?", now).
		Where("end_time > ?", recentCutoff).
		Order("start_time ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&contests).Error; err != nil {
		return nil, err
	}
	return contests, nil
}

func (r *AWDRepository) ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error) {
	var challenges []model.Challenge
	if err := r.dbWithContext(ctx).
		Table("challenges AS c").
		Select("c.*").
		Joins("JOIN contest_challenges AS cc ON cc.challenge_id = c.id").
		Where("cc.contest_id = ?", contestID).
		Order("c.id ASC").
		Scan(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (r *AWDRepository) ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	var definitions []contestports.AWDServiceDefinition
	if err := r.dbWithContext(ctx).
		Table("contest_challenges AS cc").
		Select(`
			cc.challenge_id AS challenge_id,
			c.flag_prefix AS flag_prefix,
			cc.awd_checker_type AS awd_checker_type,
			cc.awd_checker_config AS awd_checker_config,
			cc.awd_sla_score AS awd_sla_score,
			cc.awd_defense_score AS awd_defense_score
		`).
		Joins("JOIN challenges AS c ON c.id = cc.challenge_id").
		Where("cc.contest_id = ?", contestID).
		Order("cc.challenge_id ASC").
		Scan(&definitions).Error; err != nil {
		return nil, err
	}
	return definitions, nil
}

func (r *AWDRepository) ContestHasChallenge(ctx context.Context, contestID, challengeID int64) (bool, error) {
	var count int64
	if err := r.dbWithContext(ctx).
		Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AWDRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
