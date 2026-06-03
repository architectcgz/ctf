package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	"ctf-platform/internal/middleware"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type adminRouteDeps struct {
	identityHandler *identityhttp.Handler
	auditRecorder   auditlog.Recorder
	auditLogger     *zap.Logger
	assessment      *composition.AssessmentModule
	challenge       *composition.ChallengeModule
	contest         *composition.ContestModule
	ops             *composition.OpsModule
	practice        *composition.PracticeModule
	tokenService    authcontracts.TokenService
}

type userRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	contest       *composition.ContestModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
	teachingQuery *composition.TeachingQueryModule
}

type challengeLookup interface {
	FindByID(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error)
}

func routeAudit(recorder auditlog.Recorder, logger *zap.Logger, options middleware.AuditOptions) gin.HandlerFunc {
	return middleware.Audit(recorder, options, logger)
}

func challengeOwnerGuard(catalog challengeLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := authctx.MustCurrentUser(c)
		if currentUser.Role == identitycontracts.RoleAdmin {
			c.Next()
			return
		}

		challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.InvalidParams(c, "无效的ID")
			c.Abort()
			return
		}

		challenge, err := catalog.FindByID(c.Request.Context(), challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, challengecontracts.ErrChallengeNotFound)
			} else {
				response.FromError(c, err)
			}
			c.Abort()
			return
		}
		if challenge.CreatedBy == nil || *challenge.CreatedBy != currentUser.UserID {
			response.Error(c, apperror.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func registerTeacherAuthoringRoutes(adminAuthoring *gin.RouterGroup, deps adminRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}
	ownerGuard := challengeOwnerGuard(deps.challenge.Catalog)

	adminAuthoring.POST("/challenge-imports",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "challenge_import",
		}),
		deps.challenge.Handler.PreviewChallengeImport,
	)
	adminAuthoring.GET("/challenge-imports", deps.challenge.Handler.ListChallengeImports)
	adminAuthoring.GET("/challenge-imports/:id", deps.challenge.Handler.GetChallengeImport)
	adminAuthoring.POST("/challenge-imports/:id/commit",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "challenge_import_commit",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.CommitChallengeImport,
	)

	adminAuthoring.POST("/images",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "image",
		}),
		deps.challenge.ImageHandler.CreateImage,
	)
	adminAuthoring.GET("/images", deps.challenge.ImageHandler.ListImages)
	adminAuthoring.GET("/images/:id", deps.challenge.ImageHandler.GetImage)
	adminAuthoring.PUT("/images/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.UpdateImage,
	)
	adminAuthoring.DELETE("/images/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.DeleteImage,
	)

	adminAuthoring.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "challenge",
		}),
		deps.challenge.Handler.CreateChallenge,
	)
	adminAuthoring.GET("/challenges", deps.challenge.Handler.ListChallenges)
	adminAuthoring.GET("/challenges/:id", deps.challenge.Handler.GetChallenge)
	adminAuthoring.PUT("/challenges/:id",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.UpdateChallenge,
	)
	adminAuthoring.DELETE("/challenges/:id",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.DeleteChallenge,
	)
	adminAuthoring.POST("/challenges/:id/publish-requests",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "challenge_publish_request",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.RequestPublishCheck,
	)
	adminAuthoring.GET("/challenges/:id/publish-requests/latest", deps.challenge.Handler.GetLatestPublishCheck)
	adminAuthoring.POST("/challenges/:id/self-check",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "challenge_self_check",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.SelfCheckChallenge,
	)
	adminAuthoring.GET("/challenges/:id/writeup", deps.challenge.WriteupHandler.GetAdmin)
	adminAuthoring.PUT("/challenges/:id/writeup",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Upsert,
	)
	adminAuthoring.POST("/challenges/:id/writeup/recommend",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RecommendOfficial,
	)
	adminAuthoring.DELETE("/challenges/:id/writeup/recommend",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UnrecommendOfficial,
	)
	adminAuthoring.DELETE("/challenges/:id/writeup",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Delete,
	)
	adminAuthoring.GET("/challenges/:id/topology", deps.challenge.TopologyHandler.GetChallengeTopology)
	adminAuthoring.PUT("/challenges/:id/topology",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.SaveChallengeTopology,
	)
	adminAuthoring.POST("/challenges/:id/package-export",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "challenge_package_export",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.ExportChallengePackage,
	)
	adminAuthoring.GET("/challenges/:id/package-export/download", deps.challenge.Handler.DownloadChallengePackageExport)
	adminAuthoring.DELETE("/challenges/:id/topology",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteChallengeTopology,
	)
	adminAuthoring.GET("/environment-templates", deps.challenge.TopologyHandler.ListTemplates)
	adminAuthoring.POST("/environment-templates",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "environment_template",
		}),
		deps.challenge.TopologyHandler.CreateTemplate,
	)
	adminAuthoring.GET("/environment-templates/:id", deps.challenge.TopologyHandler.GetTemplate)
	adminAuthoring.PUT("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.UpdateTemplate,
	)
	adminAuthoring.DELETE("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteTemplate,
	)

	adminAuthoring.GET("/awd-challenges", deps.challenge.AWDChallengeHandler.ListChallenges)
	adminAuthoring.POST("/awd-challenge-imports",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "awd_challenge_import",
		}),
		deps.challenge.AWDChallengeHandler.PreviewImport,
	)
	adminAuthoring.GET("/awd-challenge-imports", deps.challenge.AWDChallengeHandler.ListImports)
	adminAuthoring.GET("/awd-challenge-imports/:id", deps.challenge.AWDChallengeHandler.GetImport)
	adminAuthoring.POST("/awd-challenge-imports/:id/commit",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "awd_challenge_import_commit",
			ResourceIDParam: "id",
		}),
		deps.challenge.AWDChallengeHandler.CommitImport,
	)
	adminAuthoring.POST("/awd-challenges",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "awd_challenge",
		}),
		deps.challenge.AWDChallengeHandler.CreateChallenge,
	)
	adminAuthoring.GET("/awd-challenges/:id", deps.challenge.AWDChallengeHandler.GetChallenge)
	adminAuthoring.PUT("/awd-challenges/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "awd_challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.AWDChallengeHandler.UpdateChallenge,
	)
	adminAuthoring.DELETE("/awd-challenges/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "awd_challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.AWDChallengeHandler.DeleteChallenge,
	)

	adminAuthoring.PUT("/challenges/:id/flag",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "challenge_flag",
			ResourceIDParam: "id",
		}),
		deps.challenge.FlagHandler.ConfigureFlag,
	)
	adminAuthoring.GET("/challenges/:id/flag", deps.challenge.FlagHandler.GetFlagConfig)
}

