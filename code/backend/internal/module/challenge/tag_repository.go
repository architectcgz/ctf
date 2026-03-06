package challenge

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

func (r *TagRepository) FindByID(id int64) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) List(tagType string) ([]*model.Tag, error) {
	var tags []*model.Tag
	query := r.db.Model(&model.Tag{})
	if tagType != "" {
		query = query.Where("type = ?", tagType)
	}
	err := query.Order("type, name").Find(&tags).Error
	return tags, err
}

func (r *TagRepository) FindByIDs(ids []int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}

func (r *TagRepository) AttachToChallenge(challengeID, tagID int64) error {
	ct := &model.ChallengeTag{
		ChallengeID: challengeID,
		TagID:       tagID,
	}
	return r.db.Create(ct).Error
}

func (r *TagRepository) DetachFromChallenge(challengeID, tagID int64) error {
	return r.db.Where("challenge_id = ? AND tag_id = ?", challengeID, tagID).
		Delete(&model.ChallengeTag{}).Error
}

func (r *TagRepository) FindByChallengeID(challengeID int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.Table("tags").
		Joins("JOIN challenge_tags ON tags.id = challenge_tags.tag_id").
		Where("challenge_tags.challenge_id = ?", challengeID).
		Order("tags.type, tags.name").
		Find(&tags).Error
	return tags, err
}

func (r *TagRepository) AttachTagsInTx(challengeID int64, tagIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, tid := range tagIDs {
			ct := &model.ChallengeTag{
				ChallengeID: challengeID,
				TagID:       tid,
			}
			if err := tx.Create(ct).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *TagRepository) Delete(id int64) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

func (r *TagRepository) CountChallengesByTagID(tagID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.ChallengeTag{}).Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}
