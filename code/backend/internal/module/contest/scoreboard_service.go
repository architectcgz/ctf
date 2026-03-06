package contest

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/pkg/redis"
	"errors"
	"math"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ScoreboardService struct {
	repo      *Repository
	redis     *redislib.Client
	baseScore float64
	minScore  float64
	decay     float64
}

func NewScoreboardService(repo *Repository, redis *redislib.Client) *ScoreboardService {
	return &ScoreboardService{
		repo:      repo,
		redis:     redis,
		baseScore: 1000,
		minScore:  100,
		decay:     0.9,
	}
}

// UpdateScore 更新队伍分数
func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
	key := redis.RankContestTeamKey(contestID)
	return s.redis.ZIncrBy(ctx, key, points, teamIDToMember(teamID)).Err()
}

// GetScoreboard 获取排行榜
func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64) (*dto.ScoreboardResp, error) {
	contest, err := s.repo.FindByID(contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("竞赛不存在")
		}
		return nil, err
	}

	isFrozen := contest.FreezeTime != nil && time.Now().After(*contest.FreezeTime)

	var key string
	if isFrozen {
		key = redis.RankContestFrozenKey(contestID)
		if s.redis.Exists(ctx, key).Val() == 0 {
			key = redis.RankContestTeamKey(contestID)
		}
	} else {
		key = redis.RankContestTeamKey(contestID)
	}

	results, err := s.redis.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	items := make([]*dto.ScoreboardItem, 0, len(results))
	for i, z := range results {
		teamID := memberToTeamID(z.Member.(string))
		team, _ := s.repo.FindTeamByID(teamID)

		teamName := ""
		if team != nil {
			teamName = team.Name
		}

		items = append(items, &dto.ScoreboardItem{
			TeamID:   teamID,
			TeamName: teamName,
			Score:    z.Score,
			Rank:     i + 1,
		})
	}

	return &dto.ScoreboardResp{
		ContestID:  contestID,
		IsFrozen:   isFrozen,
		FreezeTime: contest.FreezeTime,
		Items:      items,
	}, nil
}

// GetTeamRank 获取队伍排名
func (s *ScoreboardService) GetTeamRank(ctx context.Context, contestID, teamID int64) (*dto.TeamRankResp, error) {
	key := redis.RankContestTeamKey(contestID)

	score, err := s.redis.ZScore(ctx, key, teamIDToMember(teamID)).Result()
	if err != nil {
		if err == redislib.Nil {
			return &dto.TeamRankResp{TeamID: teamID, Rank: 0, Score: 0}, nil
		}
		return nil, err
	}

	rank, err := s.redis.ZRevRank(ctx, key, teamIDToMember(teamID)).Result()
	if err != nil {
		return nil, err
	}

	return &dto.TeamRankResp{
		TeamID: teamID,
		Rank:   int(rank) + 1,
		Score:  score,
	}, nil
}

// CalculateDynamicScore 计算动态分数
func (s *ScoreboardService) CalculateDynamicScore(solveCount int) float64 {
	score := s.baseScore * math.Pow(s.decay, float64(solveCount))
	return math.Max(s.minScore, score)
}

func teamIDToMember(teamID int64) string {
	return string(rune(teamID))
}

func memberToTeamID(member string) int64 {
	return int64([]rune(member)[0])
}
