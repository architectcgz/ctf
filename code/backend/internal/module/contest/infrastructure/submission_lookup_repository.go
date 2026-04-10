package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (r *SubmissionRepository) FindRegistration(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *SubmissionRepository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	var contestChallenge model.ContestChallenge
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

func (r *SubmissionRepository) FindActiveSharedProofByHash(ctx context.Context, proofHash string) (*model.SharedProof, error) {
	var proof model.SharedProof
	if err := r.dbWithContext(ctx).
		Where("proof_hash = ? AND status = ?", proofHash, model.SharedProofStatusActive).
		First(&proof).Error; err != nil {
		return nil, err
	}
	return &proof, nil
}
