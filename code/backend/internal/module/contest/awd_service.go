package contest

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestapp "ctf-platform/internal/module/contest/application"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type AWDService interface {
	CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error)
	ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error)
	ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error)
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error)
}

type awdService struct {
	repo        *AWDRepository
	redis       *redislib.Client
	contestRepo contestapp.Repository
	flagSecret  string
	awdConfig   config.ContestAWDConfig
	log         *zap.Logger
}

func NewAWDService(
	repo *AWDRepository,
	contestRepo contestapp.Repository,
	redis *redislib.Client,
	flagSecret string,
	awdConfig config.ContestAWDConfig,
	log *zap.Logger,
) AWDService {
	if log == nil {
		log = zap.NewNop()
	}
	return &awdService{repo: repo, redis: redis, contestRepo: contestRepo, flagSecret: flagSecret, awdConfig: awdConfig, log: log}
}

func (s *awdService) CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round := &model.AWDRound{
		ContestID:    contestID,
		RoundNumber:  req.RoundNumber,
		Status:       model.AWDRoundStatusPending,
		AttackScore:  50,
		DefenseScore: 50,
	}
	if req.Status != nil && *req.Status != "" {
		round.Status = *req.Status
	}
	if req.AttackScore != nil {
		round.AttackScore = *req.AttackScore
	}
	if req.DefenseScore != nil {
		round.DefenseScore = *req.DefenseScore
	}
	if err := s.repo.CreateRound(ctx, round); err != nil {
		if isUniqueConstraintError(err) {
			return nil, errcode.ErrConflict
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toAWDRoundResp(round), nil
}

func (s *awdService) ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	rounds, err := s.repo.ListRoundsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*dto.AWDRoundResp, 0, len(rounds))
	for _, round := range rounds {
		roundCopy := round
		resp = append(resp, toAWDRoundResp(&roundCopy))
	}
	return resp, nil
}

func (s *awdService) RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}
	round, err := s.resolveCurrentRoundForContest(ctx, contest)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RunRoundServiceChecks(ctx, s.redis, s.awdConfig, s.flagSecret, contest, round, awdCheckSourceManualCurrent, s.log.Named("awd_manual_checker")); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *awdService) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RunRoundServiceChecks(ctx, s.redis, s.awdConfig, s.flagSecret, contest, round, awdCheckSourceManualSelected, s.log.Named("awd_manual_checker")); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *awdService) buildCheckerRunResp(ctx context.Context, contestID int64, round *model.AWDRound) (*dto.AWDCheckerRunResp, error) {
	services, err := s.ListServices(ctx, contestID, round.ID)
	if err != nil {
		return nil, err
	}
	return &dto.AWDCheckerRunResp{
		Round:    toAWDRoundResp(round),
		Services: services,
	}, nil
}

func (s *awdService) UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	teamMap, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	team, ok := teamMap[req.TeamID]
	if !ok {
		return nil, errcode.ErrNotFound
	}
	if err := s.ensureContestChallenge(ctx, contestID, req.ChallengeID); err != nil {
		return nil, err
	}

	normalizedCheckResult := normalizeManualAWDCheckResult(req.CheckResult)
	checkResult, err := marshalAWDCheckResult(normalizedCheckResult)
	if err != nil {
		return nil, errcode.ErrInvalidParams
	}
	defenseScore := 0
	if req.ServiceStatus == model.AWDServiceStatusUp {
		defenseScore = round.DefenseScore
	}

	now := time.Now()
	var record *model.AWDTeamService
	if err := s.repo.WithinTransaction(ctx, func(txRepo *AWDRepository) error {
		var txErr error
		record, txErr = txRepo.UpsertServiceCheck(ctx, roundID, req.TeamID, req.ChallengeID, req.ServiceStatus, checkResult, defenseScore, now)
		if txErr != nil {
			return txErr
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	}); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.RebuildContestScoreboardCache(ctx, s.redis, contestID); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(ctx, s.redis, contestID, roundID, currentRoundID, req.TeamID, req.ChallengeID, req.ServiceStatus); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return toAWDTeamServiceResp(record, team.Name), nil
}

