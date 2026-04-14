package app

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type adminRouteDeps struct {
	identityHandler *identityhttp.Handler
	auditRecorder   auditlog.Recorder
	auditLogger     *zap.Logger
	assessment      *composition.AssessmentModule
	challenge       *composition.ChallengeModule
	contest         *composition.ContestModule
	ops             *composition.OpsModule
}

type userRouteDeps struct {
	auditRecorder     auditlog.Recorder
	auditLogger       *zap.Logger
	assessment        *composition.AssessmentModule
	challenge         *composition.ChallengeModule
	contest           *composition.ContestModule
	practice          *composition.PracticeModule
	practiceReadmodel *composition.PracticeReadmodelModule
	runtime           *composition.RuntimeModule
	teachingReadmodel *composition.TeachingReadmodelModule
}

type challengeLookup interface {
	FindByID(id int64) (*model.Challenge, error)
}

func routeAudit(recorder auditlog.Recorder, logger *zap.Logger, options middleware.AuditOptions) gin.HandlerFunc {
	return middleware.Audit(recorder, options, logger)
}

func challengeOwnerGuard(catalog challengeLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := authctx.MustCurrentUser(c)
		if currentUser.Role == model.RoleAdmin {
			c.Next()
			return
		}

		challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.InvalidParams(c, "无效的ID")
			c.Abort()
			return
		}

		challenge, err := catalog.FindByID(challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, errcode.ErrChallengeNotFound)
			} else {
				response.FromError(c, err)
			}
			c.Abort()
			return
		}
		if challenge.CreatedBy == nil || *challenge.CreatedBy != currentUser.UserID {
			response.Error(c, errcode.ErrForbidden)
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
			Action:       model.AuditActionCreate,
			ResourceType: "challenge_import",
		}),
		deps.challenge.Handler.PreviewChallengeImport,
	)
	adminAuthoring.GET("/challenge-imports", deps.challenge.Handler.ListChallengeImports)
	adminAuthoring.GET("/challenge-imports/:id", deps.challenge.Handler.GetChallengeImport)
	adminAuthoring.POST("/challenge-imports/:id/commit",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "challenge_import_commit",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.CommitChallengeImport,
	)

	adminAuthoring.POST("/images",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "image",
		}),
		deps.challenge.ImageHandler.CreateImage,
	)
	adminAuthoring.GET("/images", deps.challenge.ImageHandler.ListImages)
	adminAuthoring.GET("/images/:id", deps.challenge.ImageHandler.GetImage)
	adminAuthoring.PUT("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.UpdateImage,
	)
	adminAuthoring.DELETE("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.DeleteImage,
	)

	adminAuthoring.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "challenge",
		}),
		deps.challenge.Handler.CreateChallenge,
	)
	adminAuthoring.GET("/challenges", deps.challenge.Handler.ListChallenges)
	adminAuthoring.GET("/challenges/:id", ownerGuard, deps.challenge.Handler.GetChallenge)
	adminAuthoring.PUT("/challenges/:id",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.UpdateChallenge,
	)
	adminAuthoring.DELETE("/challenges/:id",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.DeleteChallenge,
	)
	adminAuthoring.POST("/challenges/:id/publish-requests",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "challenge_publish_request",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.RequestPublishCheck,
	)
	adminAuthoring.GET("/challenges/:id/publish-requests/latest",
		ownerGuard,
		deps.challenge.Handler.GetLatestPublishCheck,
	)
	adminAuthoring.POST("/challenges/:id/self-check",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "challenge_self_check",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.SelfCheckChallenge,
	)
	adminAuthoring.GET("/challenges/:id/writeup", ownerGuard, deps.challenge.WriteupHandler.GetAdmin)
	adminAuthoring.PUT("/challenges/:id/writeup",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Upsert,
	)
	adminAuthoring.POST("/challenges/:id/writeup/recommend",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RecommendOfficial,
	)
	adminAuthoring.DELETE("/challenges/:id/writeup/recommend",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UnrecommendOfficial,
	)
	adminAuthoring.DELETE("/challenges/:id/writeup",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Delete,
	)
	adminAuthoring.GET("/challenges/:id/topology", ownerGuard, deps.challenge.TopologyHandler.GetChallengeTopology)
	adminAuthoring.PUT("/challenges/:id/topology",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.SaveChallengeTopology,
	)
	adminAuthoring.DELETE("/challenges/:id/topology",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteChallengeTopology,
	)
	adminAuthoring.GET("/environment-templates", deps.challenge.TopologyHandler.ListTemplates)
	adminAuthoring.POST("/environment-templates",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "environment_template",
		}),
		deps.challenge.TopologyHandler.CreateTemplate,
	)
	adminAuthoring.GET("/environment-templates/:id", deps.challenge.TopologyHandler.GetTemplate)
	adminAuthoring.PUT("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.UpdateTemplate,
	)
	adminAuthoring.DELETE("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteTemplate,
	)

	adminAuthoring.PUT("/challenges/:id/flag",
		ownerGuard,
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_flag",
			ResourceIDParam: "id",
		}),
		deps.challenge.FlagHandler.ConfigureFlag,
	)
	adminAuthoring.GET("/challenges/:id/flag", ownerGuard, deps.challenge.FlagHandler.GetFlagConfig)
}

