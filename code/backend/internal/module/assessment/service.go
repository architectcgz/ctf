package assessment

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

// CalculateSkillProfile 计算用户能力画像
func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
	// 查询各维度总分和用户已得分
	type DimensionScore struct {
		Dimension  string
		TotalScore int
		UserScore  int
	}

	var scores []DimensionScore
	err := s.db.Raw(`
		SELECT
			c.category as dimension,
			SUM(c.points) as total_score,
			COALESCE(SUM(CASE WHEN s.is_correct = 1 THEN c.points ELSE 0 END), 0) as user_score
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
		WHERE c.status = 'published'
		GROUP BY c.category
	`, userID).Scan(&scores).Error

	if err != nil {
		return nil, err
	}

	dimensions := make([]*dto.SkillDimension, 0, len(scores))
	profiles := make([]*model.SkillProfile, 0, len(scores))
	now := time.Now()

	for _, score := range scores {
		var rate float64
		if score.TotalScore > 0 {
			rate = float64(score.UserScore) / float64(score.TotalScore)
		}

		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: score.Dimension,
			Score:     rate,
		})

		profiles = append(profiles, &model.SkillProfile{
			UserID:    userID,
			Dimension: score.Dimension,
			Score:     rate,
			UpdatedAt: now,
		})
	}

	// 保存到数据库
	if err := s.repo.BatchUpsert(profiles); err != nil {
		return nil, err
	}

	return dimensions, nil
}

// GetSkillProfile 获取用户能力画像
func (s *Service) GetSkillProfile(userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	dimensions := make([]*dto.SkillDimension, 0, len(profiles))
	var latestUpdate time.Time

	for _, p := range profiles {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: p.Dimension,
			Score:     p.Score,
		})
		if p.UpdatedAt.After(latestUpdate) {
			latestUpdate = p.UpdatedAt
		}
	}

	return &dto.SkillProfileResp{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  latestUpdate.Format(time.RFC3339),
	}, nil
}
