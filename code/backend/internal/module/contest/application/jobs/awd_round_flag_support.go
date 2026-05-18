package jobs

import (
	"context"
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*contestentity.AWDRound, error) {
	return u.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}

func (u *AWDRoundUpdater) buildRoundFlagAssignments(ctx context.Context, contestID int64, round *contestentity.AWDRound) ([]contestports.AWDFlagAssignment, error) {
	teams, err := u.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	definitions, err := u.loadContestServiceDefinitions(ctx, contestID)
	if err != nil {
		return nil, err
	}

	assignments := make([]contestports.AWDFlagAssignment, 0, len(teams)*len(definitions))
	for _, team := range teams {
		for _, definition := range definitions {
			assignments = append(assignments, contestports.AWDFlagAssignment{
				ServiceID:      definition.ServiceID,
				TeamID:         team.ID,
				AWDChallengeID: definition.AWDChallengeID,
				Flag:           contestdomain.BuildAWDRoundFlag(contestID, round.RoundNumber, team.ID, definition.AWDChallengeID, u.flagSecret, definition.FlagPrefix),
			})
		}
	}
	return assignments, nil
}

func (u *AWDRoundUpdater) loadContestTeams(ctx context.Context, contestID int64) ([]contestentity.Team, error) {
	teamPtrs, err := u.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	teams := make([]contestentity.Team, 0, len(teamPtrs))
	for _, team := range teamPtrs {
		if team != nil {
			teams = append(teams, *team)
		}
	}
	return teams, nil
}

func (u *AWDRoundUpdater) currentRoundTTL(contest *contestentity.Contest, round *contestentity.AWDRound, now time.Time) time.Duration {
	if contest == nil || round == nil {
		return 0
	}
	roundEnd := contestdomain.ContestEffectiveEndTime(contest)
	if round.StartedAt != nil {
		candidate := round.StartedAt.Add(u.cfg.RoundInterval).Add(contestdomain.ContestPausedDuration(contest))
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
