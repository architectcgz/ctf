package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

type ScoreboardService struct {
	repo      Repository
	redis     *redislib.Client
	logger    *zap.Logger
	baseScore float64
	minScore  float64
	decay     float64
}

func NewScoreboardService(repo Repository, redis *redislib.Client, cfg *config.ContestConfig, logger *zap.Logger) *ScoreboardService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScoreboardService{
		repo:      repo,
		redis:     redis,
		logger:    logger,
		baseScore: cfg.BaseScore,
		minScore:  cfg.MinScore,
		decay:     cfg.Decay,
	}
}

func (s *ScoreboardService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
	key := rediskeys.RankContestTeamKey(contestID)
	return s.redis.ZIncrBy(ctx, key, points, TeamIDToMember(teamID)).Err()
}

func (s *ScoreboardService) RebuildScoreboard(ctx context.Context, contestID int64) error {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	key := rediskeys.RankContestTeamKey(contestID)
	pipe := s.redis.TxPipeline()
	pipe.Del(ctx, key)

	entries := make([]redislib.Z, 0, len(teams))
	for _, team := range teams {
		if team == nil || team.TotalScore <= 0 {
			continue
		}
		entries = append(entries, redislib.Z{
			Score:  float64(team.TotalScore),
			Member: TeamIDToMember(team.ID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, key, entries...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ScoreboardService) GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, false)
}

func (s *ScoreboardService) GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error) {
	return s.getScoreboard(ctx, contestID, page, pageSize, true)
}

func (s *ScoreboardService) getScoreboard(ctx context.Context, contestID int64, page, pageSize int, live bool) (*dto.ScoreboardResp, error) {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == ErrContestNotFound {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	frozen := !live && isFrozenContest(contest, time.Now())
	key := rediskeys.RankContestTeamKey(contestID)
	if frozen {
		key = rediskeys.RankContestFrozenKey(contestID)
		exists, existsErr := s.redis.Exists(ctx, key).Result()
		if existsErr != nil {
			return nil, errcode.ErrInternal.WithCause(existsErr)
		}
		if exists == 0 {
			if snapshotErr := s.createSnapshotFromLive(ctx, contestID); snapshotErr != nil {
				return nil, snapshotErr
			}
		}
	}

	total, err := s.redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	start := int64((page - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	results, err := s.redis.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, 0, len(results))
	for _, item := range results {
		teamIDs = append(teamIDs, MemberToTeamID(item.Member))
	}

	teams, err := s.repo.FindTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	statsMap, err := s.repo.FindScoreboardTeamStats(ctx, contestID, contest.Mode, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	teamMap := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	items := make([]*dto.ScoreboardItem, 0, len(results))
	for idx, item := range results {
		teamID := teamIDs[idx]
		team := teamMap[teamID]
		stats := statsMap[teamID]
		items = append(items, &dto.ScoreboardItem{
			Rank:             int(start) + idx + 1,
			TeamID:           teamID,
			Score:            item.Score,
			TeamName:         teamName(team),
			SolvedCount:      stats.SolvedCount,
			LastSubmissionAt: stats.LastSubmissionAt,
		})
	}

	return &dto.ScoreboardResp{
		Contest: &dto.ScoreboardContestInfo{
			ID:        contest.ID,
			Title:     contest.Title,
			Status:    contest.Status,
			StartedAt: contest.StartTime,
			EndsAt:    contest.EndTime,
		},
		Scoreboard: &dto.ScoreboardPage{
			List:     items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Frozen: frozen,
	}, nil
}

func (s *ScoreboardService) GetTeamRank(ctx context.Context, contestID, teamID int64) (*dto.TeamRankResp, error) {
	key := rediskeys.RankContestTeamKey(contestID)
	score, err := s.redis.ZScore(ctx, key, TeamIDToMember(teamID)).Result()
	if err != nil {
		if err == redislib.Nil {
			return &dto.TeamRankResp{TeamID: teamID, Rank: 0, Score: 0}, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rank, err := s.redis.ZRevRank(ctx, key, TeamIDToMember(teamID)).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TeamRankResp{
		TeamID: teamID,
		Rank:   int(rank) + 1,
		Score:  score,
	}, nil
}

func (s *ScoreboardService) CalculateDynamicScore(solveCount int) float64 {
	return float64(s.CalculateDynamicScoreWithBase(s.baseScore, int64(solveCount)))
}

func (s *ScoreboardService) CalculateDynamicScoreWithBase(baseScore float64, solveCount int64) int {
	if baseScore <= 0 {
		baseScore = s.baseScore
	}
	return calculateDynamicScore(baseScore, s.minScore, s.decay, solveCount)
}

func (s *ScoreboardService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
		return errcode.ErrContestEnded
	}

	freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime
	if !now.Before(freezeTime) {
		contest.Status = model.ContestStatusFrozen
		if err := s.createSnapshotFromLive(ctx, contestID); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != model.ContestStatusFrozen {
		return errcode.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	if contest.Status == model.ContestStatusFrozen && time.Now().Before(contest.EndTime) {
		contest.Status = model.ContestStatusRunning
	}
	if err := s.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardService) createSnapshotFromLive(ctx context.Context, contestID int64) error {
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func isFrozenContest(contest *model.Contest, now time.Time) bool {
	if contest.Status == model.ContestStatusFrozen {
		return true
	}
	if contest.FreezeTime == nil {
		return false
	}
	return !now.Before(*contest.FreezeTime) && now.Before(contest.EndTime)
}

func teamName(team *model.Team) string {
	if team == nil {
		return ""
	}
	return team.Name
}

func TeamIDToMember(teamID int64) string {
	return strconv.FormatInt(teamID, 10)
}

func MemberToTeamID(member any) int64 {
	switch value := member.(type) {
	case string:
		id, _ := strconv.ParseInt(value, 10, 64)
		return id
	case []byte:
		id, _ := strconv.ParseInt(string(value), 10, 64)
		return id
	default:
		id, _ := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		return id
	}
}
