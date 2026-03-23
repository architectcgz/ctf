package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

const (
	defaultAWDRoundAttackScore  = 50
	defaultAWDRoundDefenseScore = 50
)

type AWDRoundUpdater struct {
	repo       AWDRepository
	redis      *redislib.Client
	cfg        config.ContestAWDConfig
	flagSecret string
	injector   AWDFlagInjector
	httpClient *http.Client
	log        *zap.Logger
}

type awdServiceTargetKey struct {
	teamID      int64
	challengeID int64
}

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

type awdProbeAttemptResult struct {
	Probe     string `json:"probe"`
	Healthy   bool   `json:"healthy"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type awdInstanceProbeResult struct {
	healthy   bool
	latencyMS int64
	probe     string
	errorCode string
	err       string
	attempts  []awdProbeAttemptResult
}

type awdCheckError struct {
	code    string
	message string
}

type noopAWDFlagInjector struct {
	log *zap.Logger
}

func (e awdCheckError) Error() string {
	return e.message
}

func (i *noopAWDFlagInjector) InjectRoundFlags(_ context.Context, contest *model.Contest, round *model.AWDRound, assignments []AWDFlagAssignment) error {
	if i == nil || i.log == nil || contest == nil || round == nil {
		return nil
	}
	i.log.Debug("skip_awd_flag_injection",
		zap.Int64("contest_id", contest.ID),
		zap.Int64("round_id", round.ID),
		zap.Int("assignment_count", len(assignments)),
	)
	return nil
}

func NewAWDRoundUpdater(
	repo AWDRepository,
	redis *redislib.Client,
	cfg config.ContestAWDConfig,
	flagSecret string,
	injector AWDFlagInjector,
	log *zap.Logger,
) *AWDRoundUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if injector == nil {
		injector = &noopAWDFlagInjector{log: log.Named("awd_flag_injector")}
	}
	return &AWDRoundUpdater{
		repo:       repo,
		redis:      redis,
		cfg:        cfg,
		flagSecret: flagSecret,
		injector:   injector,
		httpClient: &http.Client{Timeout: normalizedAWDCheckerTimeout(cfg.CheckerTimeout)},
		log:        log,
	}
}

func (u *AWDRoundUpdater) Start(ctx context.Context) {
	u.UpdateRoundsAt(ctx, time.Now())

	ticker := time.NewTicker(u.cfg.SchedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.UpdateRoundsAt(ctx, time.Now())
		}
	}
}

func (u *AWDRoundUpdater) UpdateRoundsAt(ctx context.Context, now time.Time) {
	if u.repo == nil {
		return
	}
	lock, acquired, err := redislock.Acquire(ctx, u.redis, rediskeys.AWDSchedulerLockKey(), u.cfg.SchedulerLockTTL)
	if err != nil {
		u.log.Error("acquire_awd_scheduler_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("awd_scheduler_lock_held_elsewhere")
		return
	}
	if lock != nil {
		defer func() {
			released, releaseErr := lock.Release(ctx)
			if releaseErr != nil {
				u.log.Error("release_awd_scheduler_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("awd_scheduler_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}

	recentCutoff := now.Add(-u.cfg.RoundInterval)
	contests, err := u.repo.ListSchedulableAWDContests(ctx, now, recentCutoff, u.cfg.SchedulerBatchSize)
	if err != nil {
		u.log.Error("list_awd_contests_failed", zap.Error(err))
		return
	}

	for i := range contests {
		contestCopy := contests[i]
		u.syncContestRounds(ctx, &contestCopy, now)
	}
}

func (u *AWDRoundUpdater) syncContestRounds(ctx context.Context, contest *model.Contest, now time.Time) {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok {
		return
	}

	lockRound := activeRound
	if lockRound == 0 {
		lockRound = totalRounds
	}
	if lockRound <= 0 {
		return
	}

	acquired, err := u.acquireRoundLock(ctx, contest.ID, lockRound)
	if err != nil {
		u.log.Error("acquire_awd_round_lock_failed", zap.Int64("contest_id", contest.ID), zap.Int("round_number", lockRound), zap.Error(err))
		return
	}
	if !acquired {
		return
	}

	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		u.log.Error("sync_awd_rounds_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Int("total_rounds", totalRounds), zap.Error(err))
		return
	}

	if err := u.syncRoundFlags(ctx, contest, activeRound, now); err != nil {
		u.log.Error("sync_awd_round_flags_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
	if err := u.syncRoundServiceChecks(ctx, contest, activeRound); err != nil {
		u.log.Error("sync_awd_service_checks_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
}

func (u *AWDRoundUpdater) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok || activeRound <= 0 {
		return gorm.ErrRecordNotFound
	}
	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		return err
	}
	return u.syncRoundFlags(ctx, contest, activeRound, now)
}

func (u *AWDRoundUpdater) SetHTTPClient(client *http.Client) {
	if u == nil || client == nil {
		return
	}
	u.httpClient = client
}

func (u *AWDRoundUpdater) SyncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	return u.syncRoundServiceChecks(ctx, contest, activeRound)
}

func (u *AWDRoundUpdater) calculateRoundPlan(contest *model.Contest, now time.Time) (int, int, bool) {
	if contest == nil {
		return 0, 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, 0, false
	}
	if now.Before(contest.StartTime) {
		return 0, 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + u.cfg.RoundInterval - 1) / u.cfg.RoundInterval)
	if totalRounds <= 0 {
		return 0, 0, false
	}
	if !now.Before(contest.EndTime) {
		return 0, totalRounds, true
	}

	activeRound := int(now.Sub(contest.StartTime)/u.cfg.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, totalRounds, true
}

func (u *AWDRoundUpdater) reconcileRounds(ctx context.Context, contest *model.Contest, activeRound, totalRounds int) error {
	scheduledRounds := totalRounds
	if activeRound > 0 {
		scheduledRounds = activeRound
	}

	return u.repo.WithinTransaction(ctx, func(txRepo AWDRepository) error {
		existingRounds, err := txRepo.ListRoundsByContest(ctx, contest.ID)
		if err != nil {
			return err
		}

		scoreByRound := make(map[int][2]int, len(existingRounds))
		lastAttackScore := defaultAWDRoundAttackScore
		lastDefenseScore := defaultAWDRoundDefenseScore
		for _, item := range existingRounds {
			scoreByRound[item.RoundNumber] = [2]int{item.AttackScore, item.DefenseScore}
			lastAttackScore = item.AttackScore
			lastDefenseScore = item.DefenseScore
		}

		for roundNumber := 1; roundNumber <= scheduledRounds; roundNumber++ {
			roundStart := contest.StartTime.Add(time.Duration(roundNumber-1) * u.cfg.RoundInterval)
			roundEnd := roundStart.Add(u.cfg.RoundInterval)
			if roundEnd.After(contest.EndTime) {
				roundEnd = contest.EndTime
			}

			status := model.AWDRoundStatusFinished
			var endedAt *time.Time
			if roundNumber == activeRound {
				status = model.AWDRoundStatusRunning
			} else {
				endedAt = &roundEnd
			}

			attackScore := lastAttackScore
			defenseScore := lastDefenseScore
			if scores, ok := scoreByRound[roundNumber]; ok {
				attackScore = scores[0]
				defenseScore = scores[1]
			} else {
				scoreByRound[roundNumber] = [2]int{attackScore, defenseScore}
			}
			lastAttackScore = attackScore
			lastDefenseScore = defenseScore

			record := &model.AWDRound{
				ContestID:    contest.ID,
				RoundNumber:  roundNumber,
				Status:       status,
				StartedAt:    &roundStart,
				EndedAt:      endedAt,
				AttackScore:  attackScore,
				DefenseScore: defenseScore,
			}
			if err := txRepo.UpsertRound(ctx, record); err != nil {
				return err
			}
		}
		return nil
	})
}

func (u *AWDRoundUpdater) acquireRoundLock(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if u.redis == nil {
		return true, nil
	}
	return u.redis.SetNX(ctx, rediskeys.AWDRoundLockKey(contestID, roundNumber), "1", u.cfg.RoundLockTTL).Result()
}

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

func (u *AWDRoundUpdater) buildRoundFlagAssignments(ctx context.Context, contestID int64, round *model.AWDRound) ([]AWDFlagAssignment, error) {
	teams, err := u.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	challenges, err := u.loadContestChallenges(ctx, contestID)
	if err != nil {
		return nil, err
	}

	assignments := make([]AWDFlagAssignment, 0, len(teams)*len(challenges))
	for _, team := range teams {
		for _, challenge := range challenges {
			assignments = append(assignments, AWDFlagAssignment{
				TeamID:      team.ID,
				ChallengeID: challenge.ID,
				Flag:        buildAWDRoundFlag(contestID, round.RoundNumber, team.ID, challenge.ID, u.flagSecret, challenge.FlagPrefix),
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
	return u.runRoundServiceChecks(ctx, contest, round, awdCheckSourceScheduler)
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

	grouped := make(map[awdServiceTargetKey][]AWDServiceInstance, len(instances))
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
		if err := u.repo.WithinTransaction(ctx, func(txRepo AWDRepository) error {
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

func (u *AWDRoundUpdater) loadContestServiceInstances(ctx context.Context, contestID int64, challenges []model.Challenge) ([]AWDServiceInstance, error) {
	if len(challenges) == 0 {
		return nil, nil
	}
	challengeIDs := make([]int64, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}

	return u.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
}

func (u *AWDRoundUpdater) checkTeamChallengeServices(ctx context.Context, instances []AWDServiceInstance, source string) (*awdServiceCheckOutcome, error) {
	healthPath := normalizedAWDCheckerHealthPath(u.cfg.CheckerHealthPath)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          normalizedAWDCheckSource(source),
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

func (u *AWDRoundUpdater) probeServiceInstance(ctx context.Context, accessURL, healthPath string) awdInstanceProbeResult {
	startedAt := time.Now()
	attempts := make([]awdProbeAttemptResult, 0, 1)
	targetURL, err := buildAWDHealthCheckURL(accessURL, healthPath)
	if err == nil {
		client := u.httpClient
		if client == nil {
			client = &http.Client{Timeout: normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout)}
		}
		reqCtx, cancel := context.WithTimeout(ctx, normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout))
		defer cancel()

		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
		if reqErr == nil {
			resp, doErr := client.Do(req)
			if doErr == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
					attempts = append(attempts, awdProbeAttemptResult{
						Probe:     "http",
						Healthy:   true,
						LatencyMS: time.Since(startedAt).Milliseconds(),
					})
					return awdInstanceProbeResult{
						healthy:   true,
						latencyMS: time.Since(startedAt).Milliseconds(),
						probe:     "http",
						attempts:  attempts,
					}
				}
				err = newAWDCheckError("unexpected_http_status", fmt.Sprintf("unexpected_http_status:%d", resp.StatusCode))
			} else {
				err = newAWDCheckError("http_request_failed", sanitizeAWDCheckError(doErr))
			}
		} else {
			err = newAWDCheckError("http_request_failed", sanitizeAWDCheckError(reqErr))
		}
		errorCode, errorMessage := normalizeAWDCheckError(err, "http_request_failed")
		attempts = append(attempts, awdProbeAttemptResult{
			Probe:     "http",
			Healthy:   false,
			ErrorCode: errorCode,
			Error:     errorMessage,
		})
	} else {
		errorCode, errorMessage := normalizeAWDCheckError(err, "invalid_access_url")
		attempts = append(attempts, awdProbeAttemptResult{
			Probe:     "http",
			Healthy:   false,
			ErrorCode: errorCode,
			Error:     errorMessage,
		})
	}

	errorCode, errorMessage := normalizeAWDCheckError(err, "unknown_checker_error")
	if errorCode == "" && len(attempts) > 0 {
		lastAttempt := attempts[len(attempts)-1]
		errorCode = lastAttempt.ErrorCode
		errorMessage = lastAttempt.Error
	}
	return awdInstanceProbeResult{
		healthy:   false,
		probe:     "http",
		errorCode: errorCode,
		err:       errorMessage,
		attempts:  attempts,
	}
}

func buildAWDHealthCheckURL(accessURL, healthPath string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(accessURL))
	if err != nil {
		return "", newAWDCheckError("invalid_access_url", sanitizeAWDCheckError(err))
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", newAWDCheckError("invalid_access_url", "invalid_access_url")
	}
	parsed.Path = path.Join("/", strings.TrimSpace(parsed.Path), strings.TrimSpace(healthPath))
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func newAWDCheckError(code, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		message = code
	}
	return awdCheckError{code: code, message: message}
}

func normalizeAWDCheckError(err error, fallbackCode string) (string, string) {
	if err == nil {
		return "", ""
	}
	var typedErr awdCheckError
	if ok := errors.As(err, &typedErr); ok {
		return typedErr.code, sanitizeAWDCheckError(typedErr)
	}
	return fallbackCode, sanitizeAWDCheckError(err)
}

func sanitizeAWDCheckError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "unknown_checker_error"
	}
	return msg
}

func normalizedAWDCheckerTimeout(value time.Duration) time.Duration {
	if value <= 0 {
		return 3 * time.Second
	}
	return value
}

func normalizedAWDCheckerHealthPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "/health"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func normalizedAWDCheckSource(value string) string {
	switch strings.TrimSpace(value) {
	case awdCheckSourceManualCurrent:
		return awdCheckSourceManualCurrent
	case awdCheckSourceManualSelected:
		return awdCheckSourceManualSelected
	case awdCheckSourceManualService:
		return awdCheckSourceManualService
	default:
		return awdCheckSourceScheduler
	}
}