func (s *awdService) ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}

	records, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.AWDTeamServiceResp, 0, len(records))
	for _, record := range records {
		recordCopy := record
		teamName := ""
		if team := teams[record.TeamID]; team != nil {
			teamName = team.Name
		}
		resp = append(resp, toAWDTeamServiceResp(&recordCopy, teamName))
	}
	return resp, nil
}

func (s *awdService) CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error) {
	return s.createAttackLog(ctx, contestID, roundID, req, model.AWDAttackSourceManual)
}

func (s *awdService) createAttackLog(
	ctx context.Context,
	contestID, roundID int64,
	req *dto.CreateAWDAttackLogReq,
	source string,
) (*dto.AWDAttackLogResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	if req.AttackerTeamID == req.VictimTeamID {
		return nil, errcode.ErrInvalidParams
	}
	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if teams[req.AttackerTeamID] == nil || teams[req.VictimTeamID] == nil {
		return nil, errcode.ErrNotFound
	}
	if err := s.ensureContestChallenge(ctx, contestID, req.ChallengeID); err != nil {
		return nil, err
	}

	scoreGained := 0
	if req.IsSuccess {
		count, err := s.repo.CountSuccessfulAttacks(ctx, roundID, req.AttackerTeamID, req.VictimTeamID, req.ChallengeID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if count == 0 {
			scoreGained = round.AttackScore
		}
	}

	logRecord := &model.AWDAttackLog{
		RoundID:        roundID,
		AttackerTeamID: req.AttackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ChallengeID:    req.ChallengeID,
		AttackType:     req.AttackType,
		Source:         normalizeAWDAttackSource(source),
		SubmittedFlag:  req.SubmittedFlag,
		IsSuccess:      req.IsSuccess,
		ScoreGained:    scoreGained,
	}
	now := time.Now()
	if err := s.repo.WithinTransaction(ctx, func(txRepo *AWDRepository) error {
		if err := txRepo.CreateAttackLog(ctx, logRecord); err != nil {
			return err
		}
		if req.IsSuccess {
			if err := txRepo.ApplyAttackImpactToVictimService(ctx, round.ID, req.VictimTeamID, req.ChallengeID, scoreGained, now); err != nil {
				return err
			}
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	}); err != nil {
		return nil, err
	}
	if err := s.repo.RebuildContestScoreboardCache(ctx, s.redis, contestID); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(ctx, s.redis, contestID, roundID, currentRoundID, req.VictimTeamID, req.ChallengeID, model.AWDServiceStatusCompromised); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return toAWDAttackLogResp(logRecord, teams[req.AttackerTeamID].Name, teams[req.VictimTeamID].Name), nil
}

func (s *awdService) SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}

	attackerTeamID, err := s.resolveUserTeamID(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}
	round, err := s.resolveCurrentRoundForContest(ctx, contest)
	if err != nil {
		return nil, err
	}
	challengeItem, err := s.loadChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureContestChallenge(ctx, contestID, challengeID); err != nil {
		return nil, err
	}

	acceptedFlags, err := s.resolveAcceptedRoundFlags(ctx, contestID, round, req.VictimTeamID, challengeItem, now)
	if err != nil {
		return nil, err
	}
	isSuccess := false
	for _, candidate := range acceptedFlags {
		if crypto.ValidateFlag(req.Flag, candidate) {
			isSuccess = true
			break
		}
	}

	return s.createAttackLog(ctx, contestID, round.ID, &dto.CreateAWDAttackLogReq{
		AttackerTeamID: attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ChallengeID:    challengeID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      isSuccess,
	}, model.AWDAttackSourceSubmission)
}

func (s *awdService) ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}

	logs, err := s.repo.ListAttackLogsByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.AWDAttackLogResp, 0, len(logs))
	for _, item := range logs {
		logCopy := item
		attackerName := ""
		victimName := ""
		if team := teams[item.AttackerTeamID]; team != nil {
			attackerName = team.Name
		}
		if team := teams[item.VictimTeamID]; team != nil {
			victimName = team.Name
		}
		resp = append(resp, toAWDAttackLogResp(&logCopy, attackerName, victimName))
	}
	return resp, nil
}

