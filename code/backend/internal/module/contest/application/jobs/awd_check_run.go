package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) runRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error {
	if contest == nil || round == nil {
		return nil
	}

	teams, err := u.loadContestTeams(ctx, contest.ID)
	if err != nil {
		return err
	}
	challenges, err := u.loadContestChallenges(ctx, contest.ID)
	if err != nil {
		return err
	}
	instances, err := u.loadContestServiceInstances(ctx, contest.ID, challenges)
	if err != nil {
		return err
	}

	grouped := make(map[awdServiceTargetKey][]contestports.AWDServiceInstance, len(instances))
	for _, instance := range instances {
		key := awdServiceTargetKey{teamID: instance.TeamID, challengeID: instance.ChallengeID}
		grouped[key] = append(grouped[key], instance)
	}

	now := time.Now()
	records := make([]model.AWDTeamService, 0, len(teams)*len(challenges))
	statusFields := make(map[string]any, len(teams)*len(challenges))
	for _, team := range teams {
		for _, challenge := range challenges {
			key := awdServiceTargetKey{teamID: team.ID, challengeID: challenge.ID}
			outcome, checkErr := u.checkTeamChallengeServices(ctx, grouped[key], source)
			if checkErr != nil {
				return checkErr
			}

			defenseScore := 0
			if outcome.serviceStatus == model.AWDServiceStatusUp {
				defenseScore = round.DefenseScore
			}
			records = append(records, model.AWDTeamService{
				RoundID:       round.ID,
				TeamID:        team.ID,
				ChallengeID:   challenge.ID,
				ServiceStatus: outcome.serviceStatus,
				CheckResult:   outcome.checkResult,
				DefenseScore:  defenseScore,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			statusFields[rediskeys.AWDRoundFlagField(team.ID, challenge.ID)] = outcome.serviceStatus
		}
	}

	return u.persistRoundServiceChecks(ctx, contest, round, records, statusFields)
}
