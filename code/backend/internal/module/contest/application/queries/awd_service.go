package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type AWDService struct {
	repo        contestports.AWDRepository
	contestRepo contestports.ContestLookupRepository
}

func NewAWDService(repo contestports.AWDRepository, contestRepo contestports.ContestLookupRepository) *AWDService {
	return &AWDService{
		repo:        repo,
		contestRepo: contestRepo,
	}
}

func (s *AWDService) ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error) {
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
		resp = append(resp, contestdomain.AWDRoundRespFromModel(&roundCopy))
	}
	return resp, nil
}

func (s *AWDService) ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
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
		resp = append(resp, contestdomain.AWDTeamServiceRespFromModel(&recordCopy, teamName))
	}
	return resp, nil
}

func (s *AWDService) ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error) {
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
		resp = append(resp, contestdomain.AWDAttackLogRespFromModel(&logCopy, attackerName, victimName))
	}
	return resp, nil
}

func (s *AWDService) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error) {
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
		switch contestdomain.NormalizeAWDCheckSource(contestdomain.ParseAWDCheckResult(service.CheckResult)["check_source"]) {
		case contestdomain.AWDCheckSourceScheduler:
			metrics.SchedulerCheckCount++
		case contestdomain.AWDCheckSourceManualCurrent:
			metrics.ManualCurrentRoundChecks++
		case contestdomain.AWDCheckSourceManualSelected:
			metrics.ManualSelectedRoundChecks++
		case contestdomain.AWDCheckSourceManualService:
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
		switch contestdomain.NormalizeAWDAttackSource(logEntry.Source) {
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
	contestdomain.SortAWDSummaryItems(respItems)

	return &dto.AWDRoundSummaryResp{
		Round:   contestdomain.AWDRoundRespFromModel(round),
		Metrics: metrics,
		Items:   respItems,
	}, nil
}

func (s *AWDService) ensureAWDContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrForbidden
	}
	return contest, nil
}

func (s *AWDService) ensureAWDRound(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
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

func (s *AWDService) loadContestTeams(ctx context.Context, contestID int64) (map[int64]*model.Team, error) {
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
