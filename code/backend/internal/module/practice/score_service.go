package practice

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 难度权重映射
var difficultyWeights = map[string]float64{
	model.ChallengeDifficultyBeginner: 1.0,
	model.ChallengeDifficultyEasy:     1.2,
	model.ChallengeDifficultyMedium:   1.5,
	model.ChallengeDifficultyHard:     2.0,
	model.ChallengeDifficultyInsane:   3.0,
}

// ScoreService 计分服务
type ScoreService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

func NewScoreService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *ScoreService {
	return &ScoreService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CalculateScore 计算题目得分
func (s *ScoreService) CalculateScore(challengeID int64) int {
	var challenge model.Challenge
	if err := s.db.Select("points, difficulty").Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		s.logger.Error("查询题目失败", zap.Int64("challengeID", challengeID), zap.Error(err))
		return 0
	}

	weight := difficultyWeights[challenge.Difficulty]
	if weight == 0 {
		weight = 1.0
	}

	return int(float64(challenge.Points) * weight)
}

// UpdateUserScore 更新用户总分
func (s *ScoreService) UpdateUserScore(userID int64) error {
	ctx := context.Background()

	// 查询用户已解决的题目
	var submissions []model.Submission
	err := s.db.Where("user_id = ? AND is_correct = ?", userID, true).
		Select("DISTINCT challenge_id").
		Find(&submissions).Error
	if err != nil {
		return err
	}

	// 计算总分
	totalScore := 0
	for _, sub := range submissions {
		totalScore += s.CalculateScore(sub.ChallengeID)
	}

	// 更新数据库
	now := time.Now()
	err = s.db.Exec(`
		INSERT INTO user_scores (user_id, total_score, solved_count, rank, updated_at)
		VALUES (?, ?, ?, 0, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			total_score = EXCLUDED.total_score,
			solved_count = EXCLUDED.solved_count,
			updated_at = EXCLUDED.updated_at
	`, userID, totalScore, len(submissions), now).Error
	if err != nil {
		return err
	}

	// 更新 Redis 缓存
	cacheKey := fmt.Sprintf("ctf:score:user:%d", userID)
	s.redis.Set(ctx, cacheKey, totalScore, 5*time.Minute)

	// 更新排行榜
	s.redis.ZAdd(ctx, "ctf:ranking", redis.Z{
		Score:  float64(totalScore),
		Member: userID,
	})

	return nil
}

// GetUserScore 获取用户得分信息
func (s *ScoreService) GetUserScore(userID int64) (*dto.UserScoreInfo, error) {
	var userScore model.UserScore
	err := s.db.Where("user_id = ?", userID).First(&userScore).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.UserScoreInfo{
				UserID:      userID,
				TotalScore:  0,
				SolvedCount: 0,
				Rank:        0,
			}, nil
		}
		return nil, err
	}

	var user model.User
	s.db.Select("username").Where("id = ?", userID).First(&user)

	return &dto.UserScoreInfo{
		UserID:      userScore.UserID,
		Username:    user.Username,
		TotalScore:  userScore.TotalScore,
		SolvedCount: userScore.SolvedCount,
		Rank:        userScore.Rank,
	}, nil
}

// GetRanking 获取排行榜
func (s *ScoreService) GetRanking(limit int) ([]*dto.RankingItem, error) {
	ctx := context.Background()

	// 尝试从 Redis 获取
	members, err := s.redis.ZRevRangeWithScores(ctx, "ctf:ranking", 0, int64(limit-1)).Result()
	if err == nil && len(members) > 0 {
		result := make([]*dto.RankingItem, 0, len(members))
		for i, member := range members {
			userID := member.Member.(string)
			var user model.User
			s.db.Select("id, username").Where("id = ?", userID).First(&user)

			result = append(result, &dto.RankingItem{
				Rank:       i + 1,
				UserID:     user.ID,
				Username:   user.Username,
				TotalScore: int(member.Score),
			})
		}
		return result, nil
	}

	// Redis 失败，从数据库查询
	var scores []model.UserScore
	err = s.db.Order("total_score DESC, updated_at ASC").Limit(limit).Find(&scores).Error
	if err != nil {
		return nil, err
	}

	result := make([]*dto.RankingItem, 0, len(scores))
	for i, score := range scores {
		var user model.User
		s.db.Select("username").Where("id = ?", score.UserID).First(&user)

		result = append(result, &dto.RankingItem{
			Rank:       i + 1,
			UserID:     score.UserID,
			Username:   user.Username,
			TotalScore: score.TotalScore,
			SolvedCount: score.SolvedCount,
		})
	}

	return result, nil
}
