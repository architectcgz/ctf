package http

import (
	"ctf-platform/internal/dto"
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
		mapped := contestRequestMapper.ToContestAWDServiceResp(item)
		mapped.LastPreviewResult = awdCheckerPreviewResultToDTO(contestdomain.ParseAWDCheckerPreviewResult(item.LastPreviewResultRaw))
		resp = append(resp, &mapped)
	}
	return resp
}

func awdCheckerPreviewResultToDTO(item *contestdomain.AWDCheckerPreviewResult) *dto.AWDCheckerPreviewResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDCheckerPreviewResp(*item)
	return &mapped
}

func (h *AWDHandler) CreateContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateContestAWDServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	input := contestRequestMapper.ToCreateContestAWDServiceInput(req)

	resp, err := h.serviceCommands.CreateContestAWDService(c.Request.Context(), contestID, input)
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
	input := contestRequestMapper.ToUpdateContestAWDServiceInput(req)

	if err := h.serviceCommands.UpdateContestAWDService(c.Request.Context(), contestID, serviceID, input); err != nil {
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
