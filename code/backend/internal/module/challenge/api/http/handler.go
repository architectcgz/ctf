package http

import (
	"context"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service challengeService
}

type challengeService interface {
	CreateChallenge(req *dto.CreateChallengeReq) (*dto.ChallengeResp, error)
	UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error
	DeleteChallenge(id int64) error
	GetChallenge(id int64) (*dto.ChallengeResp, error)
	ListChallenges(query *dto.ChallengeQuery) (*dto.PageResult, error)
	PublishChallenge(id int64) error
	ListPublishedChallengesWithContext(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error)
	GetPublishedChallengeWithContext(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error)
}

func NewHandler(service challengeService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateChallenge(c *gin.Context) {
	var req dto.CreateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateChallenge(&req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	var req dto.UpdateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.service.UpdateChallenge(id, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	if err := h.service.DeleteChallenge(id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	resp, err := h.service.GetChallenge(id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) ListChallenges(c *gin.Context) {
	var query dto.ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.service.ListChallenges(&query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *Handler) PublishChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	if err := h.service.PublishChallenge(id); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListPublishedChallenges 靶场列表（学员视图）
func (h *Handler) ListPublishedChallenges(c *gin.Context) {
	var query dto.ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.service.ListPublishedChallengesWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

// GetPublishedChallenge 靶场详情（学员视图）
func (h *Handler) GetPublishedChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	detail, err := h.service.GetPublishedChallengeWithContext(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, detail)
}

// DownloadAttachment 下载导入题包中的附件文件。
func (h *Handler) DownloadAttachment(c *gin.Context) {
	relativePath := strings.TrimSpace(strings.TrimPrefix(c.Param("path"), "/"))
	if relativePath == "" {
		response.InvalidParams(c, "无效的附件路径")
		return
	}

	cleanPath := filepath.ToSlash(filepath.Clean(relativePath))
	if cleanPath == "." || strings.HasPrefix(cleanPath, "../") || strings.Contains(cleanPath, "/../") {
		response.InvalidParams(c, "无效的附件路径")
		return
	}

	baseDir := strings.TrimSpace(os.Getenv("CHALLENGE_PACKS_DIR"))
	if baseDir == "" {
		baseDir = "../../docs/challenges/packs"
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		response.FromError(c, err)
		return
	}

	target := filepath.Clean(filepath.Join(baseAbs, filepath.FromSlash(cleanPath)))
	prefix := baseAbs + string(os.PathSeparator)
	if target != baseAbs && !strings.HasPrefix(target, prefix) {
		response.InvalidParams(c, "无效的附件路径")
		return
	}

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			response.Error(c, errcode.ErrNotFound)
			return
		}
		response.FromError(c, err)
		return
	}
	if info.IsDir() {
		response.Error(c, errcode.ErrNotFound)
		return
	}

	c.FileAttachment(target, filepath.Base(target))
}
