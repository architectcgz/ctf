package queries

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

type scoreboardRepoStub struct{}

func (s *scoreboardRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *scoreboardRepoStub) FindTeamsByIDs(context.Context, []int64) ([]*model.Team, error) {
	return nil, nil
}

func (s *scoreboardRepoStub) FindScoreboardTeamStats(context.Context, int64, string, []int64) (map[int64]contestports.ScoreboardTeamStats, error) {
	return nil, nil
}

func TestScoreboardServiceGetScoreboardReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	redisClient := redislib.NewClient(&redislib.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	service := NewScoreboardService(&scoreboardRepoStub{}, redisClient, &config.ContestConfig{}, nil)

	_, err := service.GetScoreboard(context.Background(), 42, 1, 20)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}

type scoreboardQueryRepoStub struct {
	contest  *model.Contest
	teams    map[int64]*model.Team
	statsMap map[int64]contestports.ScoreboardTeamStats
}

func (s *scoreboardQueryRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return s.contest, nil
}

func (s *scoreboardQueryRepoStub) FindTeamsByIDs(_ context.Context, ids []int64) ([]*model.Team, error) {
	teams := make([]*model.Team, 0, len(ids))
	for _, id := range ids {
		if team := s.teams[id]; team != nil {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

func (s *scoreboardQueryRepoStub) FindScoreboardTeamStats(_ context.Context, _ int64, _ string, ids []int64) (map[int64]contestports.ScoreboardTeamStats, error) {
	result := make(map[int64]contestports.ScoreboardTeamStats, len(ids))
	for _, id := range ids {
		if stats, ok := s.statsMap[id]; ok {
			result[id] = stats
		}
	}
	return result, nil
}

func TestScoreboardServiceGetScoreboardSkipsInvalidRedisMembers(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = redisClient.Close()
	}()

	now := time.Now()
	contestID := int64(77)
	key := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), key,
		redislib.Z{Score: 500, Member: contestdomain.TeamIDToMember(11)},
		redislib.Z{Score: 450, Member: "broken-member"},
		redislib.Z{Score: 400, Member: contestdomain.TeamIDToMember(12)},
	).Err(); err != nil {
		t.Fatalf("seed redis scoreboard: %v", err)
	}

	service := NewScoreboardService(&scoreboardQueryRepoStub{
		contest: &model.Contest{
			ID:        contestID,
			Title:     "scoreboard",
			Mode:      model.ContestModeJeopardy,
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Status:    model.ContestStatusRunning,
		},
		teams: map[int64]*model.Team{
			11: {ID: 11, Name: "Alpha"},
			12: {ID: 12, Name: "Beta"},
		},
		statsMap: map[int64]contestports.ScoreboardTeamStats{
			11: {SolvedCount: 5},
			12: {SolvedCount: 4},
		},
	}, redisClient, &config.ContestConfig{}, zap.NewNop())

	resp, err := service.GetScoreboard(context.Background(), contestID, 1, 10)
	if err != nil {
		t.Fatalf("GetScoreboard() error = %v", err)
	}
	if resp.Scoreboard == nil {
		t.Fatalf("expected scoreboard payload, got %+v", resp)
	}
	if len(resp.Scoreboard.List) != 2 {
		t.Fatalf("expected invalid redis member to be skipped, got %+v", resp.Scoreboard.List)
	}
	if resp.Scoreboard.List[0].TeamID != 11 || resp.Scoreboard.List[1].TeamID != 12 {
		t.Fatalf("unexpected scoreboard items: %+v", resp.Scoreboard.List)
	}
	if resp.Scoreboard.Total != 2 {
		t.Fatalf("expected scoreboard total to exclude invalid redis member, got %d", resp.Scoreboard.Total)
	}
}