func (s *awdService) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	services, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	attackLogs, err := s.repo.ListAttackLogsByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make(map[int64]*dto.AWDRoundSummaryItem, len(teams))
	metrics := &dto.AWDRoundMetrics{}
	for teamID, team := range teams {
		items[teamID] = &dto.AWDRoundSummaryItem{
			TeamID:   teamID,
			TeamName: team.Name,
		}
	}

	for _, service := range services {
		metrics.TotalServiceCount++
		if service.AttackReceived > 0 {
			metrics.AttackedServiceCount++
		}
		switch normalizeAWDCheckSource(parseAWDCheckResult(service.CheckResult)["check_source"]) {
		case awdCheckSourceScheduler:
			metrics.SchedulerCheckCount++
		case awdCheckSourceManualCurrent:
			metrics.ManualCurrentRoundChecks++
		case awdCheckSourceManualSelected:
			metrics.ManualSelectedRoundChecks++
		case awdCheckSourceManualService:
			metrics.ManualServiceCheckCount++
		}

		item := items[service.TeamID]
		if item == nil {
			continue
		}
		switch service.ServiceStatus {
		case model.AWDServiceStatusUp:
			metrics.ServiceUpCount++
			if service.AttackReceived > 0 {
				metrics.DefenseSuccessCount++
			}
			item.ServiceUpCount++
		case model.AWDServiceStatusDown:
			metrics.ServiceDownCount++
			item.ServiceDownCount++
		case model.AWDServiceStatusCompromised:
			metrics.ServiceCompromisedCount++
			item.ServiceCompromisedCount++
		}
		item.DefenseScore += service.DefenseScore
	}

	uniqueAttackersAgainst := make(map[int64]map[int64]struct{}, len(teams))
	for _, logEntry := range attackLogs {
		metrics.TotalAttackCount++
		if logEntry.IsSuccess {
			metrics.SuccessfulAttackCount++
		} else {
			metrics.FailedAttackCount++
		}
		switch normalizeAWDAttackSource(logEntry.Source) {
		case model.AWDAttackSourceSubmission:
			metrics.SubmissionAttackCount++
		case model.AWDAttackSourceManual:
			metrics.ManualAttackLogCount++
		default:
			metrics.LegacyAttackLogCount++
		}

		if item := items[logEntry.AttackerTeamID]; item != nil {
			if logEntry.IsSuccess {
				item.SuccessfulAttackCount++
			}
			item.AttackScore += logEntry.ScoreGained
		}
		if logEntry.IsSuccess {
			if item := items[logEntry.VictimTeamID]; item != nil {
				item.SuccessfulBreachCount++
			}
			if uniqueAttackersAgainst[logEntry.VictimTeamID] == nil {
				uniqueAttackersAgainst[logEntry.VictimTeamID] = make(map[int64]struct{})
			}
			uniqueAttackersAgainst[logEntry.VictimTeamID][logEntry.AttackerTeamID] = struct{}{}
		}
	}

	respItems := make([]*dto.AWDRoundSummaryItem, 0, len(items))
	for teamID, item := range items {
		item.UniqueAttackersAgainst = len(uniqueAttackersAgainst[teamID])
		item.TotalScore = item.AttackScore + item.DefenseScore
		respItems = append(respItems, item)
	}
	sortAWDSummaryItems(respItems)

	return &dto.AWDRoundSummaryResp{
		Round:   toAWDRoundResp(round),
		Metrics: metrics,
		Items:   respItems,
	}, nil
}

func (s *awdService) ensureAWDContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestapp.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrForbidden
	}
	return contest, nil
}

func (s *awdService) ensureAWDRound(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round, err := s.repo.FindRoundByContestAndID(ctx, contestID, roundID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return round, nil
}

func (s *awdService) loadContestTeams(ctx context.Context, contestID int64) (map[int64]*model.Team, error) {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	result := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		result[team.ID] = team
	}
	return result, nil
}

func (s *awdService) ensureContestChallenge(ctx context.Context, contestID, challengeID int64) error {
	ok, err := s.repo.ContestHasChallenge(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !ok {
		return errcode.ErrChallengeNotInContest
	}
	return nil
}

func (s *awdService) resolveUserTeamID(ctx context.Context, userID, contestID int64) (int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := registrationStatusError(registration.Status); err != nil {
			return 0, err
		}
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return 0, errcode.ErrAWDTeamRequired
		}
		return *registration.TeamID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrNotRegistered
		}
		return 0, errcode.ErrInternal.WithCause(err)
	}
	return team.ID, nil
}

