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
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *TagRepository) Create(tag *model.Tag) error {
	return r.CreateWithContext(context.Background(), tag)
}

func (r *TagRepository) CreateWithContext(ctx context.Context, tag *model.Tag) error {
	return r.dbWithContext(ctx).Create(tag).Error
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
	return r.ListWithContext(context.Background(), tagType)
}

func (r *TagRepository) ListWithContext(ctx context.Context, tagType string) ([]*model.Tag, error) {
	var tags []*model.Tag
	query := r.dbWithContext(ctx).Model(&model.Tag{})
	if tagType != "" {
		query = query.Where("type = ?", tagType)
	}
	err := query.Order("type, name").Find(&tags).Error
	return tags, err
}

func (r *TagRepository) FindByIDs(ids []int64) ([]*model.Tag, error) {
	return r.FindByIDsWithContext(context.Background(), ids)
}

func (r *TagRepository) FindByIDsWithContext(ctx context.Context, ids []int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.dbWithContext(ctx).Where("id IN ?", ids).Find(&tags).Error
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
	return r.DetachFromChallengeWithContext(context.Background(), challengeID, tagID)
}

func (r *TagRepository) DetachFromChallengeWithContext(ctx context.Context, challengeID, tagID int64) error {
	return r.dbWithContext(ctx).Where("challenge_id = ? AND tag_id = ?", challengeID, tagID).
		Delete(&model.ChallengeTag{}).Error
}

func (r *TagRepository) FindByChallengeID(challengeID int64) ([]*model.Tag, error) {
	return r.FindByChallengeIDWithContext(context.Background(), challengeID)
}

func (r *TagRepository) FindByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.dbWithContext(ctx).Table("tags").
		Joins("JOIN challenge_tags ON tags.id = challenge_tags.tag_id").
		Where("challenge_tags.challenge_id = ?", challengeID).
		Order("tags.type, tags.name").
		Find(&tags).Error
	return tags, err
}

func (r *TagRepository) AttachTagsInTx(challengeID int64, tagIDs []int64) error {
	return r.AttachTagsInTxWithContext(context.Background(), challengeID, tagIDs)
}

func (r *TagRepository) AttachTagsInTxWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
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

func (r *TagRepository) Delete(id int64) error {
	return r.DeleteWithContext(context.Background(), id)
}

func (r *TagRepository) DeleteWithContext(ctx context.Context, id int64) error {
	return r.dbWithContext(ctx).Delete(&model.Tag{}, id).Error
}

func (r *TagRepository) CountChallengesByTagID(tagID int64) (int64, error) {
	return r.CountChallengesByTagIDWithContext(context.Background(), tagID)
}

func (r *TagRepository) CountChallengesByTagIDWithContext(ctx context.Context, tagID int64) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).Model(&model.ChallengeTag{}).Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}
