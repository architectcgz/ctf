package contest

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type ChallengeRepository struct {
	db *gorm.DB
}

func NewChallengeRepository(db *gorm.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) AddChallenge(cc *model.ContestChallenge) error {
	return r.db.Create(cc).Error
}

func (r *ChallengeRepository) RemoveChallenge(contestID, challengeID int64) error {
	return r.db.Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Delete(&model.ContestChallenge{}).Error
}

func (r *ChallengeRepository) UpdatePoints(contestID, challengeID int64, points int) error {
	return r.db.Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Update("points", points).Error
}

func (r *ChallengeRepository) ListChallenges(contestID int64) ([]*model.ContestChallenge, error) {
	var challenges []*model.ContestChallenge
	err := r.db.Where("contest_id = ?", contestID).
		Order("`order` ASC, created_at ASC").
		Find(&challenges).Error
	return challenges, err
}

func (r *ChallengeRepository) Exists(contestID, challengeID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error
	return count > 0, err
}
