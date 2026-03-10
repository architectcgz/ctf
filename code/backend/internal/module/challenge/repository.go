package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *Repository) CreateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(challenge).Error; err != nil {
			return err
		}
		if len(hints) == 0 {
			return nil
		}
		for _, hint := range hints {
			hint.ChallengeID = challenge.ID
		}
		return tx.Create(&hints).Error
	})
}

func (r *Repository) FindByID(id int64) (*model.Challenge, error) {
	var challenge model.Challenge
	err := r.db.Where("id = ?", id).First(&challenge).Error
	return &challenge, err
}

func (r *Repository) Update(challenge *model.Challenge) error {
	return r.db.Save(challenge).Error
}

func (r *Repository) UpdateWithHints(challenge *model.Challenge, hints []*model.ChallengeHint, replaceHints bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(challenge).Error; err != nil {
			return err
		}
		if !replaceHints {
			return nil
		}
		if err := tx.Where("challenge_id = ?", challenge.ID).Delete(&model.ChallengeHint{}).Error; err != nil {
			return err
		}
		if len(hints) == 0 {
			return nil
		}
		for _, hint := range hints {
			hint.ChallengeID = challenge.ID
		}
		return tx.Create(&hints).Error
	})
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

func (r *Repository) CountByImageID(imageID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Challenge{}).
		Where("image_id = ?", imageID).
		Count(&count).Error
	return count, err
}

func (r *Repository) ListHintsByChallengeID(challengeID int64) ([]*model.ChallengeHint, error) {
	var hints []*model.ChallengeHint
	err := r.db.Where("challenge_id = ?", challengeID).Order("level ASC, id ASC").Find(&hints).Error
	return hints, err
}

func (r *Repository) FindHintByLevel(challengeID int64, level int) (*model.ChallengeHint, error) {
	var hint model.ChallengeHint
	err := r.db.Where("challenge_id = ? AND level = ?", challengeID, level).First(&hint).Error
	return &hint, err
}

func (r *Repository) GetUnlockedHintIDs(userID, challengeID int64) (map[int64]bool, error) {
	var unlocks []model.ChallengeHintUnlock
	err := r.db.
		Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		Find(&unlocks).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]bool, len(unlocks))
	for _, unlock := range unlocks {
		result[unlock.ChallengeHintID] = true
	}
	return result, nil
}

func (r *Repository) CreateHintUnlock(unlock *model.ChallengeHintUnlock) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "challenge_hint_id"}},
		DoNothing: true,
	}).Create(unlock).Error
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
		// GORM 会自动转义参数，防止 SQL 注入
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

	db = r.applyPagination(db, query.Page, query.Size)
	err := db.Find(&challenges).Error
	return challenges, total, err
}

// applyPagination 应用分页逻辑
func (r *Repository) applyPagination(db *gorm.DB, page, size int) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	return db.Offset(offset).Limit(size)
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

// BatchGetSolvedStatus 批量获取用户完成状态
func (r *Repository) BatchGetSolvedStatus(userID int64, challengeIDs []int64) (map[int64]bool, error) {
	if userID == 0 || len(challengeIDs) == 0 {
		return make(map[int64]bool), nil
	}

	var results []struct {
		ChallengeID int64
	}
	err := r.db.Table("submissions").
		Select("DISTINCT challenge_id").
		Where("user_id = ? AND challenge_id IN ? AND is_correct = ?", userID, challengeIDs, true).
		Find(&results).Error

	statusMap := make(map[int64]bool)
	for _, r := range results {
		statusMap[r.ChallengeID] = true
	}
	return statusMap, err
}

// BatchGetSolvedCount 批量获取靶场完成人数
func (r *Repository) BatchGetSolvedCount(challengeIDs []int64) (map[int64]int64, error) {
	if len(challengeIDs) == 0 {
		return make(map[int64]int64), nil
	}

	var results []struct {
		ChallengeID int64
		Count       int64
	}
	err := r.db.Table("submissions").
		Select("challenge_id, COUNT(DISTINCT user_id) as count").
		Where("challenge_id IN ? AND is_correct = ?", challengeIDs, true).
		Group("challenge_id").
		Find(&results).Error

	countMap := make(map[int64]int64)
	for _, r := range results {
		countMap[r.ChallengeID] = r.Count
	}
	return countMap, err
}

// BatchGetTotalAttempts 批量获取靶场尝试次数
func (r *Repository) BatchGetTotalAttempts(challengeIDs []int64) (map[int64]int64, error) {
	if len(challengeIDs) == 0 {
		return make(map[int64]int64), nil
	}

	var results []struct {
		ChallengeID int64
		Count       int64
	}
	err := r.db.Table("submissions").
		Select("challenge_id, COUNT(*) as count").
		Where("challenge_id IN ?", challengeIDs).
		Group("challenge_id").
		Find(&results).Error

	countMap := make(map[int64]int64)
	for _, r := range results {
		countMap[r.ChallengeID] = r.Count
	}
	return countMap, err
}

func (r *Repository) FindPublishedForRecommendation(limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error) {
	if len(dimensions) == 0 || limit <= 0 {
		return []*model.Challenge{}, nil
	}

	normalized := make([]string, 0, len(dimensions))
	seen := make(map[string]struct{}, len(dimensions))
	for _, dimension := range dimensions {
		key := strings.ToLower(strings.TrimSpace(dimension))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	if len(normalized) == 0 {
		return []*model.Challenge{}, nil
	}

	var challenges []*model.Challenge
	query := r.db.Model(&model.Challenge{}).
		Distinct("challenges.*").
		Joins("LEFT JOIN challenge_tags ON challenge_tags.challenge_id = challenges.id").
		Joins("LEFT JOIN tags ON tags.id = challenge_tags.tag_id").
		Where("challenges.status = ?", model.ChallengeStatusPublished).
		Where(
			"(LOWER(challenges.category) IN ? OR (tags.type = ? AND LOWER(tags.name) IN ?))",
			normalized,
			model.TagTypeKnowledge,
			normalized,
		)

	if len(excludeSolved) > 0 {
		query = query.Where("challenges.id NOT IN ?", excludeSolved)
	}

	err := query.
		Order(`
			CASE challenges.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'insane' THEN 5
				ELSE 6
			END ASC
		`).
		Order("challenges.points ASC").
		Order("challenges.created_at DESC").
		Limit(limit).
		Find(&challenges).Error
	return challenges, err
}
