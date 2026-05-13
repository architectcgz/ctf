package queries

import (
	"context"
	"errors"
	"sort"
	"strings"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

const (
	awdWorkspaceEventDirectionOutgoing = "attack_out"
	awdWorkspaceEventDirectionIncoming = "attack_in"
)

func (s *AWDService) GetUserWorkspace(ctx context.Context, userID, contestID int64) (*AWDWorkspaceResult, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	resp := &AWDWorkspaceResult{
		ContestID:    contestID,
		Services:     []*AWDWorkspaceServiceResult{},
		Targets:      []*AWDWorkspaceTargetTeamResult{},
		RecentEvents: []*AWDWorkspaceRecentEventResult{},
	}

	currentRound, err := s.repo.FindRunningRound(ctx, contestID)
	if err != nil && !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if currentRound != nil {
		resp.CurrentRound = &AWDRoundResult{
			ID:           currentRound.ID,
			ContestID:    currentRound.ContestID,
			RoundNumber:  currentRound.RoundNumber,
			Status:       currentRound.Status,
			StartedAt:    currentRound.StartedAt,
			EndedAt:      currentRound.EndedAt,
			AttackScore:  currentRound.AttackScore,
			DefenseScore: currentRound.DefenseScore,
			CreatedAt:    currentRound.CreatedAt,
			UpdatedAt:    currentRound.UpdatedAt,
		}
	}

	myTeam, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil && !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if myTeam == nil {
		return resp, nil
	}
	resp.MyTeam = &AWDWorkspaceTeamResult{
		TeamID:   myTeam.ID,
		TeamName: myTeam.Name,
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	definitions, err := s.repo.ListServiceDefinitionsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	serviceMap := make(map[int64]*AWDWorkspaceServiceResult)
	serviceIDs := make([]int64, 0, len(definitions))
	for _, definition := range definitions {
		serviceIDs = append(serviceIDs, definition.ServiceID)
		item := ensureAWDWorkspaceService(serviceMap, definition.ServiceID, definition.AWDChallengeID)
		mergeAWDWorkspaceDefenseConnection(item, definition.DefenseWorkspace)
	}
	if err := s.populateAWDWorkspaceDefenseConnections(ctx, contestID, myTeam.ID, serviceIDs, serviceMap); err != nil {
		return nil, err
	}
	targetMap := make(map[int64]*AWDWorkspaceTargetTeamResult)
	for teamID, team := range teams {
		if teamID == myTeam.ID {
			continue
		}
		targetMap[teamID] = &AWDWorkspaceTargetTeamResult{
			TeamID:   teamID,
			TeamName: team.Name,
			Services: []*AWDWorkspaceTargetServiceResult{},
		}
	}

	instances, err := s.repo.ListServiceInstancesByContest(ctx, contestID, serviceIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	targetServiceSeen := make(map[int64]map[int64]struct{})
	for _, instance := range instances {
		if instance.TeamID == myTeam.ID {
			item := ensureAWDWorkspaceService(serviceMap, instance.ServiceID, instance.AWDChallengeID)
			if item.InstanceID == 0 {
				item.InstanceID = instance.InstanceID
				item.InstanceStatus = instance.Status
			}
			continue
		}

		target := targetMap[instance.TeamID]
		if target == nil {
			continue
		}
		seenServices := targetServiceSeen[instance.TeamID]
		if seenServices == nil {
			seenServices = make(map[int64]struct{})
			targetServiceSeen[instance.TeamID] = seenServices
		}
		if _, ok := seenServices[instance.ServiceID]; ok {
			continue
		}
		seenServices[instance.ServiceID] = struct{}{}
		target.Services = append(target.Services, &AWDWorkspaceTargetServiceResult{
			ServiceID:      instance.ServiceID,
			AWDChallengeID: instance.AWDChallengeID,
			Reachable:      instance.Status == model.InstanceStatusRunning && instance.AccessURL != "",
		})
	}
	if err := s.populateAWDWorkspaceLatestOperations(ctx, contestID, myTeam.ID, serviceMap); err != nil {
		return nil, err
	}

	if currentRound != nil {
		if err := s.populateAWDWorkspaceCurrentRound(ctx, currentRound.ID, myTeam.ID, teams, serviceMap, resp); err != nil {
			return nil, err
		}
	}

	resp.Services = sortAWDWorkspaceServices(serviceMap)
	resp.Targets = sortAWDWorkspaceTargets(targetMap)
	return resp, nil
}

func (s *AWDService) populateAWDWorkspaceLatestOperations(
	ctx context.Context,
	contestID, myTeamID int64,
	serviceMap map[int64]*AWDWorkspaceServiceResult,
) error {
	operations, err := s.repo.ListLatestServiceOperationsByContest(ctx, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	for _, operation := range operations {
		if operation.TeamID != myTeamID || operation.ServiceID <= 0 {
			continue
		}
		item := ensureAWDWorkspaceService(serviceMap, operation.ServiceID, 0)
		item.OperationStatus = operation.Status
		item.OperationType = operation.OperationType
		item.OperationReason = operation.Reason
		slaBillable := operation.SLABillable
		item.OperationSLABillable = &slaBillable
	}
	return nil
}

func (s *AWDService) populateAWDWorkspaceDefenseConnections(
	ctx context.Context,
	contestID, myTeamID int64,
	serviceIDs []int64,
	serviceMap map[int64]*AWDWorkspaceServiceResult,
) error {
	summaries, err := s.repo.ListDefenseWorkspaceSummariesByContestTeam(ctx, contestID, myTeamID, serviceIDs)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	for _, summary := range summaries {
		item := ensureAWDWorkspaceService(serviceMap, summary.ServiceID, 0)
		mergeAWDWorkspaceDefenseConnection(item, summary.Summary)
	}
	return nil
}

func (s *AWDService) populateAWDWorkspaceCurrentRound(
	ctx context.Context,
	roundID, myTeamID int64,
	teams map[int64]*model.Team,
	serviceMap map[int64]*AWDWorkspaceServiceResult,
	resp *AWDWorkspaceResult,
) error {
	records, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	for _, record := range records {
		if record.TeamID != myTeamID {
			continue
		}
		if record.ServiceID <= 0 {
			continue
		}
		item := ensureAWDWorkspaceService(serviceMap, record.ServiceID, record.AWDChallengeID)
		item.ServiceStatus = record.ServiceStatus
		item.CheckerType = string(record.CheckerType)
		item.AttackReceived = record.AttackReceived
		item.SLAScore = record.SLAScore
		item.DefenseScore = record.DefenseScore
		item.AttackScore = record.AttackScore
		updatedAt := record.UpdatedAt
		item.UpdatedAt = &updatedAt
	}

	logs, err := s.repo.ListAttackLogsByRound(ctx, roundID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	for _, log := range logs {
		if log.AttackerTeamID != myTeamID && log.VictimTeamID != myTeamID {
			continue
		}

		event := &AWDWorkspaceRecentEventResult{
			ID:             log.ID,
			ServiceID:      log.ServiceID,
			AWDChallengeID: log.AWDChallengeID,
			IsSuccess:      log.IsSuccess,
			ScoreGained:    log.ScoreGained,
			CreatedAt:      log.CreatedAt,
		}
		if log.AttackerTeamID == myTeamID {
			event.Direction = awdWorkspaceEventDirectionOutgoing
			event.PeerTeamID = log.VictimTeamID
			if team := teams[log.VictimTeamID]; team != nil {
				event.PeerTeamName = team.Name
			}
		} else {
			event.Direction = awdWorkspaceEventDirectionIncoming
			event.PeerTeamID = log.AttackerTeamID
			if team := teams[log.AttackerTeamID]; team != nil {
				event.PeerTeamName = team.Name
			}
		}
		resp.RecentEvents = append(resp.RecentEvents, event)
	}

	sort.Slice(resp.RecentEvents, func(i, j int) bool {
		if resp.RecentEvents[i].CreatedAt.Equal(resp.RecentEvents[j].CreatedAt) {
			return resp.RecentEvents[i].ID > resp.RecentEvents[j].ID
		}
		return resp.RecentEvents[i].CreatedAt.After(resp.RecentEvents[j].CreatedAt)
	})

	return nil
}

func ensureAWDWorkspaceService(items map[int64]*AWDWorkspaceServiceResult, serviceID, awdChallengeID int64) *AWDWorkspaceServiceResult {
	item := items[serviceID]
	if item != nil {
		return item
	}
	item = &AWDWorkspaceServiceResult{
		ServiceID:      serviceID,
		AWDChallengeID: awdChallengeID,
	}
	items[serviceID] = item
	return item
}

func mergeAWDWorkspaceDefenseConnection(item *AWDWorkspaceServiceResult, summary contestports.AWDDefenseWorkspaceSummary) {
	if item == nil {
		return
	}

	entryMode := strings.TrimSpace(summary.EntryMode)
	workspaceStatus := strings.TrimSpace(summary.WorkspaceStatus)
	if entryMode == "" && workspaceStatus == "" && summary.WorkspaceRevision <= 0 {
		return
	}

	if item.DefenseConnection == nil {
		item.DefenseConnection = &AWDDefenseConnectionResult{}
	}
	if entryMode != "" {
		item.DefenseConnection.EntryMode = entryMode
	}
	if workspaceStatus != "" {
		item.DefenseConnection.WorkspaceStatus = workspaceStatus
	}
	if summary.WorkspaceRevision > 0 {
		item.DefenseConnection.WorkspaceRevision = summary.WorkspaceRevision
	}
}

func sortAWDWorkspaceServices(items map[int64]*AWDWorkspaceServiceResult) []*AWDWorkspaceServiceResult {
	resp := make([]*AWDWorkspaceServiceResult, 0, len(items))
	for _, item := range items {
		resp = append(resp, item)
	}
	sort.Slice(resp, func(i, j int) bool {
		if resp[i].ServiceID == resp[j].ServiceID {
			return resp[i].AWDChallengeID < resp[j].AWDChallengeID
		}
		return resp[i].ServiceID < resp[j].ServiceID
	})
	return resp
}

func sortAWDWorkspaceTargets(items map[int64]*AWDWorkspaceTargetTeamResult) []*AWDWorkspaceTargetTeamResult {
	resp := make([]*AWDWorkspaceTargetTeamResult, 0, len(items))
	for _, item := range items {
		sort.Slice(item.Services, func(i, j int) bool {
			if item.Services[i].ServiceID == item.Services[j].ServiceID {
				return item.Services[i].AWDChallengeID < item.Services[j].AWDChallengeID
			}
			return item.Services[i].ServiceID < item.Services[j].ServiceID
		})
		resp = append(resp, item)
	}
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].TeamID < resp[j].TeamID
	})
	return resp
}