func registerAdminRoutes(adminOnly *gin.RouterGroup, deps adminRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	adminOnly.GET("/audit-logs", deps.ops.AuditHandler.ListAuditLogs)
	adminOnly.GET("/dashboard", deps.ops.DashboardHandler.GetDashboard)
	adminOnly.GET("/cheat-detection", deps.ops.RiskHandler.GetCheatDetection)
	adminOnly.POST("/notifications",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionAdminOp,
			ResourceType: "notification_batch",
		}),
		deps.ops.NotificationHandler.PublishAdminNotification,
	)
	adminOnly.GET("/users", deps.identityHandler.ListUsers)
	adminOnly.POST("/users",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "user",
		}),
		deps.identityHandler.CreateUser,
	)
	adminOnly.PUT("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.UpdateUser,
	)
	adminOnly.DELETE("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.DeleteUser,
	)
	adminOnly.POST("/users/import",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "user_import",
		}),
		deps.identityHandler.ImportUsers,
	)

	// Session management
	adminOnly.GET("/users/:id/sessions",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionRead,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
		}),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			sessions, err := deps.tokenService.ListUserSessions(c.Request.Context(), userID)
			if err != nil {
				response.FromError(c, err)
				return
			}
			resps := make([]identityhttp.UserSessionResp, 0, len(sessions))
			for _, s := range sessions {
				resps = append(resps, identityhttp.UserSessionResp{
					ID:        s.ID,
					Username:  s.Username,
					Role:      s.Role,
					ExpiresAt: s.ExpiresAt,
				})
			}
			response.Success(c, gin.H{"sessions": resps})
		},
	)
	adminOnly.DELETE("/users/:id/sessions/:sid",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		func(c *gin.Context) {
			sessionID := c.Param("sid")
			if sessionID == "" {
				response.InvalidParams(c, "缺少会话ID")
				return
			}
			// 验证会话归属：先获取会话，确认属于该用户
			session, err := deps.tokenService.GetSession(c.Request.Context(), sessionID)
			if err != nil {
				if errors.Is(err, authcontracts.ErrAccessTokenExpired) {
					response.Error(c, apperror.ErrNotFound.WithMessage("该会话已不活跃或已被撤销"))
					return
				}
				response.FromError(c, err)
				return
			}
			userID := c.GetInt64("id")
			if session.UserID != userID {
				response.Error(c, apperror.ErrForbidden)
				return
			}
			if err := deps.tokenService.DeleteSession(c.Request.Context(), sessionID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "会话已撤销"})
		},
	)
	adminOnly.DELETE("/users/:id/sessions",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "user_session",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		func(c *gin.Context) {
			userID := c.GetInt64("id")
			if err := deps.tokenService.RevokeAllUserSessions(c.Request.Context(), userID); err != nil {
				response.FromError(c, err)
				return
			}
			response.Success(c, gin.H{"message": "已撤销该用户所有会话"})
		},
	)

	registerAdminContestRoutes(adminOnly, adminContestRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		assessment:    deps.assessment,
		contest:       deps.contest,
		practice:      deps.practice,
	})
}

