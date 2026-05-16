package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
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
		key := awdServiceTargetKey{teamID: instance.TeamID, serviceID: instance.ServiceID}
		grouped[key] = append(grouped[key], instance)
	}

	now := time.Now().UTC()
	records := make([]model.AWDTeamService, 0, len(teams)*len(definitions))
	statusEntries := make([]contestports.AWDServiceStatusEntry, 0, len(teams)*len(definitions))
	for _, team := range teams {
		for _, definition := range definitions {
			key := awdServiceTargetKey{teamID: team.ID, serviceID: definition.ServiceID}
			outcome, checkErr := u.checkTeamChallengeServices(ctx, contest, contest.ID, team.ID, definition, grouped[key], round, source)
			if checkErr != nil {
				return checkErr
			}

			records = append(records, model.AWDTeamService{
				RoundID:        round.ID,
				TeamID:         team.ID,
				ServiceID:      definition.ServiceID,
				AWDChallengeID: definition.AWDChallengeID,
				ServiceStatus:  outcome.serviceStatus,
				CheckResult:    outcome.checkResult,
				CheckerType:    outcome.checkerType,
				SLAScore:       outcome.slaScore,
				DefenseScore:   outcome.defenseScore,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
			statusEntries = append(statusEntries, contestports.AWDServiceStatusEntry{
				TeamID:    team.ID,
				ServiceID: definition.ServiceID,
				Status:    outcome.serviceStatus,
			})
		}
	}

	return u.persistRoundServiceChecks(ctx, contest, round, records, statusEntries)
}
