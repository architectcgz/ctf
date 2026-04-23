package http

import (
	"context"
	"strconv"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type FlagHandler struct {
	commands flagCommandService
	queries  flagQueryService
}

type flagCommandService interface {
	ConfigureStaticFlag(challengeID int64, flag, flagPrefix string) error
	ConfigureStaticFlagWithContext(ctx context.Context, challengeID int64, flag, flagPrefix string) error
	ConfigureDynamicFlag(challengeID int64, flagPrefix string) error
	ConfigureDynamicFlagWithContext(ctx context.Context, challengeID int64, flagPrefix string) error
	ConfigureRegexFlag(challengeID int64, flagRegex, flagPrefix string) error
	ConfigureRegexFlagWithContext(ctx context.Context, challengeID int64, flagRegex, flagPrefix string) error
	ConfigureManualReviewFlag(challengeID int64) error
	ConfigureManualReviewFlagWithContext(ctx context.Context, challengeID int64) error
}

type flagQueryService interface {
	GetFlagConfig(challengeID int64) (*dto.FlagResp, error)
	GetFlagConfigWithContext(ctx context.Context, challengeID int64) (*dto.FlagResp, error)
}

func NewFlagHandler(commands flagCommandService, queries flagQueryService) *FlagHandler {
	return &FlagHandler{commands: commands, queries: queries}
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
		err = h.commands.ConfigureStaticFlagWithContext(c.Request.Context(), challengeID, req.Flag, req.FlagPrefix)
	} else if req.FlagType == model.FlagTypeDynamic {
		err = h.commands.ConfigureDynamicFlagWithContext(c.Request.Context(), challengeID, req.FlagPrefix)
	} else if req.FlagType == model.FlagTypeRegex {
		err = h.commands.ConfigureRegexFlagWithContext(c.Request.Context(), challengeID, req.FlagRegex, req.FlagPrefix)
	} else {
		err = h.commands.ConfigureManualReviewFlagWithContext(c.Request.Context(), challengeID)
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

	flagResp, err := h.queries.GetFlagConfigWithContext(c.Request.Context(), challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, flagResp)
}