func registerUserRoutes(apiV1, protected, teacherOrAbove *gin.RouterGroup, deps userRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	contestGroup := apiV1.Group("/contests")
	contestGroup.GET("", deps.contest.Handler.ListContests)
	contestGroup.GET("/:id", middleware.ParseInt64Param("id"), deps.contest.Handler.GetContest)
	contestGroup.GET("/:id/scoreboard", deps.contest.Handler.GetScoreboard)
	contestGroup.GET("/:id/announcements", deps.contest.ParticipationHandler.ListAnnouncements)

	protected.POST("/contests/:id/register",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_registration",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ParticipationHandler.RegisterContest,
	)
	protected.GET("/contests/:id/challenges", deps.contest.ChallengeHandler.ListChallenges)
	protected.GET("/contests/:id/my-progress", deps.contest.ParticipationHandler.GetMyProgress)
	protected.GET("/contests/:id/awd/workspace",
		middleware.ParseInt64Param("id"),
		deps.contest.AWDHandler.GetUserWorkspace,
	)
	protected.POST("/contests/:id/challenges/:cid/submissions",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionSubmit,
			ResourceType:    "contest_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.SubmissionHandler.SubmitFlag,
	)
	protected.POST("/contests/:id/awd/services/:sid/submissions",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionSubmit,
			ResourceType:    "awd_attack_submission",
			ResourceIDParam: "sid",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		deps.contest.AWDHandler.SubmitAttack,
	)
	protected.POST("/contests/:id/awd/services/:sid/targets/:team_id/access",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		middleware.ParseInt64Param("team_id"),
		deps.instance.Handler.AccessAWDTarget,
	)
	protected.POST("/contests/:id/awd/services/:sid/defense/ssh",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		deps.instance.Handler.AccessAWDDefenseSSH,
	)
	apiV1.GET("/contests/:id/awd/services/:sid/targets/:team_id/proxy",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		middleware.ParseInt64Param("team_id"),
		deps.instance.Handler.ProxyAWDTarget,
	)
	apiV1.Any("/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		middleware.ParseInt64Param("team_id"),
		deps.instance.Handler.ProxyAWDTarget,
	)
	protected.GET("/contests/:id/teams", deps.contest.TeamHandler.ListTeams)
	protected.GET("/contests/:id/my-team", deps.contest.TeamHandler.GetMyTeam)
	protected.POST("/contests/:id/teams",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "team",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.TeamHandler.CreateTeam,
	)
	protected.POST("/contests/:id/teams/:tid/join",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.JoinTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/leave",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.LeaveTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "team",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.DismissTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/members/:uid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "uid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid", "uid"),
		}),
		deps.contest.TeamHandler.KickMember,
	)

	protected.GET("/challenges", deps.challenge.Handler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionRead,
			ResourceType:    "challenge_detail",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.GetPublishedChallenge,
	)
	protected.GET("/challenges/attachments/*path", deps.challenge.Handler.DownloadAttachment)
	protected.GET("/challenges/:id/writeup", deps.challenge.WriteupHandler.GetPublished)
	protected.GET("/challenges/:id/solutions/recommended", deps.challenge.WriteupHandler.ListRecommendedSolutions)
	protected.GET("/challenges/:id/solutions/community", deps.challenge.WriteupHandler.ListCommunitySolutions)
	protected.POST("/challenges/:id/writeup-submissions",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "submission_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UpsertSubmission,
	)
	protected.GET("/challenges/:id/writeup-submissions/me", deps.challenge.WriteupHandler.GetMySubmission)
	protected.POST("/challenges/:id/instances",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.practice.Handler.StartChallenge,
	)
	protected.POST("/contests/:id/challenges/:cid/instances",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("cid"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "contest_instance",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.practice.Handler.StartContestChallenge,
	)
	protected.POST("/contests/:id/awd/services/:sid/instances",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionCreate,
			ResourceType:    "contest_awd_instance",
			ResourceIDParam: "sid",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		deps.practice.Handler.StartContestAWDService,
	)
	protected.POST("/contests/:id/awd/services/:sid/instances/restart",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("sid"),
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest_awd_instance",
			ResourceIDParam: "sid",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		deps.practice.Handler.RestartContestAWDService,
	)
	protected.POST("/challenges/:id/submit",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}),
		deps.practice.Handler.SubmitFlag,
	)
	protected.GET("/challenges/:id/submissions/mine", deps.practice.Handler.ListMyChallengeSubmissions)
	protected.GET("/scoreboard/ranking", deps.practice.Handler.GetRanking)
	protected.GET("/instances", deps.instance.Handler.ListInstances)
	protected.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.instance.Handler.DestroyInstance,
	)
	protected.POST("/instances/:id/extend",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.instance.Handler.ExtendInstance,
	)
	protected.POST("/instances/:id/access", deps.instance.Handler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", deps.instance.Handler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", deps.instance.Handler.ProxyInstance)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", deps.practice.Handler.GetProgress)
	usersGroup.GET("/me/timeline", deps.practice.Handler.GetTimeline)
	usersGroup.GET("/me/skill-profile", deps.assessment.Handler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", deps.assessment.Handler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(identitycontracts.RoleTeacher), deps.assessment.Handler.GetStudentSkillProfile)

	teacherOrAbove.GET("/overview", deps.teachingQuery.Handler.GetOverview)
	teacherOrAbove.GET("/classes", deps.teachingQuery.Handler.ListClasses)
	teacherOrAbove.GET("/students", deps.teachingQuery.Handler.ListStudents)
	teacherOrAbove.GET("/classes/:name/students", deps.teachingQuery.Handler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", deps.teachingQuery.Handler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", deps.teachingQuery.Handler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", deps.teachingQuery.Handler.GetClassReview)
	teacherOrAbove.GET("/instances", deps.instance.Handler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.instance.Handler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", deps.teachingQuery.Handler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", deps.assessment.Handler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", deps.teachingQuery.Handler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", deps.teachingQuery.Handler.GetStudentTimeline)
	teacherOrAbove.GET("/students/:id/evidence", deps.teachingQuery.Handler.GetStudentEvidence)
	teacherOrAbove.GET("/students/:id/attack-sessions", deps.teachingQuery.Handler.GetStudentAttackSessions)
	teacherOrAbove.GET("/students/:id/review-archive", deps.assessment.ReportHandler.GetStudentReviewArchive)
	teacherOrAbove.POST("/students/:id/review-archive/export",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "review_archive_export",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.assessment.ReportHandler.CreateStudentReviewArchive,
	)
	teacherOrAbove.GET("/awd/reviews", deps.assessment.TeacherAWDReviewHandler.ListReviews)
	teacherOrAbove.GET("/awd/reviews/:id",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.GetReview,
	)
	teacherOrAbove.POST("/awd/reviews/:id/export/archive",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.ExportArchive,
	)
	teacherOrAbove.POST("/awd/reviews/:id/export/report",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.ExportReport,
	)
	teacherOrAbove.GET("/manual-review-submissions", deps.practice.Handler.ListTeacherManualReviewSubmissions)
	teacherOrAbove.GET("/manual-review-submissions/:id", deps.practice.Handler.GetTeacherManualReviewSubmission)
	teacherOrAbove.PUT("/manual-review-submissions/:id/review",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "manual_review_submission",
			ResourceIDParam: "id",
		}),
		deps.practice.Handler.ReviewManualReviewSubmission,
	)
	teacherOrAbove.POST("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RecommendCommunity,
	)
	teacherOrAbove.DELETE("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UnrecommendCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/hide",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_visibility",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.HideCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/restore",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_visibility",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RestoreCommunity,
	)
	teacherOrAbove.GET("/writeup-submissions", deps.challenge.WriteupHandler.ListTeacherSubmissions)
	teacherOrAbove.GET("/writeup-submissions/:id", deps.challenge.WriteupHandler.GetTeacherSubmission)

	protected.POST("/reports/personal", deps.assessment.ReportHandler.CreatePersonalReport)
	protected.GET("/reports/:id", deps.assessment.ReportHandler.GetReportStatus)
	protected.GET("/reports/:id/download", deps.assessment.ReportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(identitycontracts.RoleTeacher), deps.assessment.ReportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", deps.assessment.ReportHandler.CreateClassReport)
}