func registerAdminRoutes(adminOnly *gin.RouterGroup, deps adminRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}
	awdReadinessAudit := func() gin.HandlerFunc {
		return middleware.AWDReadinessAudit(deps.auditRecorder, deps.auditLogger)
	}

	adminOnly.GET("/audit-logs", deps.ops.AuditHandler.ListAuditLogs)
	adminOnly.GET("/dashboard", deps.ops.DashboardHandler.GetDashboard)
	adminOnly.GET("/cheat-detection", deps.ops.RiskHandler.GetCheatDetection)
	adminOnly.POST("/notifications",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionAdminOp,
			ResourceType: "notification_batch",
		}),
		deps.ops.NotificationHandler.PublishAdminNotification,
	)
	adminOnly.GET("/users", deps.identityHandler.ListUsers)
	adminOnly.POST("/users",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user",
		}),
		deps.identityHandler.CreateUser,
	)
	adminOnly.PUT("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.UpdateUser,
	)
	adminOnly.DELETE("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.identityHandler.DeleteUser,
	)
	adminOnly.POST("/users/import",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user_import",
		}),
		deps.identityHandler.ImportUsers,
	)

	adminOnly.POST("/contests",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "contest",
		}),
		deps.contest.Handler.CreateContest,
	)
	adminOnly.GET("/contests/:id",
		middleware.ParseInt64Param("id"),
		deps.contest.Handler.GetContest,
	)
	adminOnly.PUT("/contests/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest",
			ResourceIDParam: "id",
		}),
		awdReadinessAudit(),
		deps.contest.Handler.UpdateContest,
	)
	adminOnly.GET("/contests", deps.contest.Handler.ListContests)
	adminOnly.POST("/contests/:id/freeze",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contest.Handler.FreezeScoreboard,
	)
	adminOnly.POST("/contests/:id/unfreeze",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contest.Handler.UnfreezeScoreboard,
	)
	adminOnly.POST("/contests/:id/export",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest_export",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.assessment.ReportHandler.CreateContestExport,
	)
	adminOnly.GET("/contests/:id/challenges", deps.contest.ChallengeHandler.ListAdminChallenges)
	adminOnly.POST("/contests/:id/challenges",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_challenge",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ChallengeHandler.AddChallenge,
	)
	adminOnly.PUT("/contests/:id/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.UpdatePoints,
	)
	adminOnly.DELETE("/contests/:id/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.RemoveChallenge,
	)
	adminOnly.GET("/contests/:id/registrations", deps.contest.ParticipationHandler.ListRegistrations)
	adminOnly.PUT("/contests/:id/registrations/:rid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_registration",
			ResourceIDParam: "rid",
			DetailBuilder:   middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.ParticipationHandler.ReviewRegistration,
	)
	adminOnly.GET("/contests/:id/announcements", deps.contest.ParticipationHandler.ListAnnouncements)
	adminOnly.POST("/contests/:id/announcements",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_announcement",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ParticipationHandler.CreateAnnouncement,
	)
	adminOnly.DELETE("/contests/:id/announcements/:aid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_announcement",
			ResourceIDParam: "aid",
			DetailBuilder:   middleware.DetailFromParams("id", "aid"),
		}),
		deps.contest.ParticipationHandler.DeleteAnnouncement,
	)
	adminOnly.GET("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		deps.contest.AWDHandler.ListRounds,
	)
	adminOnly.GET("/contests/:id/awd/readiness",
		middleware.ParseInt64Param("id"),
		deps.contest.AWDHandler.GetReadiness,
	)
	adminOnly.GET("/contests/:id/scoreboard/live", deps.contest.Handler.GetLiveScoreboard)
	adminOnly.POST("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_round",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		awdReadinessAudit(),
		deps.contest.AWDHandler.CreateRound,
	)
	adminOnly.POST("/contests/:id/awd/current-round/check",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		awdReadinessAudit(),
		deps.contest.AWDHandler.RunCurrentRoundChecks,
	)
	adminOnly.POST("/contests/:id/awd/checker-preview",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_preview",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.AWDHandler.PreviewChecker,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.RunRoundChecks,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/services",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.contest.AWDHandler.ListServices,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/services/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_service_check",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.UpsertServiceCheck,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.contest.AWDHandler.ListAttackLogs,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/traffic/summary",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.contest.AWDHandler.GetTrafficSummary,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/traffic/events",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.contest.AWDHandler.ListTrafficEvents,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_attack_log",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.CreateAttackLog,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/summary",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.contest.AWDHandler.GetRoundSummary,
	)
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
			Action:        model.AuditActionCreate,
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
			Action:          model.AuditActionSubmit,
			ResourceType:    "contest_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.SubmissionHandler.SubmitFlag,
	)
	protected.POST("/contests/:id/awd/challenges/:cid/submissions",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("cid"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "awd_attack_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.AWDHandler.SubmitAttack,
	)
	protected.GET("/contests/:id/teams", deps.contest.TeamHandler.ListTeams)
	protected.GET("/contests/:id/my-team", deps.contest.TeamHandler.GetMyTeam)
	protected.POST("/contests/:id/teams",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "team",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.TeamHandler.CreateTeam,
	)
	protected.POST("/contests/:id/teams/:tid/join",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.JoinTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/leave",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.LeaveTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.contest.TeamHandler.DismissTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/members/:uid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "uid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid", "uid"),
		}),
		deps.contest.TeamHandler.KickMember,
	)

	protected.GET("/challenges", deps.challenge.Handler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionRead,
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
			Action:          model.AuditActionCreate,
			ResourceType:    "submission_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UpsertSubmission,
	)
	protected.GET("/challenges/:id/writeup-submissions/me", deps.challenge.WriteupHandler.GetMySubmission)
	protected.POST("/challenges/:id/instances",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.practice.Handler.StartChallenge,
	)
	protected.POST("/contests/:id/challenges/:cid/instances",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("cid"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "contest_instance",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.practice.Handler.StartContestChallenge,
	)
	protected.POST("/challenges/:id/submit",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}),
		deps.practice.Handler.SubmitFlag,
	)
	protected.GET("/challenges/:id/submissions/mine", deps.practice.Handler.ListMyChallengeSubmissions)
	protected.GET("/scoreboard/ranking", deps.practice.Handler.GetRanking)
	protected.GET("/instances", deps.runtime.Handler.ListInstances)
	protected.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.runtime.Handler.DestroyInstance,
	)
	protected.POST("/instances/:id/extend",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.runtime.Handler.ExtendInstance,
	)
	protected.POST("/instances/:id/access", deps.runtime.Handler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", deps.runtime.Handler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", deps.runtime.Handler.ProxyInstance)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", deps.practiceReadmodel.Handler.GetProgress)
	usersGroup.GET("/me/timeline", deps.practiceReadmodel.Handler.GetTimeline)
	usersGroup.GET("/me/skill-profile", deps.assessment.Handler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", deps.assessment.Handler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(model.RoleTeacher), deps.assessment.Handler.GetStudentSkillProfile)

	teacherOrAbove.GET("/classes", deps.teachingReadmodel.Handler.ListClasses)
	teacherOrAbove.GET("/classes/:name/students", deps.teachingReadmodel.Handler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", deps.teachingReadmodel.Handler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", deps.teachingReadmodel.Handler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", deps.teachingReadmodel.Handler.GetClassReview)
	teacherOrAbove.GET("/instances", deps.runtime.Handler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.runtime.Handler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", deps.teachingReadmodel.Handler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", deps.assessment.Handler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", deps.teachingReadmodel.Handler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", deps.teachingReadmodel.Handler.GetStudentTimeline)
	teacherOrAbove.GET("/students/:id/evidence", deps.teachingReadmodel.Handler.GetStudentEvidence)
	teacherOrAbove.GET("/students/:id/review-archive", deps.assessment.ReportHandler.GetStudentReviewArchive)
	teacherOrAbove.POST("/students/:id/review-archive/export",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
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
			Action:          model.AuditActionUpdate,
			ResourceType:    "manual_review_submission",
			ResourceIDParam: "id",
		}),
		deps.practice.Handler.ReviewManualReviewSubmission,
	)
	teacherOrAbove.POST("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RecommendCommunity,
	)
	teacherOrAbove.DELETE("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UnrecommendCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/hide",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "community_writeup_visibility",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.HideCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/restore",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
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
	protected.POST("/reports/class", middleware.RequireRole(model.RoleTeacher), deps.assessment.ReportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", deps.assessment.ReportHandler.CreateClassReport)
}