func TestScoreboardServiceGetScoreboardSkipsMissingTeams(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = redisClient.Close()
	}()

	now := time.Now()
	contestID := int64(78)
	key := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), key,
		redislib.Z{Score: 500, Member: contestdomain.TeamIDToMember(11)},
		redislib.Z{Score: 450, Member: contestdomain.TeamIDToMember(12)},
	).Err(); err != nil {
		t.Fatalf("seed redis scoreboard: %v", err)
	}

	service := NewScoreboardService(&scoreboardQueryRepoStub{
		contest: &model.Contest{
			ID:        contestID,
			Title:     "scoreboard",
			Mode:      model.ContestModeJeopardy,
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Status:    model.ContestStatusRunning,
		},
		teams: map[int64]*model.Team{
			11: {ID: 11, Name: "Alpha"},
		},
		statsMap: map[int64]contestports.ScoreboardTeamStats{
			11: {SolvedCount: 5},
			12: {SolvedCount: 4},
		},
	}, redisClient, &config.ContestConfig{}, zap.NewNop())

	resp, err := service.GetScoreboard(context.Background(), contestID, 1, 10)
	if err != nil {
		t.Fatalf("GetScoreboard() error = %v", err)
	}
	if resp.Scoreboard == nil {
		t.Fatalf("expected scoreboard payload, got %+v", resp)
	}
	if len(resp.Scoreboard.List) != 1 {
		t.Fatalf("expected missing team to be skipped, got %+v", resp.Scoreboard.List)
	}
	if resp.Scoreboard.List[0].TeamID != 11 || resp.Scoreboard.List[0].TeamName != "Alpha" {
		t.Fatalf("unexpected scoreboard item: %+v", resp.Scoreboard.List[0])
	}
}

func TestScoreboardServiceGetScoreboardPaginatesAfterFilteringInvisibleTeams(t *testing.T) {
	t.Parallel()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mini.Close()

	redisClient := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	defer func() {
		_ = redisClient.Close()
	}()

	now := time.Now()
	contestID := int64(79)
	key := rediskeys.RankContestTeamKey(contestID)
	if err := redisClient.ZAdd(context.Background(), key,
		redislib.Z{Score: 600, Member: contestdomain.TeamIDToMember(11)},
		redislib.Z{Score: 500, Member: contestdomain.TeamIDToMember(12)},
		redislib.Z{Score: 400, Member: contestdomain.TeamIDToMember(13)},
	).Err(); err != nil {
		t.Fatalf("seed redis scoreboard: %v", err)
	}

	service := NewScoreboardService(&scoreboardQueryRepoStub{
		contest: &model.Contest{
			ID:        contestID,
			Title:     "scoreboard",
			Mode:      model.ContestModeJeopardy,
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Status:    model.ContestStatusRunning,
		},
		teams: map[int64]*model.Team{
			11: {ID: 11, Name: "Alpha"},
			13: {ID: 13, Name: "Gamma"},
		},
		statsMap: map[int64]contestports.ScoreboardTeamStats{
			11: {SolvedCount: 6},
			13: {SolvedCount: 4},
		},
	}, redisClient, &config.ContestConfig{}, zap.NewNop())

	resp, err := service.GetScoreboard(context.Background(), contestID, 2, 1)
	if err != nil {
		t.Fatalf("GetScoreboard() error = %v", err)
	}
	if resp.Scoreboard == nil {
		t.Fatalf("expected scoreboard payload, got %+v", resp)
	}
	if resp.Scoreboard.Total != 2 {
		t.Fatalf("expected filtered total 2, got %d", resp.Scoreboard.Total)
	}
	if len(resp.Scoreboard.List) != 1 {
		t.Fatalf("expected second visible team on page 2, got %+v", resp.Scoreboard.List)
	}
	if resp.Scoreboard.List[0].Rank != 2 || resp.Scoreboard.List[0].TeamID != 13 {
		t.Fatalf("unexpected page-2 scoreboard item: %+v", resp.Scoreboard.List[0])
	}
}
