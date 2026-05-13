package ports_test

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyContestScoreboardStateStore struct{}

func (ctxOnlyContestScoreboardStateStore) HasFrozenScoreboardSnapshot(context.Context, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyContestScoreboardStateStore) CreateFrozenScoreboardSnapshot(context.Context, int64) error {
	return nil
}

func (ctxOnlyContestScoreboardStateStore) ClearFrozenScoreboardSnapshot(context.Context, int64) error {
	return nil
}

func (ctxOnlyContestScoreboardStateStore) ListLiveScoreboard(context.Context, int64) ([]contestports.ScoreboardMemberScore, error) {
	return nil, nil
}

func (ctxOnlyContestScoreboardStateStore) ListFrozenScoreboard(context.Context, int64) ([]contestports.ScoreboardMemberScore, error) {
	return nil, nil
}

func (ctxOnlyContestScoreboardStateStore) GetLiveTeamRank(context.Context, int64, int64) (contestports.ScoreboardTeamRank, bool, error) {
	return contestports.ScoreboardTeamRank{}, false, nil
}

func (ctxOnlyContestScoreboardStateStore) IncrementLiveTeamScore(context.Context, int64, int64, float64) error {
	return nil
}

func (ctxOnlyContestScoreboardStateStore) ReplaceLiveScoreboard(context.Context, int64, []contestports.ScoreboardTeamScoreEntry) error {
	return nil
}

var _ contestports.ContestScoreboardStateStore = (*ctxOnlyContestScoreboardStateStore)(nil)
