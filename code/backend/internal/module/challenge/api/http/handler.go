package http

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
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
	commands        challengeCommandService
	queries         challengeQueryService
	packageDelivery packageDeliveryService
}

type challengeCommandService interface {
	CreateChallenge(ctx context.Context, actorUserID int64, req challengecmd.CreateChallengeInput) (*challengecontracts.ChallengeResp, error)
	UpdateChallenge(ctx context.Context, id int64, req challengecmd.UpdateChallengeInput) error
	DeleteChallenge(ctx context.Context, id int64) error
	RequestPublishCheck(ctx context.Context, actorUserID, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error)
	GetLatestPublishCheck(ctx context.Context, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error)
	SelfCheckChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeSelfCheckResp, error)
	PreviewChallengeImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error)
	ListChallengeImports(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error)
	GetChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error)
	CommitChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error)
	ExportChallengePackage(ctx context.Context, actorUserID int64, challengeID int64) (*challengecontracts.ChallengePackageExportResp, error)
	GetChallengePackageExport(ctx context.Context, challengeID int64, revisionID *int64) (*challengecontracts.ChallengePackageExportResp, error)
}

type challengeQueryService interface {
	GetChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeResp, error)
	ListChallenges(ctx context.Context, query *challengecontracts.ChallengeQuery) (*challengecontracts.PageResult[*challengecontracts.ChallengeResp], error)
	ListPublishedChallenges(ctx context.Context, userID int64, query *challengecontracts.ChallengeQuery) (*challengecontracts.PageResult[*challengecontracts.ChallengeListItem], error)
	GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*challengecontracts.ChallengeDetailResp, error)
}

type packageDeliveryService interface {
	Preview(ctx context.Context, req challengecmd.PackageDeliveryPreviewRequest) (*challengecmd.PackageDeliveryPreviewResult, error)
	Commit(ctx context.Context, req challengecmd.PackageDeliveryCommitRequest) (*challengecmd.PackageDeliveryCommitResult, error)
}

func NewHandler(commands challengeCommandService, queries challengeQueryService) *Handler {
	return &Handler{
		commands:        commands,
		queries:         queries,
		packageDelivery: challengecmd.NewPackageDeliveryService(commands, nil),
	}
}

func (h *Handler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateChallenge(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeRequestMapper.ToCreateChallengeInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, toChallengeResp(resp))
}

func (h *Handler) UpdateChallenge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的ID")
		return
	}

	var req UpdateChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.commands.UpdateChallenge(c.Request.Context(), id, challengeRequestMapper.ToUpdateChallengeInput(req)); err != nil {
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

	response.Success(c, toChallengeResp(resp))
}

func (h *Handler) ListChallenges(c *gin.Context) {
	var query ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	mappedQuery := challengeRequestMapper.ToChallengeQuery(query)
	result, err := h.queries.ListChallenges(c.Request.Context(), &mappedQuery)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, toChallengePageResult(result))
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

	result, err := h.packageDelivery.Preview(c.Request.Context(), challengecmd.PackageDeliveryPreviewRequest{
		Mode:        challengecmd.PackageDeliveryModeJeopardy,
		ActorUserID: authctx.MustCurrentUser(c).UserID,
		FileName:    fileHeader.Filename,
		Reader:      file,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.SuccessWithStatus(c, nethttp.StatusCreated, toChallengeImportPreviewResp(result.Jeopardy))
}

func (h *Handler) ListChallengeImports(c *gin.Context) {
	resp, err := h.commands.ListChallengeImports(c.Request.Context(), authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toChallengeImportPreviewRespList(resp))
}

func (h *Handler) GetChallengeImport(c *gin.Context) {
	resp, err := h.commands.GetChallengeImport(c.Request.Context(), authctx.MustCurrentUser(c).UserID, strings.TrimSpace(c.Param("id")))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toChallengeImportPreviewResp(resp))
}

func (h *Handler) CommitChallengeImport(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.InvalidParams(c, "无效的导入 ID")
		return
	}

	result, err := h.packageDelivery.Commit(c.Request.Context(), challengecmd.PackageDeliveryCommitRequest{
		Mode:        challengecmd.PackageDeliveryModeJeopardy,
		ActorUserID: authctx.MustCurrentUser(c).UserID,
		PreviewID:   id,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, toChallengeImportCommitResp(result.Jeopardy))
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
	response.SuccessWithStatus(c, nethttp.StatusCreated, toChallengePackageExportResp(resp))
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

	response.Success(c, toChallengeSelfCheckResp(resp))
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
	response.SuccessWithStatus(c, nethttp.StatusAccepted, toChallengePublishCheckJobResp(resp))
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
	response.Success(c, toChallengePublishCheckJobResp(resp))
}

// ListPublishedChallenges 靶场列表（学员视图）
func (h *Handler) ListPublishedChallenges(c *gin.Context) {
	var query ChallengeQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	mappedQuery := challengeRequestMapper.ToChallengeQuery(query)
	result, err := h.queries.ListPublishedChallenges(c.Request.Context(), authctx.MustCurrentUser(c).UserID, &mappedQuery)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, toChallengeListItemPageResult(result))
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

	response.Success(c, toChallengeDetailResp(detail))
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
			response.Error(c, apperror.ErrNotFound)
			return
		}
		response.FromError(c, err)
		return
	}
	if info.IsDir() {
		response.Error(c, apperror.ErrNotFound)
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
