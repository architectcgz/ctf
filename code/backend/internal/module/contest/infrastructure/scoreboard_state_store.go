package infrastructure

import (
	"context"
	"errors"
	"fmt"

	redislib "github.com/redis/go-redis/v9"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

var _ contestports.ContestScoreboardStateStore = (*ContestScoreboardStateStore)(nil)

type ContestScoreboardStateStore struct {
	cache *redislib.Client
}

func NewContestScoreboardStateStore(cache *redislib.Client) *ContestScoreboardStateStore {
	if cache == nil {
		return nil
	}
	return &ContestScoreboardStateStore{cache: cache}
}

func (s *ContestScoreboardStateStore) HasFrozenScoreboardSnapshot(ctx context.Context, contestID int64) (bool, error) {
	if s == nil || s.cache == nil || contestID <= 0 {
		return false, nil
	}
	exists, err := s.cache.Exists(ctx, rediskeys.RankContestFrozenKey(contestID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *ContestScoreboardStateStore) CreateFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return createFrozenScoreboardSnapshot(ctx, s.cache, contestID)
}

func (s *ContestScoreboardStateStore) ClearFrozenScoreboardSnapshot(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return clearFrozenScoreboardSnapshot(ctx, s.cache, contestID)
}

func (s *ContestScoreboardStateStore) ListLiveScoreboard(ctx context.Context, contestID int64) ([]contestports.ScoreboardMemberScore, error) {
	if s == nil || s.cache == nil {
		return []contestports.ScoreboardMemberScore{}, nil
	}
	return listScoreboardMembers(ctx, s.cache, rediskeys.RankContestTeamKey(contestID))
}

func (s *ContestScoreboardStateStore) ListFrozenScoreboard(ctx context.Context, contestID int64) ([]contestports.ScoreboardMemberScore, error) {
	if s == nil || s.cache == nil {
		return []contestports.ScoreboardMemberScore{}, nil
	}
	return listScoreboardMembers(ctx, s.cache, rediskeys.RankContestFrozenKey(contestID))
}

func (s *ContestScoreboardStateStore) GetLiveTeamRank(ctx context.Context, contestID, teamID int64) (contestports.ScoreboardTeamRank, bool, error) {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 {
		return contestports.ScoreboardTeamRank{}, false, nil
	}

	key := rediskeys.RankContestTeamKey(contestID)
	score, err := s.cache.ZScore(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return contestports.ScoreboardTeamRank{}, false, nil
		}
		return contestports.ScoreboardTeamRank{}, false, err
	}

	rank, err := s.cache.ZRevRank(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		return contestports.ScoreboardTeamRank{}, false, err
	}

	return contestports.ScoreboardTeamRank{
		Rank:  int(rank) + 1,
		Score: score,
	}, true, nil
}

func (s *ContestScoreboardStateStore) IncrementLiveTeamScore(ctx context.Context, contestID, teamID int64, points float64) error {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 {
		return nil
	}
	return s.cache.ZIncrBy(ctx, rediskeys.RankContestTeamKey(contestID), points, contestdomain.TeamIDToMember(teamID)).Err()
}

func (s *ContestScoreboardStateStore) ReplaceLiveScoreboard(ctx context.Context, contestID int64, entries []contestports.ScoreboardTeamScoreEntry) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return replaceLiveScoreboard(ctx, s.cache, contestID, entries)
}

func createFrozenScoreboardSnapshot(ctx context.Context, cache *redislib.Client, contestID int64) error {
	if cache == nil || contestID <= 0 {
		return nil
	}

	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	return cache.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err()
}

func clearFrozenScoreboardSnapshot(ctx context.Context, cache *redislib.Client, contestID int64) error {
	if cache == nil || contestID <= 0 {
		return nil
	}
	return cache.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err()
}

func listScoreboardMembers(ctx context.Context, cache *redislib.Client, key string) ([]contestports.ScoreboardMemberScore, error) {
	if cache == nil || key == "" {
		return []contestports.ScoreboardMemberScore{}, nil
	}

	results, err := cache.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]contestports.ScoreboardMemberScore, 0, len(results))
	for _, item := range results {
		entries = append(entries, contestports.ScoreboardMemberScore{
			Member: scoreboardMemberString(item.Member),
			Score:  item.Score,
		})
	}
	return entries, nil
}

func replaceLiveScoreboard(ctx context.Context, cache *redislib.Client, contestID int64, entries []contestports.ScoreboardTeamScoreEntry) error {
	if cache == nil || contestID <= 0 {
		return nil
	}

	key := rediskeys.RankContestTeamKey(contestID)
	pipe := cache.TxPipeline()
	pipe.Del(ctx, key)

	redisEntries := make([]redislib.Z, 0, len(entries))
	for _, entry := range entries {
		if entry.TeamID <= 0 || entry.Score <= 0 {
			continue
		}
		redisEntries = append(redisEntries, redislib.Z{
			Score:  entry.Score,
			Member: contestdomain.TeamIDToMember(entry.TeamID),
		})
	}
	if len(redisEntries) > 0 {
		pipe.ZAdd(ctx, key, redisEntries...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func scoreboardMemberString(member any) string {
	switch value := member.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}