func (s *awdService) resolveCurrentRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	return s.resolveCurrentRoundForContest(ctx, contest)
}

func (s *awdService) resolveCurrentRoundForContest(ctx context.Context, contest *model.Contest) (*model.AWDRound, error) {
	if contest == nil {
		return nil, errcode.ErrContestNotFound
	}
	now := time.Now()
	if activeRoundNumber, ok := s.calculateActiveRoundNumber(contest, now); ok {
		round, err := s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
		if err == nil {
			return round, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if err := s.ensureActiveRoundMaterialized(ctx, contest, now); err != nil {
			return nil, err
		}
		round, err = s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
		if err == nil {
			return round, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrAWDRoundNotActive
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	round, err := s.repo.FindRunningRound(ctx, contest.ID)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if s.redis != nil {
		roundNumberStr, err := s.redis.Get(ctx, rediskeys.AWDCurrentRoundKey(contest.ID)).Result()
		if err == nil {
			roundNumber, convErr := strconv.Atoi(strings.TrimSpace(roundNumberStr))
			if convErr == nil && roundNumber > 0 {
				round, findErr := s.repo.FindRoundByNumber(ctx, contest.ID, roundNumber)
				if findErr == nil {
					return round, nil
				}
			}
		} else if !errors.Is(err, redislib.Nil) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	return nil, errcode.ErrAWDRoundNotActive
}

func (s *awdService) ensureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	if contest == nil {
		return errcode.ErrContestNotFound
	}
	if err := s.repo.EnsureActiveRoundMaterialized(ctx, s.redis, s.awdConfig, s.flagSecret, contest, now, s.log.Named("awd_round_materializer")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrAWDRoundNotActive
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *awdService) calculateActiveRoundNumber(contest *model.Contest, now time.Time) (int, bool) {
	if contest == nil || s.awdConfig.RoundInterval <= 0 {
		return 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, false
	}
	if now.Before(contest.StartTime) || !now.Before(contest.EndTime) {
		return 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + s.awdConfig.RoundInterval - 1) / s.awdConfig.RoundInterval)
	if totalRounds <= 0 {
		return 0, false
	}

	activeRound := int(now.Sub(contest.StartTime)/s.awdConfig.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, true
}

func (s *awdService) resolveCurrentRoundID(ctx context.Context, contestID int64) (int64, error) {
	if !s.isLiveContestWindow(ctx, contestID) {
		return 0, nil
	}
	round, err := s.resolveCurrentRound(ctx, contestID)
	if err != nil {
		if err == errcode.ErrAWDRoundNotActive {
			return 0, nil
		}
		return 0, err
	}
	return round.ID, nil
}

func (s *awdService) isLiveContestWindow(ctx context.Context, contestID int64) bool {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil || contest == nil {
		return false
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return false
	}
	return contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusFrozen
}

func (s *awdService) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	return s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}

func (s *awdService) loadChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	challenge, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return challenge, nil
}

func (s *awdService) resolveAcceptedRoundFlags(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	victimTeamID int64,
	challenge *model.Challenge,
	now time.Time,
) ([]string, error) {
	currentFlag, err := s.resolveRoundFlag(ctx, contestID, round, victimTeamID, challenge)
	if err != nil {
		return nil, err
	}
	flags := []string{currentFlag}

	if !s.allowPreviousRoundFlag(round, now) {
		return flags, nil
	}

	previousRound, err := s.findRoundByNumber(ctx, contestID, round.RoundNumber-1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flags, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	previousFlag, err := s.resolveRoundFlag(ctx, contestID, previousRound, victimTeamID, challenge)
	if err != nil {
		if err == errcode.ErrAWDFlagUnavailable {
			return flags, nil
		}
		return nil, err
	}
	return append(flags, previousFlag), nil
}

func (s *awdService) allowPreviousRoundFlag(round *model.AWDRound, now time.Time) bool {
	if round == nil || round.RoundNumber <= 1 || s.awdConfig.PreviousRoundGrace <= 0 || round.StartedAt == nil {
		return false
	}
	return now.Before(round.StartedAt.Add(s.awdConfig.PreviousRoundGrace))
}

func (s *awdService) resolveRoundFlag(ctx context.Context, contestID int64, round *model.AWDRound, victimTeamID int64, challenge *model.Challenge) (string, error) {
	if round == nil || challenge == nil {
		return "", errcode.ErrAWDFlagUnavailable
	}
	if s.redis != nil {
		flag, err := s.redis.HGet(ctx, rediskeys.AWDRoundFlagsKey(contestID, round.ID), rediskeys.AWDRoundFlagField(victimTeamID, challenge.ID)).Result()
		if err == nil && strings.TrimSpace(flag) != "" {
			return flag, nil
		}
		if err != nil && !errors.Is(err, redislib.Nil) {
			return "", errcode.ErrInternal.WithCause(err)
		}
	}
	if strings.TrimSpace(s.flagSecret) == "" {
		return "", errcode.ErrAWDFlagUnavailable
	}
	return buildAWDRoundFlag(contestID, round.RoundNumber, victimTeamID, challenge.ID, s.flagSecret, challenge.FlagPrefix), nil
}

func toAWDRoundResp(round *model.AWDRound) *dto.AWDRoundResp {
	return &dto.AWDRoundResp{
		ID:           round.ID,
		ContestID:    round.ContestID,
		RoundNumber:  round.RoundNumber,
		Status:       round.Status,
		StartedAt:    round.StartedAt,
		EndedAt:      round.EndedAt,
		AttackScore:  round.AttackScore,
		DefenseScore: round.DefenseScore,
		CreatedAt:    round.CreatedAt,
		UpdatedAt:    round.UpdatedAt,
	}
}

func toAWDTeamServiceResp(record *model.AWDTeamService, teamName string) *dto.AWDTeamServiceResp {
	return &dto.AWDTeamServiceResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		TeamID:         record.TeamID,
		TeamName:       teamName,
		ChallengeID:    record.ChallengeID,
		ServiceStatus:  record.ServiceStatus,
		CheckResult:    parseAWDCheckResult(record.CheckResult),
		AttackReceived: record.AttackReceived,
		DefenseScore:   record.DefenseScore,
		AttackScore:    record.AttackScore,
		UpdatedAt:      record.UpdatedAt,
	}
}

func toAWDAttackLogResp(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
	return &dto.AWDAttackLogResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		AttackerTeamID: record.AttackerTeamID,
		AttackerTeam:   attackerTeam,
		VictimTeamID:   record.VictimTeamID,
		VictimTeam:     victimTeam,
		ChallengeID:    record.ChallengeID,
		AttackType:     record.AttackType,
		Source:         normalizeAWDAttackSource(record.Source),
		SubmittedFlag:  record.SubmittedFlag,
		IsSuccess:      record.IsSuccess,
		ScoreGained:    record.ScoreGained,
		CreatedAt:      record.CreatedAt,
	}
}

func normalizeAWDAttackSource(value string) string {
	switch strings.TrimSpace(value) {
	case model.AWDAttackSourceManual:
		return model.AWDAttackSourceManual
	case model.AWDAttackSourceSubmission:
		return model.AWDAttackSourceSubmission
	default:
		return model.AWDAttackSourceLegacy
	}
}

func normalizeAWDCheckSource(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	switch strings.TrimSpace(raw) {
	case awdCheckSourceScheduler:
		return awdCheckSourceScheduler
	case awdCheckSourceManualCurrent:
		return awdCheckSourceManualCurrent
	case awdCheckSourceManualSelected:
		return awdCheckSourceManualSelected
	case awdCheckSourceManualService:
		return awdCheckSourceManualService
	default:
		return ""
	}
}

func marshalAWDCheckResult(value map[string]any) (string, error) {
	if len(value) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func normalizeManualAWDCheckResult(value map[string]any) map[string]any {
	result := make(map[string]any, len(value)+2)
	for key, item := range value {
		result[key] = item
	}
	result["check_source"] = awdCheckSourceManualService
	if checkedAt, ok := result["checked_at"].(string); !ok || strings.TrimSpace(checkedAt) == "" {
		result["checked_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	return result
}

func parseAWDCheckResult(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}

func sortAWDSummaryItems(items []*dto.AWDRoundSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalScore != items[j].TotalScore {
			return items[i].TotalScore > items[j].TotalScore
		}
		return items[i].TeamID < items[j].TeamID
	})
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
