package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FlagHandler struct {
	flagService flagService
}

type flagService interface {
	ConfigureStaticFlag(challengeID int64, flag, flagPrefix string) error
	ConfigureDynamicFlag(challengeID int64, flagPrefix string) error
	GetFlagConfig(challengeID int64) (*dto.FlagResp, error)
}

func NewFlagHandler(flagService flagService) *FlagHandler {
	return &FlagHandler{flagService: flagService}
}

// ConfigureFlag 配置 Flag
// PUT /api/v1/admin/challenges/:id/flag
func (h *FlagHandler) ConfigureFlag(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.ConfigureFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.FlagType == model.FlagTypeStatic {
		err = h.flagService.ConfigureStaticFlag(challengeID, req.Flag, req.FlagPrefix)
	} else {
		err = h.flagService.ConfigureDynamicFlag(challengeID, req.FlagPrefix)
	}

	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Flag 配置成功"})
}

// GetFlagConfig 获取 Flag 配置
// GET /api/v1/admin/challenges/:id/flag
func (h *FlagHandler) GetFlagConfig(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	flagResp, err := h.flagService.GetFlagConfig(challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, flagResp)
}
