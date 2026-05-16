package http

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service        practiceService
	rankingService practiceRankingService
	progressQuery  practiceProgressTimelineQueryService
}

type practiceService interface {
	StartChallenge(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error)
	StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error)
	StartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*dto.InstanceResp, error)
	RestartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*dto.InstanceResp, error)
	GetContestAWDInstanceOrchestration(ctx context.Context, contestID int64) (*dto.AdminAWDInstanceOrchestrationResp, error)
	StartAdminContestAWDTeamService(ctx context.Context, contestID, teamID, serviceID int64) (*dto.AdminAWDInstanceItemResp, error)
	SetAdminContestAWDTeamRetired(ctx context.Context, contestID, teamID, actorUserID int64, retired bool, reason string) (*dto.AdminAWDScopeControlResp, error)
	SetAdminContestAWDTeamServiceDisabled(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, disabled bool, reason string) (*dto.AdminAWDScopeControlResp, error)
	SetAdminContestAWDDesiredReconcileSuppressed(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, suppressed bool, reason string) (*dto.AdminAWDScopeControlResp, error)
	PrewarmAdminContestAWDInstances(ctx context.Context, contestID int64, req *dto.PrewarmAdminContestAWDInstancesReq) (*dto.AdminAWDInstancePrewarmResp, error)
	SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*dto.SubmissionResp, error)
	ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*dto.ChallengeSubmissionRecordResp, error)
	ListTeacherManualReviewSubmissions(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherManualReviewSubmissionQuery) (*dto.PageResult[*dto.TeacherManualReviewSubmissionItemResp], error)
	GetTeacherManualReviewSubmission(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.TeacherManualReviewSubmissionDetailResp, error)
	ReviewManualReviewSubmission(ctx context.Context, submissionID, reviewerID int64, reviewerRole string, req *dto.ReviewManualReviewSubmissionReq) (*dto.TeacherManualReviewSubmissionDetailResp, error)
}

type practiceRankingService interface {
	GetRanking(ctx context.Context, limit int) ([]*dto.RankingItem, error)
}

type practiceProgressTimelineQueryService interface {
	GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error)
	GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error)
}

func NewHandler(service practiceService, rankingService practiceRankingService, progressQuery practiceProgressTimelineQueryService) *Handler {
	return &Handler{service: service, rankingService: rankingService, progressQuery: progressQuery}
}

// StartChallenge 启动靶机实例
// POST /api/v1/challenges/:id/instances
func (h *Handler) StartChallenge(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartChallenge(c.Request.Context(), userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// StartContestChallenge 启动竞赛靶机实例
// POST /api/v1/contests/:id/challenges/:cid/instances
func (h *Handler) StartContestChallenge(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartContestChallenge(c.Request.Context(), userID, contestID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// StartContestAWDService 启动 AWD 服务实例
// POST /api/v1/contests/:id/awd/services/:sid/instances
func (h *Handler) StartContestAWDService(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartContestAWDService(c.Request.Context(), userID, contestID, serviceID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// RestartContestAWDService 重启本队 AWD 服务实例
// POST /api/v1/contests/:id/awd/services/:sid/instances/restart
func (h *Handler) RestartContestAWDService(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.RestartContestAWDService(c.Request.Context(), userID, contestID, serviceID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// GetAdminContestAWDInstanceOrchestration 查看 AWD 队伍服务实例编排
// GET /api/v1/admin/contests/:id/awd/instances
func (h *Handler) GetAdminContestAWDInstanceOrchestration(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, err := h.service.GetContestAWDInstanceOrchestration(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// StartAdminContestAWDInstance 启动指定队伍的 AWD 服务实例
// POST /api/v1/admin/contests/:id/awd/instances
func (h *Handler) StartAdminContestAWDInstance(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.StartAdminContestAWDInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	if req.TeamID <= 0 || req.ServiceID <= 0 {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, err := h.service.StartAdminContestAWDTeamService(
		c.Request.Context(),
		contestID,
		req.TeamID,
		req.ServiceID,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// PrewarmAdminContestAWDInstances 赛前批量预热 AWD 队伍服务实例
// POST /api/v1/admin/contests/:id/awd/instances/prewarm
func (h *Handler) PrewarmAdminContestAWDInstances(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.PrewarmAdminContestAWDInstancesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	if req.TeamID != nil && *req.TeamID <= 0 {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, err := h.service.PrewarmAdminContestAWDInstances(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// SetAdminContestAWDTeamRetired 设置队伍退赛控制
// PUT /api/v1/admin/contests/:id/awd/teams/:team_id/retirement
func (h *Handler) SetAdminContestAWDTeamRetired(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SetAdminContestAWDTeamRetiredReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SetAdminContestAWDTeamRetired(
		c.Request.Context(),
		contestID,
		teamID,
		authctx.MustCurrentUser(c).UserID,
		req.Retired != nil && *req.Retired,
		req.Reason,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

// SetAdminContestAWDTeamServiceDisabled 设置队伍服务停用控制
// PUT /api/v1/admin/contests/:id/awd/teams/:team_id/services/:sid/disabled
func (h *Handler) SetAdminContestAWDTeamServiceDisabled(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SetAdminContestAWDServiceDisabledReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SetAdminContestAWDTeamServiceDisabled(
		c.Request.Context(),
		contestID,
		teamID,
		serviceID,
		authctx.MustCurrentUser(c).UserID,
		req.Disabled != nil && *req.Disabled,
		req.Reason,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

// SetAdminContestAWDDesiredReconcileSuppressed 设置 scope 级 desired reconcile suppress
// PUT /api/v1/admin/contests/:id/awd/teams/:team_id/services/:sid/suppression
func (h *Handler) SetAdminContestAWDDesiredReconcileSuppressed(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SetAdminContestAWDDesiredReconcileSuppressedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SetAdminContestAWDDesiredReconcileSuppressed(
		c.Request.Context(),
		contestID,
		teamID,
		serviceID,
		authctx.MustCurrentUser(c).UserID,
		req.Suppressed != nil && *req.Suppressed,
		req.Reason,
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

// SubmitFlag 提交 Flag
// POST /api/v1/challenges/:id/submit
func (h *Handler) SubmitFlag(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	auditControl := &auditlog.Control{}
	ctx := auditlog.WithControl(c.Request.Context(), auditControl)

	resp, err := h.service.SubmitFlag(ctx, userID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}
	if auditControl.Skip {
		middleware.SetSkipAudit(c)
	}

	response.Success(c, resp)
}

func (h *Handler) ListMyChallengeSubmissions(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, err := h.service.ListMyChallengeSubmissions(c.Request.Context(), userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) GetRanking(c *gin.Context) {
	limit := 100
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			response.Error(c, errcode.ErrInvalidParams)
			return
		}
		limit = parsed
	}

	resp, err := h.rankingService.GetRanking(c.Request.Context(), limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetProgress 获取个人解题进度
// GET /api/v1/users/me/progress
func (h *Handler) GetProgress(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	resp, err := h.progressQuery.GetProgress(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetTimeline 获取解题时间线
// GET /api/v1/users/me/timeline
func (h *Handler) GetTimeline(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	var req struct {
		Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
		Offset int `form:"offset" binding:"omitempty,min=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	resp, err := h.progressQuery.GetTimeline(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) ListTeacherManualReviewSubmissions(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherManualReviewSubmissionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.service.ListTeacherManualReviewSubmissions(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetTeacherManualReviewSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	resp, err := h.service.GetTeacherManualReviewSubmission(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ReviewManualReviewSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	var req dto.ReviewManualReviewSubmissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.service.ReviewManualReviewSubmission(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
