package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type TopologyHandler struct {
	commands topologyCommandService
	queries  topologyQueryService
}

type topologyCommandService interface {
	SaveChallengeTopology(ctx context.Context, challengeID int64, req *dto.SaveChallengeTopologyReq) (*dto.ChallengeTopologyResp, error)
	DeleteChallengeTopology(ctx context.Context, challengeID int64) error
	CreateTemplate(ctx context.Context, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	UpdateTemplate(ctx context.Context, id int64, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	DeleteTemplate(ctx context.Context, id int64) error
}

type topologyQueryService interface {
	GetChallengeTopology(ctx context.Context, challengeID int64) (*dto.ChallengeTopologyResp, error)
	GetTemplate(ctx context.Context, id int64) (*dto.EnvironmentTemplateResp, error)
	ListTemplates(ctx context.Context, keyword string) ([]*dto.EnvironmentTemplateResp, error)
}

func NewTopologyHandler(commands topologyCommandService, queries topologyQueryService) *TopologyHandler {
	return &TopologyHandler{commands: commands, queries: queries}
}

func (h *TopologyHandler) SaveChallengeTopology(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	var req dto.SaveChallengeTopologyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.SaveChallengeTopology(c.Request.Context(), challengeID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) GetChallengeTopology(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.GetChallengeTopology(c.Request.Context(), challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) DeleteChallengeTopology(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	if err := h.commands.DeleteChallengeTopology(c.Request.Context(), challengeID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *TopologyHandler) CreateTemplate(c *gin.Context) {
	var req dto.UpsertEnvironmentTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 template id")
		return
	}
	var req dto.UpsertEnvironmentTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.UpdateTemplate(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 template id")
		return
	}
	resp, err := h.queries.GetTemplate(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) ListTemplates(c *gin.Context) {
	resp, err := h.queries.ListTemplates(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 template id")
		return
	}
	if err := h.commands.DeleteTemplate(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
