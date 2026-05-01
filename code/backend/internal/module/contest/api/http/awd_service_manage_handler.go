package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) ListContestAWDServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.serviceQueries.ListContestAWDServices(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestAWDServiceResultsToDTO(resp))
}

func contestAWDServiceResultsToDTO(results []contestqry.ContestAWDServiceResult) []*dto.ContestAWDServiceResp {
	resp := make([]*dto.ContestAWDServiceResp, 0, len(results))
	for i := range results {
		item := results[i]
		resp = append(resp, &dto.ContestAWDServiceResp{
			ID:                item.ID,
			ContestID:         item.ContestID,
			AWDChallengeID:    item.AWDChallengeID,
			Title:             item.Title,
			Category:          item.Category,
			Difficulty:        item.Difficulty,
			DisplayName:       item.DisplayName,
			Order:             item.Order,
			IsVisible:         item.IsVisible,
			ScoreConfig:       item.ScoreConfig,
			RuntimeConfig:     item.RuntimeConfig,
			ValidationState:   model.AWDCheckerValidationState(item.ValidationState),
			LastPreviewAt:     item.LastPreviewAt,
			LastPreviewResult: awdCheckerPreviewResultToDTO(contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResultRaw)),
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		})
	}
	return resp
}

func awdCheckerPreviewResultToDTO(item *contestdomain.AWDCheckerPreviewResult) *dto.AWDCheckerPreviewResp {
	if item == nil {
		return nil
	}
	return &dto.AWDCheckerPreviewResp{
		CheckerType:   item.CheckerType,
		ServiceStatus: item.ServiceStatus,
		CheckResult:   item.CheckResult,
		PreviewContext: dto.AWDCheckerPreviewContextResp{
			ServiceID:      item.PreviewContext.ServiceID,
			AccessURL:      item.PreviewContext.AccessURL,
			PreviewFlag:    item.PreviewContext.PreviewFlag,
			RoundNumber:    item.PreviewContext.RoundNumber,
			TeamID:         item.PreviewContext.TeamID,
			AWDChallengeID: item.PreviewContext.AWDChallengeID,
		},
		PreviewToken: item.PreviewToken,
	}
}

func (h *AWDHandler) CreateContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateContestAWDServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.serviceCommands.CreateContestAWDService(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) UpdateContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	serviceID := c.GetInt64("sid")
	var req dto.UpdateContestAWDServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.serviceCommands.UpdateContestAWDService(c.Request.Context(), contestID, serviceID, &req); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *AWDHandler) DeleteContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	serviceID := c.GetInt64("sid")
	if err := h.serviceCommands.DeleteContestAWDService(c.Request.Context(), contestID, serviceID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
