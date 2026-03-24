package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type awdServiceCheckOutcome struct {
	serviceStatus string
	checkResult   string
}

type awdServiceCheckResult struct {
	CheckedAt            string                 `json:"checked_at"`
	CheckSource          string                 `json:"check_source,omitempty"`
	HealthPath           string                 `json:"health_path"`
	InstanceCount        int                    `json:"instance_count"`
	HealthyInstanceCount int                    `json:"healthy_instance_count"`
	FailedInstanceCount  int                    `json:"failed_instance_count"`
	StatusReason         string                 `json:"status_reason,omitempty"`
	Probe                string                 `json:"probe,omitempty"`
	LatencyMS            int64                  `json:"latency_ms,omitempty"`
	ErrorCode            string                 `json:"error_code,omitempty"`
	Error                string                 `json:"error,omitempty"`
	Targets              []awdCheckTargetResult `json:"targets,omitempty"`
}

type awdCheckTargetResult struct {
	AccessURL string                  `json:"access_url,omitempty"`
	Healthy   bool                    `json:"healthy"`
	Probe     string                  `json:"probe,omitempty"`
	LatencyMS int64                   `json:"latency_ms,omitempty"`
	ErrorCode string                  `json:"error_code,omitempty"`
	Error     string                  `json:"error,omitempty"`
	Attempts  []awdProbeAttemptResult `json:"attempts,omitempty"`
}

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

func (u *AWDRoundUpdater) shouldSyncLiveServiceStatusCache(ctx context.Context, contestID int64, round *model.AWDRound) (bool, error) {
	if u.redis == nil || u.repo == nil || contestID <= 0 || round == nil {
		return false, nil
	}

	currentRound, err := u.repo.FindRunningRound(ctx, contestID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
		currentRoundNumber, redisErr := u.redis.Get(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Result()
		if redisErr == nil {
			return strings.TrimSpace(currentRoundNumber) == fmt.Sprintf("%d", round.RoundNumber), nil
		}
		if !errors.Is(redisErr, redislib.Nil) {
			return false, redisErr
		}
		return false, nil
	}
	return currentRound.ID == round.ID, nil
}

func (u *AWDRoundUpdater) loadContestServiceInstances(ctx context.Context, contestID int64, challenges []model.Challenge) ([]contestports.AWDServiceInstance, error) {
	if len(challenges) == 0 {
		return nil, nil
	}
	challengeIDs := make([]int64, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}

	return u.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
}

func (u *AWDRoundUpdater) checkTeamChallengeServices(ctx context.Context, instances []contestports.AWDServiceInstance, source string) (*awdServiceCheckOutcome, error) {
	healthPath := normalizedAWDCheckerHealthPath(u.cfg.CheckerHealthPath)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
		HealthPath:           healthPath,
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if len(instances) == 0 {
		result.StatusReason = "no_running_instances"
		result.ErrorCode = "no_running_instances"
		result.Error = "no_running_instances"
		raw, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		return &awdServiceCheckOutcome{
			serviceStatus: model.AWDServiceStatusDown,
			checkResult:   string(raw),
		}, nil
	}

	healthyCount := 0
	bestLatency := int64(0)
	bestProbe := ""
	targets := make([]awdCheckTargetResult, 0, len(instances))
	lastErrCode := ""
	firstErr := ""
	firstErrCode := ""
	for _, instance := range instances {
		probe := u.probeServiceInstance(ctx, instance.AccessURL, healthPath)
		target := awdCheckTargetResult{
			AccessURL: instance.AccessURL,
			Healthy:   probe.healthy,
			Probe:     probe.probe,
			LatencyMS: probe.latencyMS,
			ErrorCode: probe.errorCode,
			Error:     probe.err,
			Attempts:  probe.attempts,
		}
		targets = append(targets, target)
		if probe.healthy {
			healthyCount++
			if bestLatency == 0 || probe.latencyMS < bestLatency {
				bestLatency = probe.latencyMS
				bestProbe = probe.probe
			}
			continue
		}
		lastErrCode = probe.errorCode
		if firstErr == "" {
			firstErr = probe.err
			firstErrCode = probe.errorCode
		}
	}

	result.Targets = targets
	result.HealthyInstanceCount = healthyCount
	result.FailedInstanceCount = len(instances) - healthyCount
	if bestLatency > 0 {
		result.LatencyMS = bestLatency
	}
	if bestProbe != "" {
		result.Probe = bestProbe
	}

	status := model.AWDServiceStatusDown
	if healthyCount > 0 {
		status = model.AWDServiceStatusUp
		if healthyCount == len(instances) {
			result.StatusReason = "healthy"
		} else {
			result.StatusReason = "partial_available"
		}
	} else {
		if firstErrCode != "" {
			result.ErrorCode = firstErrCode
		} else if lastErrCode != "" {
			result.ErrorCode = lastErrCode
		}
		if firstErr != "" {
			result.Error = firstErr
		}
		if result.ErrorCode != "" {
			result.StatusReason = result.ErrorCode
		} else {
			result.StatusReason = "all_probes_failed"
		}
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &awdServiceCheckOutcome{
		serviceStatus: status,
		checkResult:   string(raw),
	}, nil
}
