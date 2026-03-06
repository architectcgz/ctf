package contest

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/pkg/redis"
	"errors"
	"math"
	"strconv"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ScoreboardService struct {
	repo      *Repository
	redis     *redislib.Client
	logger    *zap.Logger
	baseScore float64
	minScore  float64
	decay     float64
}

func NewScoreboardService(repo *Repository, redis *redislib.Client, cfg *config.ContestConfig, logger *zap.Logger) *ScoreboardService {
	return &ScoreboardService{
		repo:      repo,
		redis:     redis,
		logger:    logger,
		baseScore: cfg.BaseScore,
		minScore:  cfg.MinScore,
		decay:     cfg.Decay,
	}
}

// UpdateScore 更新队伍分数
func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
	freezeFlagKey := redis.ContestFreezeFlagKey(contestID)
	isFrozen := s.redis.Exists(ctx, freezeFlagKey).Val() > 0
	if isFrozen {
		s.logger.Warn("attempt to update score during freeze period",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
		)
		return errors.New("排行榜已冻结，无法更新分数")
	}

	s.logger.Info("updating team score",
		zap.Int64("contest_id", contestID),
		zap.Int64("team_id", teamID),
		zap.Float64("points", points),
	)

	key := redis.RankContestTeamKey(contestID)
	if err := s.redis.ZIncrBy(ctx, key, points, teamIDToMember(teamID)).Err(); err != nil {
		s.logger.Error("failed to update score",
			zap.Int64("contest_id", contestID),
			zap.Int64("team_id", teamID),
			zap.Error(err),
		)
		return err
	}

	return nil
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

	teamIDs := make([]int64, len(results))
	for i, z := range results {
		teamIDs[i] = memberToTeamID(z.Member.(string))
	}

	teams, err := s.repo.FindTeamsByIDs(teamIDs)
	if err != nil {
		s.logger.Error("failed to fetch teams", zap.Error(err))
		return nil, err
	}

	teamMap := make(map[int64]*model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	items := make([]*dto.ScoreboardItem, 0, len(results))
	for i, z := range results {
		teamID := teamIDs[i]
		team := teamMap[teamID]
		if team == nil {
			s.logger.Warn("team not found", zap.Int64("team_id", teamID))
		}
		items = append(items, toScoreboardItem(team, z.Score, i+1))
	}

	return &dto.ScoreboardResp{
		ContestID:  contestID,
		IsFrozen:   isFrozen,
		FreezeTime: contest.FreezeTime,
		Items:      items,
	}, nil
}

func toScoreboardItem(team *model.Team, score float64, rank int) *dto.ScoreboardItem {
	item := &dto.ScoreboardItem{
		Score: score,
		Rank:  rank,
	}
	if team != nil {
		item.TeamID = team.ID
		item.TeamName = team.Name
	}
	return item
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

// CreateSnapshot 创建排行榜快照
func (s *ScoreboardService) CreateSnapshot(ctx context.Context, srcKey, dstKey string) error {
	return s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err()
}

// FreezeScoreboard 冻结排行榜
func (s *ScoreboardService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(contestID)
	if err != nil {
		return err
	}

	if contest.Status == model.ContestStatusFinished {
		return errors.New("竞赛已结束，无法冻结")
	}

	if time.Now().After(contest.EndTime) {
		return errors.New("竞赛已结束")
	}

	freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime

	freezeFlagKey := redis.ContestFreezeFlagKey(contestID)
	s.redis.Set(ctx, freezeFlagKey, "1", 0)

	srcKey := redis.RankContestTeamKey(contestID)
	dstKey := redis.RankContestFrozenKey(contestID)
	if err := s.CreateSnapshot(ctx, srcKey, dstKey); err != nil {
		s.logger.Error("failed to create scoreboard snapshot",
			zap.Int64("contest_id", contestID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("scoreboard frozen",
		zap.Int64("contest_id", contestID),
		zap.Time("freeze_time", freezeTime),
	)

	return s.repo.Update(contest)
}

// UnfreezeScoreboard 解冻排行榜
func (s *ScoreboardService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(contestID)
	if err != nil {
		return err
	}

	if contest.Status == model.ContestStatusFinished {
		return errors.New("竞赛已结束，无法解冻")
	}

	if contest.FreezeTime == nil {
		return errors.New("排行榜未冻结")
	}

	contest.FreezeTime = nil

	freezeFlagKey := redis.ContestFreezeFlagKey(contestID)
	s.redis.Del(ctx, freezeFlagKey)

	dstKey := redis.RankContestFrozenKey(contestID)
	s.redis.Del(ctx, dstKey)

	s.logger.Info("scoreboard unfrozen", zap.Int64("contest_id", contestID))

	return s.repo.Update(contest)
}

func teamIDToMember(teamID int64) string {
	return strconv.FormatInt(teamID, 10)
}

func memberToTeamID(member string) int64 {
	id, _ := strconv.ParseInt(member, 10, 64)
	return id
}
