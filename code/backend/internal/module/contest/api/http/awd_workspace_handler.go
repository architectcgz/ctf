package http

import (
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetUserWorkspace(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	contestID := c.GetInt64("id")
	resp, err := h.queries.GetUserWorkspace(c.Request.Context(), currentUser.UserID, contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdWorkspaceResultToDTO(resp))
}

func awdWorkspaceResultToDTO(item *contestqry.AWDWorkspaceResult) *dto.ContestAWDWorkspaceResp {
	if item == nil {
		return nil
	}
	result := &dto.ContestAWDWorkspaceResp{
		ContestID:    item.ContestID,
		Services:     make([]*dto.ContestAWDWorkspaceServiceResp, 0, len(item.Services)),
		Targets:      make([]*dto.ContestAWDWorkspaceTargetTeamResp, 0, len(item.Targets)),
		RecentEvents: make([]*dto.ContestAWDWorkspaceRecentEventResp, 0, len(item.RecentEvents)),
	}
	if item.CurrentRound != nil {
		result.CurrentRound = &dto.AWDRoundResp{
			ID:           item.CurrentRound.ID,
			ContestID:    item.CurrentRound.ContestID,
			RoundNumber:  item.CurrentRound.RoundNumber,
			Status:       item.CurrentRound.Status,
			StartedAt:    item.CurrentRound.StartedAt,
			EndedAt:      item.CurrentRound.EndedAt,
			AttackScore:  item.CurrentRound.AttackScore,
			DefenseScore: item.CurrentRound.DefenseScore,
			CreatedAt:    item.CurrentRound.CreatedAt,
			UpdatedAt:    item.CurrentRound.UpdatedAt,
		}
	}
	if item.MyTeam != nil {
		result.MyTeam = &dto.ContestAWDWorkspaceTeamResp{
			TeamID:   item.MyTeam.TeamID,
			TeamName: item.MyTeam.TeamName,
		}
	}
	for _, service := range item.Services {
		if service == nil {
			result.Services = append(result.Services, nil)
			continue
		}
		result.Services = append(result.Services, &dto.ContestAWDWorkspaceServiceResp{
			ServiceID:            service.ServiceID,
			AWDChallengeID:       service.AWDChallengeID,
			InstanceID:           service.InstanceID,
			InstanceStatus:       service.InstanceStatus,
			AccessURL:            service.AccessURL,
			ServiceStatus:        service.ServiceStatus,
			OperationStatus:      service.OperationStatus,
			OperationType:        service.OperationType,
			OperationReason:      service.OperationReason,
			OperationSLABillable: service.OperationSLABillable,
			CheckerType:          model.AWDCheckerType(service.CheckerType),
			AttackReceived:       service.AttackReceived,
			SLAScore:             service.SLAScore,
			DefenseScore:         service.DefenseScore,
			AttackScore:          service.AttackScore,
			UpdatedAt:            service.UpdatedAt,
		})
	}
	for _, target := range item.Targets {
		if target == nil {
			result.Targets = append(result.Targets, nil)
			continue
		}
		next := &dto.ContestAWDWorkspaceTargetTeamResp{
			TeamID:   target.TeamID,
			TeamName: target.TeamName,
			Services: make([]*dto.ContestAWDWorkspaceTargetServiceResp, 0, len(target.Services)),
		}
		for _, service := range target.Services {
			if service == nil {
				next.Services = append(next.Services, nil)
				continue
			}
			next.Services = append(next.Services, &dto.ContestAWDWorkspaceTargetServiceResp{
				ServiceID:      service.ServiceID,
				AWDChallengeID: service.AWDChallengeID,
				Reachable:      service.Reachable,
			})
		}
		result.Targets = append(result.Targets, next)
	}
	for _, event := range item.RecentEvents {
		if event == nil {
			result.RecentEvents = append(result.RecentEvents, nil)
			continue
		}
		result.RecentEvents = append(result.RecentEvents, &dto.ContestAWDWorkspaceRecentEventResp{
			ID:             event.ID,
			Direction:      event.Direction,
			ServiceID:      event.ServiceID,
			AWDChallengeID: event.AWDChallengeID,
			PeerTeamID:     event.PeerTeamID,
			PeerTeamName:   event.PeerTeamName,
			IsSuccess:      event.IsSuccess,
			ScoreGained:    event.ScoreGained,
			CreatedAt:      event.CreatedAt,
		})
	}
	return result
}
