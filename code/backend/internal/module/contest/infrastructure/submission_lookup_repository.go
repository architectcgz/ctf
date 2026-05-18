package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *SubmissionRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*contestentity.ContestRegistration, error) {
	var registration contestentity.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *SubmissionRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*contestentity.ContestChallenge, error) {
	var contestChallenge contestentity.ContestChallenge
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
		return nil, err
	}
	return &contestChallenge, nil
}

func (r *SubmissionRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).First(&challenge, challengeID).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
