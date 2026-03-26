package infrastructure

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (r *ParticipationRepository) ListSolvedProgress(ctx context.Context, contestID, userID int64) ([]*contestports.ContestParticipationSolvedProgressRow, error) {
	var rows []*contestports.ContestParticipationSolvedProgressRow
	if err := r.dbWithContext(ctx).
		Table("submissions AS s").
		Select("cc.id AS contest_challenge_id, s.submitted_at AS solved_at, s.score AS points_earned").
		Joins("JOIN contest_challenges cc ON cc.contest_id = s.contest_id AND cc.challenge_id = s.challenge_id").
		Where("s.contest_id = ? AND s.user_id = ? AND s.is_correct = ?", contestID, userID, true).
		Order("s.submitted_at ASC, s.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
