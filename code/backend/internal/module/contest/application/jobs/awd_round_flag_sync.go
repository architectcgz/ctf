package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) syncRoundFlags(ctx context.Context, contest *model.Contest, activeRound int, now time.Time) error {
	if contest == nil || u.redis == nil {
		return nil
	}
	if activeRound <= 0 {
		return u.redis.Del(ctx, rediskeys.AWDCurrentRoundKey(contest.ID)).Err()
	}
	if u.flagSecret == "" {
		u.log.Warn("skip_awd_flag_rotation_due_to_empty_secret", zap.Int64("contest_id", contest.ID))
		return nil
	}

	round, err := u.findRoundByNumber(ctx, contest.ID, activeRound)
	if err != nil {
		return err
	}
	assignments, err := u.buildRoundFlagAssignments(ctx, contest.ID, round)
	if err != nil {
		return err
	}
	if len(assignments) == 0 {
		return u.redis.Set(ctx, rediskeys.AWDCurrentRoundKey(contest.ID), round.RoundNumber, 0).Err()
	}

	fields := make(map[string]any, len(assignments))
	for _, item := range assignments {
		fields[rediskeys.AWDRoundFlagField(item.TeamID, item.ChallengeID)] = item.Flag
	}

	pipe := u.redis.TxPipeline()
	pipe.Set(ctx, rediskeys.AWDCurrentRoundKey(contest.ID), round.RoundNumber, 0)
	roundKey := rediskeys.AWDRoundFlagsKey(contest.ID, round.ID)
	pipe.Del(ctx, roundKey)
	pipe.HSet(ctx, roundKey, fields)
	if ttl := u.currentRoundTTL(contest, round, now); ttl > 0 {
		pipe.Expire(ctx, roundKey, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return u.injector.InjectRoundFlags(ctx, contest, round, assignments)
}

func (u *AWDRoundUpdater) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	return u.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}

func (u *AWDRoundUpdater) buildRoundFlagAssignments(ctx context.Context, contestID int64, round *model.AWDRound) ([]contestports.AWDFlagAssignment, error) {
	teams, err := u.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	challenges, err := u.loadContestChallenges(ctx, contestID)
	if err != nil {
		return nil, err
	}

	assignments := make([]contestports.AWDFlagAssignment, 0, len(teams)*len(challenges))
	for _, team := range teams {
		for _, challenge := range challenges {
			assignments = append(assignments, contestports.AWDFlagAssignment{
				TeamID:      team.ID,
				ChallengeID: challenge.ID,
				Flag:        contestdomain.BuildAWDRoundFlag(contestID, round.RoundNumber, team.ID, challenge.ID, u.flagSecret, challenge.FlagPrefix),
			})
		}
	}
	return assignments, nil
}

func (u *AWDRoundUpdater) loadContestTeams(ctx context.Context, contestID int64) ([]model.Team, error) {
	teamPtrs, err := u.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	teams := make([]model.Team, 0, len(teamPtrs))
	for _, team := range teamPtrs {
		if team != nil {
			teams = append(teams, *team)
		}
	}
	return teams, nil
}

func (u *AWDRoundUpdater) loadContestChallenges(ctx context.Context, contestID int64) ([]model.Challenge, error) {
	return u.repo.ListChallengesByContest(ctx, contestID)
}

func (u *AWDRoundUpdater) currentRoundTTL(contest *model.Contest, round *model.AWDRound, now time.Time) time.Duration {
	if contest == nil || round == nil {
		return 0
	}
	roundEnd := contest.EndTime
	if round.StartedAt != nil {
		candidate := round.StartedAt.Add(u.cfg.RoundInterval)
		if candidate.Before(roundEnd) {
			roundEnd = candidate
		}
	}
	ttl := roundEnd.Sub(now)
	if ttl <= 0 {
		return time.Second
	}
	return ttl
}
