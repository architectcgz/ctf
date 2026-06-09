package http

import (
	"context"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

type AWDChallengeHandler struct {
	commands        awdChallengeCommandService
	queries         awdChallengeQueryService
	packageDelivery packageDeliveryService
}

type awdChallengeCommandService interface {
	CreateChallenge(ctx context.Context, actorUserID int64, req challengecmd.CreateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error)
	UpdateChallenge(ctx context.Context, id int64, req challengecmd.UpdateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error)
	DeleteChallenge(ctx context.Context, id int64) error
	PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.AWDChallengeImportPreviewResp, error)
	ListImports(ctx context.Context, actorUserID int64) ([]challengecontracts.AWDChallengeImportPreviewResp, error)
	GetImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeImportPreviewResp, error)
	CommitImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeResp, error)
}

type awdChallengeQueryService interface {
	GetChallenge(ctx context.Context, id int64) (*challengecontracts.AWDChallengeResp, error)
	ListChallenges(ctx context.Context, req challengeqry.ListAWDChallengesInput) (*challengecontracts.AWDChallengePageResp, error)
}

func NewAWDChallengeHandler(commands awdChallengeCommandService, queries awdChallengeQueryService) *AWDChallengeHandler {
	return &AWDChallengeHandler{
		commands:        commands,
		queries:         queries,
		packageDelivery: challengecmd.NewPackageDeliveryService(nil, commands),
	}
}

func (h *AWDChallengeHandler) CreateChallenge(c *gin.Context) {
	var req CreateAWDChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.CreateChallenge(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeRequestMapper.ToCreateAWDChallengeInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengeResp(resp))
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
	response.Success(c, toAWDChallengeResp(resp))
}

func (h *AWDChallengeHandler) ListChallenges(c *gin.Context) {
	var req AWDChallengeQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.queries.ListChallenges(c.Request.Context(), challengeRequestMapper.ToListAWDChallengesInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengePageResult(resp))
}

func (h *AWDChallengeHandler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 AWD Challenge ID")
		return
	}
	var req UpdateAWDChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.UpdateChallenge(c.Request.Context(), id, challengeRequestMapper.ToUpdateAWDChallengeInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengeResp(resp))
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

	result, err := h.packageDelivery.Preview(c.Request.Context(), challengecmd.PackageDeliveryPreviewRequest{
		Mode:        challengecmd.PackageDeliveryModeAWD,
		ActorUserID: authctx.MustCurrentUser(c).UserID,
		FileName:    fileHeader.Filename,
		Reader:      file,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, nethttp.StatusCreated, toAWDChallengeImportPreviewResp(result.AWD))
}

func (h *AWDChallengeHandler) ListImports(c *gin.Context) {
	resp, err := h.commands.ListImports(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengeImportPreviewRespList(resp))
}

func (h *AWDChallengeHandler) GetImport(c *gin.Context) {
	resp, err := h.commands.GetImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengeImportPreviewResp(resp))
}

func (h *AWDChallengeHandler) CommitImport(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.InvalidParams(c, "无效的导入 ID")
		return
	}

	result, err := h.packageDelivery.Commit(c.Request.Context(), challengecmd.PackageDeliveryCommitRequest{
		Mode:        challengecmd.PackageDeliveryModeAWD,
		ActorUserID: authctx.MustCurrentUser(c).UserID,
		PreviewID:   id,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toAWDChallengeImportCommitResp(result.AWD))
}
