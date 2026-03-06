package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(challenge *model.Challenge) error {
	return r.db.Create(challenge).Error
}

func (r *Repository) FindByID(id int64) (*model.Challenge, error) {
	var challenge model.Challenge
	err := r.db.Where("id = ?", id).First(&challenge).Error
	return &challenge, err
}

func (r *Repository) Update(challenge *model.Challenge) error {
	return r.db.Save(challenge).Error
}

func (r *Repository) Delete(id int64) error {
	return r.db.Delete(&model.Challenge{}, id).Error
}

func (r *Repository) List(query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	var challenges []*model.Challenge
	var total int64

	db := r.db.Model(&model.Challenge{})

	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Difficulty != "" {
		db = db.Where("difficulty = ?", query.Difficulty)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size
	err := db.Offset(offset).Limit(size).Order("created_at DESC").Find(&challenges).Error
	return challenges, total, err
}

func (r *Repository) HasRunningInstances(challengeID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Instance{}).
		Where("challenge_id = ? AND status IN (?)", challengeID, []string{"creating", "running"}).
		Count(&count).Error
	return count > 0, err
}

// ListPublished 查询已发布的靶场列表（学员视图）
func (r *Repository) ListPublished(query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	var challenges []*model.Challenge
	var total int64

	db := r.db.Model(&model.Challenge{}).Where("status = ?", model.ChallengeStatusPublished)

	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Difficulty != "" {
		db = db.Where("difficulty = ?", query.Difficulty)
	}
	if query.Keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch query.SortBy {
	case "difficulty":
		db = db.Order("difficulty ASC, created_at DESC")
	default:
		db = db.Order("created_at DESC")
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	offset := (page - 1) * size
	err := db.Offset(offset).Limit(size).Find(&challenges).Error
	return challenges, total, err
}

// GetSolvedStatus 获取用户是否已完成靶场
func (r *Repository) GetSolvedStatus(userID, challengeID int64) (bool, error) {
	var count int64
	err := r.db.Table("submissions").
		Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
		Count(&count).Error
	return count > 0, err
}

// GetSolvedCount 获取靶场完成人数
func (r *Repository) GetSolvedCount(challengeID int64) (int64, error) {
	var count int64
	err := r.db.Table("submissions").
		Where("challenge_id = ? AND is_correct = ?", challengeID, true).
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetTotalAttempts 获取靶场总尝试次数
func (r *Repository) GetTotalAttempts(challengeID int64) (int64, error) {
	var count int64
	err := r.db.Table("submissions").
		Where("challenge_id = ?", challengeID).
		Count(&count).Error
	return count, err
}

// FindPublishedWithTags 查询匹配标签的已发布靶场（用于推荐）
func (r *Repository) FindPublishedWithTags(limit int, tagIDs []int64, excludeSolved []int64) ([]*model.Challenge, error) {
	var challenges []*model.Challenge

	db := r.db.Model(&model.Challenge{}).
		Where("status = ?", model.ChallengeStatusPublished)

	if len(tagIDs) > 0 {
		db = db.Joins("JOIN challenge_tags ON challenges.id = challenge_tags.challenge_id").
			Where("challenge_tags.tag_id IN ?", tagIDs)
	}

	if len(excludeSolved) > 0 {
		db = db.Where("challenges.id NOT IN ?", excludeSolved)
	}

	err := db.Order("difficulty ASC, points ASC").Limit(limit).Find(&challenges).Error
	return challenges, err
}

// FindTagsByDimensions 根据维度查询标签 ID
func (r *Repository) FindTagsByDimensions(dimensions []string) ([]int64, error) {
	var tagIDs []int64
	err := r.db.Model(&model.Tag{}).
		Where("dimension IN ?", dimensions).
		Pluck("id", &tagIDs).Error
	return tagIDs, err
}
