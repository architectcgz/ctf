package queries

import (
	"context"
	"errors"
	"sort"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

const (
	awdWorkspaceEventDirectionOutgoing = "attack_out"
	awdWorkspaceEventDirectionIncoming = "attack_in"
)

func (s *AWDService) GetUserWorkspace(ctx context.Context, userID, contestID int64) (*dto.ContestAWDWorkspaceResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	resp := &dto.ContestAWDWorkspaceResp{
		ContestID:    contestID,
		Services:     []*dto.ContestAWDWorkspaceServiceResp{},
		Targets:      []*dto.ContestAWDWorkspaceTargetTeamResp{},
		RecentEvents: []*dto.ContestAWDWorkspaceRecentEventResp{},
	}

	currentRound, err := s.repo.FindRunningRound(ctx, contestID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp.CurrentRound = contestdomain.AWDRoundRespFromModel(currentRound)

	myTeam, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if myTeam == nil {
		return resp, nil
	}
	resp.MyTeam = &dto.ContestAWDWorkspaceTeamResp{
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

	serviceMap := make(map[int64]*dto.ContestAWDWorkspaceServiceResp)
	serviceIDs := make([]int64, 0, len(definitions))
	for _, definition := range definitions {
		serviceIDs = append(serviceIDs, definition.ServiceID)
		item := ensureAWDWorkspaceService(serviceMap, definition.ServiceID, definition.AWDChallengeID)
		item.DefenseScope = toAWDWorkspaceDefenseScope(definition.DefenseScope)
	}
	targetMap := make(map[int64]*dto.ContestAWDWorkspaceTargetTeamResp)
	for teamID, team := range teams {
		if teamID == myTeam.ID {
			continue
		}
		targetMap[teamID] = &dto.ContestAWDWorkspaceTargetTeamResp{
			TeamID:   teamID,
			TeamName: team.Name,
			Services: []*dto.ContestAWDWorkspaceTargetServiceResp{},
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
		target.Services = append(target.Services, &dto.ContestAWDWorkspaceTargetServiceResp{
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
	serviceMap map[int64]*dto.ContestAWDWorkspaceServiceResp,
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

func (s *AWDService) populateAWDWorkspaceCurrentRound(
	ctx context.Context,
	roundID, myTeamID int64,
	teams map[int64]*model.Team,
	serviceMap map[int64]*dto.ContestAWDWorkspaceServiceResp,
	resp *dto.ContestAWDWorkspaceResp,
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
		item.CheckerType = record.CheckerType
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

		event := &dto.ContestAWDWorkspaceRecentEventResp{
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

func ensureAWDWorkspaceService(items map[int64]*dto.ContestAWDWorkspaceServiceResp, serviceID, awdChallengeID int64) *dto.ContestAWDWorkspaceServiceResp {
	item := items[serviceID]
	if item != nil {
		return item
	}
	item = &dto.ContestAWDWorkspaceServiceResp{
		ServiceID:      serviceID,
		AWDChallengeID: awdChallengeID,
	}
	items[serviceID] = item
	return item
}

func toAWDWorkspaceDefenseScope(scope contestports.AWDDefenseScope) *dto.AWDDefenseScopeResp {
	if len(scope.EditablePaths) == 0 && len(scope.ProtectedPaths) == 0 && len(scope.ServiceContracts) == 0 {
		return nil
	}
	return &dto.AWDDefenseScopeResp{
		EditablePaths:    scope.EditablePaths,
		ProtectedPaths:   scope.ProtectedPaths,
		ServiceContracts: scope.ServiceContracts,
	}
}

func sortAWDWorkspaceServices(items map[int64]*dto.ContestAWDWorkspaceServiceResp) []*dto.ContestAWDWorkspaceServiceResp {
	resp := make([]*dto.ContestAWDWorkspaceServiceResp, 0, len(items))
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

func sortAWDWorkspaceTargets(items map[int64]*dto.ContestAWDWorkspaceTargetTeamResp) []*dto.ContestAWDWorkspaceTargetTeamResp {
	resp := make([]*dto.ContestAWDWorkspaceTargetTeamResp, 0, len(items))
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
