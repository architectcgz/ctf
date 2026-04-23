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
	SaveChallengeTopology(challengeID int64, req *dto.SaveChallengeTopologyReq) (*dto.ChallengeTopologyResp, error)
	DeleteChallengeTopology(challengeID int64) error
	CreateTemplate(req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	UpdateTemplate(id int64, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	DeleteTemplate(id int64) error
}

type topologyQueryService interface {
	GetChallengeTopology(challengeID int64) (*dto.ChallengeTopologyResp, error)
	GetChallengeTopologyWithContext(ctx context.Context, challengeID int64) (*dto.ChallengeTopologyResp, error)
	GetTemplate(id int64) (*dto.EnvironmentTemplateResp, error)
	GetTemplateWithContext(ctx context.Context, id int64) (*dto.EnvironmentTemplateResp, error)
	ListTemplates(keyword string) ([]*dto.EnvironmentTemplateResp, error)
	ListTemplatesWithContext(ctx context.Context, keyword string) ([]*dto.EnvironmentTemplateResp, error)
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
	resp, err := h.commands.SaveChallengeTopology(challengeID, &req)
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
	resp, err := h.queries.GetChallengeTopologyWithContext(c.Request.Context(), challengeID)
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
	if err := h.commands.DeleteChallengeTopology(challengeID); err != nil {
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
	resp, err := h.commands.CreateTemplate(&req)
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
	resp, err := h.commands.UpdateTemplate(id, &req)
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
	resp, err := h.queries.GetTemplateWithContext(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) ListTemplates(c *gin.Context) {
	resp, err := h.queries.ListTemplatesWithContext(c.Request.Context(), c.Query("keyword"))
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
	if err := h.commands.DeleteTemplate(id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
