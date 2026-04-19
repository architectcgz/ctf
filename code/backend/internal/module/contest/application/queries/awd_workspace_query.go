package queries

import (
	"context"
	"errors"
	"sort"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
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

	challenges, err := s.repo.ListChallengesByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	challengeIDs := make([]int64, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}

	serviceMap := make(map[int64]*dto.ContestAWDWorkspaceServiceResp)
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

	instances, err := s.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	for _, instance := range instances {
		if instance.TeamID == myTeam.ID {
			item := ensureAWDWorkspaceService(serviceMap, instance.ServiceID, instance.ChallengeID)
			if item.AccessURL == "" {
				item.AccessURL = instance.AccessURL
			}
			continue
		}

		target := targetMap[instance.TeamID]
		if target == nil {
			continue
		}
		target.Services = append(target.Services, &dto.ContestAWDWorkspaceTargetServiceResp{
			ServiceID:   instance.ServiceID,
			ChallengeID: instance.ChallengeID,
			AccessURL:   instance.AccessURL,
		})
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
		item := ensureAWDWorkspaceService(serviceMap, record.ServiceID, record.ChallengeID)
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
			ID:          log.ID,
			ServiceID:   log.ServiceID,
			ChallengeID: log.ChallengeID,
			IsSuccess:   log.IsSuccess,
			ScoreGained: log.ScoreGained,
			CreatedAt:   log.CreatedAt,
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

func ensureAWDWorkspaceService(items map[int64]*dto.ContestAWDWorkspaceServiceResp, serviceID, challengeID int64) *dto.ContestAWDWorkspaceServiceResp {
	item := items[serviceID]
	if item != nil {
		return item
	}
	item = &dto.ContestAWDWorkspaceServiceResp{
		ServiceID:   serviceID,
		ChallengeID: challengeID,
	}
	items[serviceID] = item
	return item
}

func sortAWDWorkspaceServices(items map[int64]*dto.ContestAWDWorkspaceServiceResp) []*dto.ContestAWDWorkspaceServiceResp {
	resp := make([]*dto.ContestAWDWorkspaceServiceResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, item)
	}
	sort.Slice(resp, func(i, j int) bool {
		if resp[i].ServiceID == resp[j].ServiceID {
			return resp[i].ChallengeID < resp[j].ChallengeID
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
				if item.Services[i].ChallengeID == item.Services[j].ChallengeID {
					return item.Services[i].AccessURL < item.Services[j].AccessURL
				}
				return item.Services[i].ChallengeID < item.Services[j].ChallengeID
			}
			if item.Services[i].ChallengeID == item.Services[j].ChallengeID {
				return item.Services[i].AccessURL < item.Services[j].AccessURL
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
