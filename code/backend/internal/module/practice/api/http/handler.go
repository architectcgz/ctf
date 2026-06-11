package http

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type Handler struct {
	instances      practiceInstanceLifecycleService
	submissions    practiceSubmissionService
	manualReviews  practiceManualReviewService
	rankingService practiceRankingService
	progressQuery  practiceProgressTimelineQueryService
}

type practiceInstanceLifecycleService interface {
	StartChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.InstanceResp, error)
	StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*instancecontracts.InstanceResp, error)
	StartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*instancecontracts.InstanceResp, error)
	RestartContestAWDService(ctx context.Context, userID, contestID, serviceID int64) (*instancecontracts.InstanceResp, error)
	GetContestAWDInstanceOrchestration(ctx context.Context, contestID int64) (*practicecontracts.AdminAWDInstanceOrchestrationResp, error)
	StartAdminContestAWDTeamService(ctx context.Context, contestID, teamID, serviceID int64) (*practicecontracts.AdminAWDInstanceItemResp, error)
	SetAdminContestAWDTeamRetired(ctx context.Context, contestID, teamID, actorUserID int64, retired bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error)
	SetAdminContestAWDTeamServiceDisabled(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, disabled bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error)
	SetAdminContestAWDDesiredReconcileSuppressed(ctx context.Context, contestID, teamID, serviceID, actorUserID int64, suppressed bool, reason string) (*practicecontracts.AdminAWDScopeControlResp, error)
	PrewarmAdminContestAWDInstances(ctx context.Context, contestID int64, teamID *int64) (*practicecontracts.AdminAWDInstancePrewarmResp, error)
}

type practiceSubmissionService interface {
	SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*practicecontracts.SubmissionResp, error)
	ListMyChallengeSubmissions(ctx context.Context, userID, challengeID int64) ([]*practicecontracts.ChallengeSubmissionRecordResp, error)
}

type practiceManualReviewService interface {
	ListTeacherManualReviewSubmissions(ctx context.Context, requesterID int64, requesterRole string, query *practicecontracts.TeacherManualReviewSubmissionQuery) (*practicecontracts.PageResult[*practicecontracts.TeacherManualReviewSubmissionItemResp], error)
	GetTeacherManualReviewSubmission(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error)
	ReviewManualReviewSubmission(ctx context.Context, submissionID, reviewerID int64, reviewerRole string, req *practicecontracts.ReviewManualReviewSubmissionReq) (*practicecontracts.TeacherManualReviewSubmissionDetailResp, error)
}

type practiceRankingService interface {
	GetRanking(ctx context.Context, limit int) ([]*practicecontracts.RankingItem, error)
}

type practiceProgressTimelineQueryService interface {
	GetProgress(ctx context.Context, userID int64) (*practiceports.UserProgressSnapshot, error)
	GetTimeline(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error)
}

func NewHandler(
	instances practiceInstanceLifecycleService,
	submissions practiceSubmissionService,
	manualReviews practiceManualReviewService,
	rankingService practiceRankingService,
	progressQuery practiceProgressTimelineQueryService,
) *Handler {
	return &Handler{
		instances:      instances,
		submissions:    submissions,
		manualReviews:  manualReviews,
		rankingService: rankingService,
		progressQuery:  progressQuery,
	}
}

// StartChallenge 启动靶机实例
// POST /api/v1/challenges/:id/instances
func (h *Handler) StartChallenge(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	instance, err := h.instances.StartChallenge(c.Request.Context(), userID, challengeID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	instance, err := h.instances.StartContestChallenge(c.Request.Context(), userID, contestID, challengeID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	instance, err := h.instances.StartContestAWDService(c.Request.Context(), userID, contestID, serviceID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	instance, err := h.instances.RestartContestAWDService(c.Request.Context(), userID, contestID, serviceID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	resp, err := h.instances.GetContestAWDInstanceOrchestration(c.Request.Context(), contestID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req StartAdminContestAWDInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	if req.TeamID <= 0 || req.ServiceID <= 0 {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	resp, err := h.instances.StartAdminContestAWDTeamService(
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req PrewarmAdminContestAWDInstancesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	if req.TeamID != nil && *req.TeamID <= 0 {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	resp, err := h.instances.PrewarmAdminContestAWDInstances(c.Request.Context(), contestID, req.TeamID)
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req SetAdminContestAWDTeamRetiredReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.instances.SetAdminContestAWDTeamRetired(
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req SetAdminContestAWDServiceDisabledReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.instances.SetAdminContestAWDTeamServiceDisabled(
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
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}
	serviceID, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req SetAdminContestAWDDesiredReconcileSuppressedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.instances.SetAdminContestAWDDesiredReconcileSuppressed(
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

func (h *Handler) GetRanking(c *gin.Context) {
	limit := 100
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			response.Error(c, apperror.ErrInvalidParams)
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

	snapshot, err := h.progressQuery.GetProgress(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, practiceResponseMapper.ToProgressRespPtr(snapshot))
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

	timeline, err := h.progressQuery.GetTimeline(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, practiceResponseMapper.ToTimelineRespPtr(timeline))
}

func (h *Handler) ListTeacherManualReviewSubmissions(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query practicecontracts.TeacherManualReviewSubmissionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.manualReviews.ListTeacherManualReviewSubmissions(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
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
	resp, err := h.manualReviews.GetTeacherManualReviewSubmission(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
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
	var req practicecontracts.ReviewManualReviewSubmissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.manualReviews.ReviewManualReviewSubmission(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
