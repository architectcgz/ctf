package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

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
