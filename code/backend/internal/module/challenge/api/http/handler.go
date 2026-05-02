package http

import (
	"context"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
	"fmt"
	"io"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	commands challengeCommandService
	queries  challengeQueryService
}

type challengeCommandService interface {
	CreateChallenge(ctx context.Context, actorUserID int64, req challengecmd.CreateChallengeInput) (*dto.ChallengeResp, error)
	UpdateChallenge(ctx context.Context, id int64, req challengecmd.UpdateChallengeInput) error
	DeleteChallenge(ctx context.Context, id int64) error
	RequestPublishCheck(ctx context.Context, actorUserID, id int64) (*dto.ChallengePublishCheckJobResp, error)
	GetLatestPublishCheck(ctx context.Context, id int64) (*dto.ChallengePublishCheckJobResp, error)
	SelfCheckChallenge(ctx context.Context, id int64) (*dto.ChallengeSelfCheckResp, error)
	PreviewChallengeImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.ChallengeImportPreviewResp, error)
	ListChallengeImports(ctx context.Context, actorUserID int64) ([]dto.ChallengeImportPreviewResp, error)
	GetChallengeImport(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error)
	CommitChallengeImport(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeResp, error)
	ExportChallengePackage(ctx context.Context, actorUserID int64, challengeID int64) (*dto.ChallengePackageExportResp, error)
	GetChallengePackageExport(ctx context.Context, challengeID int64, revisionID *int64) (*dto.ChallengePackageExportResp, error)
}

type challengeQueryService interface {
	GetChallenge(ctx context.Context, id int64) (*dto.ChallengeResp, error)
	ListChallenges(ctx context.Context, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeResp], error)
	ListPublishedChallenges(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeListItem], error)
	GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error)
}

func NewHandler(commands challengeCommandService, queries challengeQueryService) *Handler {
	return &Handler{commands: commands, queries: queries}
}

func (h *Handler) CreateChallenge(c *gin.Context) {
	var req dto.CreateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateChallenge(c.Request.Context(), authctx.MustCurrentUser(c).UserID, createChallengeInputFromDTO(&req))
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

	if err := h.commands.UpdateChallenge(c.Request.Context(), id, updateChallengeInputFromDTO(&req)); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func createChallengeInputFromDTO(req *dto.CreateChallengeReq) challengecmd.CreateChallengeInput {
	if req == nil {
		return challengecmd.CreateChallengeInput{}
	}
	hints := make([]challengecmd.ChallengeHintInput, 0, len(req.Hints))
	for _, hint := range req.Hints {
		hints = append(hints, challengecmd.ChallengeHintInput{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}
	return challengecmd.CreateChallengeInput{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Difficulty:      req.Difficulty,
		Points:          req.Points,
		ImageID:         req.ImageID,
		AttachmentURL:   req.AttachmentURL,
		InstanceSharing: req.InstanceSharing,
		Hints:           hints,
	}
}

func updateChallengeInputFromDTO(req *dto.UpdateChallengeReq) challengecmd.UpdateChallengeInput {
	if req == nil {
		return challengecmd.UpdateChallengeInput{}
	}
	hints := make([]challengecmd.ChallengeHintInput, 0, len(req.Hints))
	for _, hint := range req.Hints {
		hints = append(hints, challengecmd.ChallengeHintInput{
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		})
	}
	return challengecmd.UpdateChallengeInput{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Difficulty:      req.Difficulty,
		Points:          req.Points,
		ImageID:         req.ImageID,
		AttachmentURL:   req.AttachmentURL,
		InstanceSharing: req.InstanceSharing,
		Hints:           hints,
	}
}

func (h *Handler) DeleteChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	if err := h.commands.DeleteChallenge(c.Request.Context(), id); err != nil {
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

	resp, err := h.queries.GetChallenge(c.Request.Context(), id)
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
	currentUser := authctx.MustCurrentUser(c)
	if currentUser.Role == model.RoleTeacher {
		query.CreatedBy = &currentUser.UserID
	}

	result, err := h.queries.ListChallenges(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *Handler) PreviewChallengeImport(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.InvalidParams(c, "缺少题目包文件")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InvalidParams(c, "无法读取题目包文件")
		return
	}
	defer file.Close()

	resp, err := h.commands.PreviewChallengeImport(
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

func (h *Handler) ListChallengeImports(c *gin.Context) {
	resp, err := h.commands.ListChallengeImports(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetChallengeImport(c *gin.Context) {
	resp, err := h.commands.GetChallengeImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) CommitChallengeImport(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.InvalidParams(c, "无效的导入 ID")
		return
	}

	resp, err := h.commands.CommitChallengeImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, &dto.ChallengeImportCommitResp{Challenge: resp})
}

func (h *Handler) ExportChallengePackage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}
	resp, err := h.commands.ExportChallengePackage(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	resp.DownloadURL = fmt.Sprintf("/api/v1/authoring/challenges/%d/package-export/download?revision_id=%d", id, resp.RevisionID)
	response.SuccessWithStatus(c, nethttp.StatusCreated, resp)
}

func (h *Handler) DownloadChallengePackageExport(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}
	var revisionID *int64
	if raw := strings.TrimSpace(c.Query("revision_id")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil {
			response.InvalidParams(c, "无效的 revision_id")
			return
		}
		revisionID = &parsed
	}
	resp, err := h.commands.GetChallengePackageExport(c.Request.Context(), challengeID, revisionID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	c.FileAttachment(resp.ArchivePath, resp.FileName)
}

func (h *Handler) SelfCheckChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	resp, err := h.commands.SelfCheckChallenge(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) RequestPublishCheck(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	resp, err := h.commands.RequestPublishCheck(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.SuccessWithStatus(c, nethttp.StatusAccepted, resp)
}

func (h *Handler) GetLatestPublishCheck(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	resp, err := h.commands.GetLatestPublishCheck(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

// ListPublishedChallenges 靶场列表（学员视图）
func (h *Handler) ListPublishedChallenges(c *gin.Context) {
	var query dto.ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.queries.ListPublishedChallenges(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &query)
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

	detail, err := h.queries.GetPublishedChallenge(c.Request.Context(), authctx.MustCurrentUser(c).UserID, id)
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

	baseDir := resolveChallengeAttachmentBaseDir(cleanPath)

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

func resolveChallengeAttachmentBaseDir(relativePath string) string {
	if strings.HasPrefix(relativePath, "imports/") {
		baseDir := strings.TrimSpace(os.Getenv("CHALLENGE_ATTACHMENT_STORAGE_DIR"))
		if baseDir == "" {
			baseDir = "./data/challenge-attachments"
		}
		return baseDir
	}

	baseDir := strings.TrimSpace(os.Getenv("CHALLENGE_PACKS_DIR"))
	if baseDir == "" {
		baseDir = "../../docs/challenges/packs"
	}
	return baseDir
}
