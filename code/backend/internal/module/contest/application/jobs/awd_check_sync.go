package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) syncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	if contest == nil {
		return nil
	}
	if activeRound <= 0 {
		if u.redis == nil {
			return nil
		}
		return u.redis.Del(ctx, rediskeys.AWDServiceStatusKey(contest.ID)).Err()
	}

	round, err := u.findRoundByNumber(ctx, contest.ID, activeRound)
	if err != nil {
		return err
	}
	return u.runRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceScheduler)
}

// RunRoundServiceChecks 允许后台运维链路手动触发轮次服务检查，并记录巡检来源。
func (u *AWDRoundUpdater) RunRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error {
	if contest == nil || round == nil {
		return nil
	}
	return u.runRoundServiceChecks(ctx, contest, round, source)
}

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

	if len(records) > 0 {
		if err := u.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
			if err := txRepo.UpsertTeamServices(ctx, records); err != nil {
				return err
			}
			return txRepo.RecalculateContestTeamScores(ctx, contest.ID)
		}); err != nil {
			return err
		}
	}

	shouldSyncLiveStatusCache, err := u.shouldSyncLiveServiceStatusCache(ctx, contest.ID, round)
	if err != nil {
		return err
	}

	if u.redis != nil && shouldSyncLiveStatusCache {
		pipe := u.redis.TxPipeline()
		statusKey := rediskeys.AWDServiceStatusKey(contest.ID)
		pipe.Del(ctx, statusKey)
		if len(statusFields) > 0 {
			pipe.HSet(ctx, statusKey, statusFields)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	if u.redis != nil {
		if err := u.repo.RebuildContestScoreboardCache(ctx, u.redis, contest.ID); err != nil {
			return err
		}
	}

	return nil
}
