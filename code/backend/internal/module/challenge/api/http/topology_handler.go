package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type TopologyHandler struct {
	service topologyService
}

type topologyService interface {
	SaveChallengeTopology(challengeID int64, req *dto.SaveChallengeTopologyReq) (*dto.ChallengeTopologyResp, error)
	GetChallengeTopology(challengeID int64) (*dto.ChallengeTopologyResp, error)
	DeleteChallengeTopology(challengeID int64) error
	CreateTemplate(req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	UpdateTemplate(id int64, req *dto.UpsertEnvironmentTemplateReq) (*dto.EnvironmentTemplateResp, error)
	GetTemplate(id int64) (*dto.EnvironmentTemplateResp, error)
	ListTemplates(keyword string) ([]*dto.EnvironmentTemplateResp, error)
	DeleteTemplate(id int64) error
}

func NewTopologyHandler(service topologyService) *TopologyHandler {
	return &TopologyHandler{service: service}
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
	resp, err := h.service.SaveChallengeTopology(challengeID, &req)
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
	resp, err := h.service.GetChallengeTopology(challengeID)
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
	if err := h.service.DeleteChallengeTopology(challengeID); err != nil {
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
	resp, err := h.service.CreateTemplate(&req)
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
	resp, err := h.service.UpdateTemplate(id, &req)
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
	resp, err := h.service.GetTemplate(id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *TopologyHandler) ListTemplates(c *gin.Context) {
	resp, err := h.service.ListTemplates(c.Query("keyword"))
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
	if err := h.service.DeleteTemplate(id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
