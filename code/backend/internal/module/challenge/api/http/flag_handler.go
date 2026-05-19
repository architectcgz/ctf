package http

import (
	"context"
	"strconv"

	"ctf-platform/internal/apperror"
	response "ctf-platform/internal/httpresponse"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"

	"github.com/gin-gonic/gin"
)

type FlagHandler struct {
	commands flagCommandService
	queries  flagQueryService
}

type flagCommandService interface {
	ConfigureStaticFlag(ctx context.Context, challengeID int64, flag, flagPrefix string) error
	ConfigureDynamicFlag(ctx context.Context, challengeID int64, flagPrefix string) error
	ConfigureRegexFlag(ctx context.Context, challengeID int64, flagRegex, flagPrefix string) error
	ConfigureManualReviewFlag(ctx context.Context, challengeID int64) error
}

type flagQueryService interface {
	GetFlagConfig(ctx context.Context, challengeID int64) (*challengecontracts.FlagResp, error)
}

const (
	flagTypeStatic       = "static"
	flagTypeDynamic      = "dynamic"
	flagTypeRegex        = "regex"
	flagTypeManualReview = "manual_review"
)

func NewFlagHandler(commands flagCommandService, queries flagQueryService) *FlagHandler {
	return &FlagHandler{commands: commands, queries: queries}
}

// ConfigureFlag 配置 Flag
// PUT /api/v1/admin/challenges/:id/flag
func (h *FlagHandler) ConfigureFlag(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req ConfigureFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.FlagType == flagTypeStatic {
		err = h.commands.ConfigureStaticFlag(c.Request.Context(), challengeID, req.Flag, req.FlagPrefix)
	} else if req.FlagType == flagTypeDynamic {
		err = h.commands.ConfigureDynamicFlag(c.Request.Context(), challengeID, req.FlagPrefix)
	} else if req.FlagType == flagTypeRegex {
		err = h.commands.ConfigureRegexFlag(c.Request.Context(), challengeID, req.FlagRegex, req.FlagPrefix)
	} else {
		err = h.commands.ConfigureManualReviewFlag(c.Request.Context(), challengeID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	flagResp, err := h.queries.GetFlagConfig(c.Request.Context(), challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, toFlagResp(flagResp))
}
