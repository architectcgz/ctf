package infrastructure

import (
	"context"

	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *TagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.dbWithContext(ctx).Create(tag).Error
}

func (r *TagRepository) List(ctx context.Context, tagType string) ([]*model.Tag, error) {
	var tags []*model.Tag
	query := r.dbWithContext(ctx).Model(&model.Tag{})
	if tagType != "" {
		query = query.Where("type = ?", tagType)
	}
	err := query.Order("type, name").Find(&tags).Error
	return tags, err
}

func (r *TagRepository) FindByIDs(ctx context.Context, ids []int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.dbWithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}

func (r *TagRepository) DetachFromChallenge(ctx context.Context, challengeID, tagID int64) error {
	return r.dbWithContext(ctx).Where("challenge_id = ? AND tag_id = ?", challengeID, tagID).
		Delete(&model.ChallengeTag{}).Error
}

func (r *TagRepository) FindByChallengeID(ctx context.Context, challengeID int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.dbWithContext(ctx).Table("tags").
		Joins("JOIN challenge_tags ON tags.id = challenge_tags.tag_id").
		Where("challenge_tags.challenge_id = ?", challengeID).
		Order("tags.type, tags.name").
		Find(&tags).Error
	return tags, err
}

func (r *TagRepository) AttachTagsInTx(ctx context.Context, challengeID int64, tagIDs []int64) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *TagRepository) DeleteWithContext(ctx context.Context, id int64) error {
	return r.dbWithContext(ctx).Delete(&model.Tag{}, id).Error
}

func (r *TagRepository) CountChallengesByTagID(ctx context.Context, tagID int64) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).Model(&model.ChallengeTag{}).Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}
