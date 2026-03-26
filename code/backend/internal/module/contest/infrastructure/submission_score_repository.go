package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
)

func (r *SubmissionRepository) LockContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	var contestChallenge model.ContestChallenge
	if err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
		return nil, err
	}
	return &contestChallenge, nil
}

func (r *SubmissionRepository) CountCorrectSubmissions(ctx context.Context, contestID, challengeID int64, teamID *int64, userID int64) (int64, error) {
	query := r.dbWithContext(ctx).
		Model(&model.Submission{}).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true)
	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	} else {
		query = query.Where("team_id IS NULL AND user_id = ?", userID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SubmissionRepository) UpdateFirstBlood(ctx context.Context, contestID, challengeID, teamID int64) error {
	return r.dbWithContext(ctx).
		Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Update("first_blood_by", teamID).Error
}

func (r *SubmissionRepository) ListCorrectSubmissions(ctx context.Context, contestID, challengeID int64) ([]model.Submission, error) {
	var submissions []model.Submission
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true).
		Order("submitted_at ASC, id ASC").
		Find(&submissions).Error; err != nil {
		return nil, err
	}
	return submissions, nil
}

func (r *SubmissionRepository) AddTeamScore(ctx context.Context, teamID int64, delta int, lastSolveAt *time.Time) error {
	updates := map[string]any{
		"total_score": gorm.Expr("total_score + ?", delta),
	}
	if lastSolveAt != nil {
		updates["last_solve_at"] = *lastSolveAt
	}
	return r.dbWithContext(ctx).
		Model(&model.Team{}).
		Where("id = ?", teamID).
		Updates(updates).Error
}
