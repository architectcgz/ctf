package http

import (
	"context"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type AWDChallengeHandler struct {
	commands awdChallengeCommandService
	queries  awdChallengeQueryService
}

type awdChallengeCommandService interface {
	CreateChallenge(ctx context.Context, actorUserID int64, req *dto.CreateAWDChallengeReq) (*dto.AWDChallengeResp, error)
	UpdateChallenge(ctx context.Context, id int64, req *dto.UpdateAWDChallengeReq) (*dto.AWDChallengeResp, error)
	DeleteChallenge(ctx context.Context, id int64) error
	PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.AWDChallengeImportPreviewResp, error)
	ListImports(ctx context.Context, actorUserID int64) ([]dto.AWDChallengeImportPreviewResp, error)
	GetImport(ctx context.Context, actorUserID int64, id string) (*dto.AWDChallengeImportPreviewResp, error)
	CommitImport(ctx context.Context, actorUserID int64, id string) (*dto.AWDChallengeResp, error)
}

type awdChallengeQueryService interface {
	GetChallenge(ctx context.Context, id int64) (*dto.AWDChallengeResp, error)
	ListChallenges(ctx context.Context, req *dto.AWDChallengeQuery) (*dto.AWDChallengePageResp, error)
}

func NewAWDChallengeHandler(commands awdChallengeCommandService, queries awdChallengeQueryService) *AWDChallengeHandler {
	return &AWDChallengeHandler{commands: commands, queries: queries}
}

func (h *AWDChallengeHandler) CreateChallenge(c *gin.Context) {
	var req dto.CreateAWDChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.CreateChallenge(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) GetChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Challenge ID")
		return
	}
	resp, err := h.queries.GetChallenge(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) ListChallenges(c *gin.Context) {
	var req dto.AWDChallengeQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.queries.ListChallenges(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Challenge ID")
		return
	}
	var req dto.UpdateAWDChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.UpdateChallenge(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) DeleteChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Challenge ID")
		return
	}
	if err := h.commands.DeleteChallenge(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *AWDChallengeHandler) PreviewImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.InvalidParams(c, "缺少 AWD 题目包文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InvalidParams(c, "无法读取 AWD 题目包文件")
		return
	}
	defer file.Close()

	resp, err := h.commands.PreviewImport(
		c.Request.Context(),
		authctx.MustCurrentUser(c).UserID,
		fileHeader.Filename,
		file,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, nethttp.StatusCreated, resp)
}

func (h *AWDChallengeHandler) ListImports(c *gin.Context) {
	resp, err := h.commands.ListImports(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) GetImport(c *gin.Context) {
	resp, err := h.commands.GetImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDChallengeHandler) CommitImport(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.InvalidParams(c, "无效的导入 ID")
		return
	}

	resp, err := h.commands.CommitImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, &dto.AWDChallengeImportCommitResp{Challenge: resp})
}
