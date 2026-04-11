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
	definitions, err := u.loadContestServiceDefinitions(ctx, contest.ID)
	if err != nil {
		return err
	}
	instances, err := u.loadContestServiceInstances(ctx, contest.ID, definitions)
	if err != nil {
		return err
	}

	grouped := make(map[awdServiceTargetKey][]contestports.AWDServiceInstance, len(instances))
	for _, instance := range instances {
		key := awdServiceTargetKey{teamID: instance.TeamID, challengeID: instance.ChallengeID}
		grouped[key] = append(grouped[key], instance)
	}

	now := time.Now()
	records := make([]model.AWDTeamService, 0, len(teams)*len(definitions))
	statusFields := make(map[string]any, len(teams)*len(definitions))
	for _, team := range teams {
		for _, definition := range definitions {
			key := awdServiceTargetKey{teamID: team.ID, challengeID: definition.ChallengeID}
			outcome, checkErr := u.checkTeamChallengeServices(ctx, definition, grouped[key], round, source)
			if checkErr != nil {
				return checkErr
			}

			records = append(records, model.AWDTeamService{
				RoundID:       round.ID,
				TeamID:        team.ID,
				ChallengeID:   definition.ChallengeID,
				ServiceStatus: outcome.serviceStatus,
				CheckResult:   outcome.checkResult,
				CheckerType:   outcome.checkerType,
				SLAScore:      outcome.slaScore,
				DefenseScore:  outcome.defenseScore,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			statusFields[rediskeys.AWDRoundFlagField(team.ID, definition.ChallengeID)] = outcome.serviceStatus
		}
	}

	return u.persistRoundServiceChecks(ctx, contest, round, records, statusFields)
}
