package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	adminUserModule "ctf-platform/internal/module/adminuser"
	assessmentModule "ctf-platform/internal/module/assessment"
	challengeModule "ctf-platform/internal/module/challenge"
	containerModule "ctf-platform/internal/module/container"
	contestModule "ctf-platform/internal/module/contest"
	practiceModule "ctf-platform/internal/module/practice"
	systemModule "ctf-platform/internal/module/system"
	teacherModule "ctf-platform/internal/module/teacher"
)

type adminRouteDeps struct {
	auditRecorder           auditlog.Recorder
	auditLogger             *zap.Logger
	imageHandler            *challengeModule.ImageHandler
	challengeHandler        *challengeModule.Handler
	writeupHandler          *challengeModule.WriteupHandler
	topologyHandler         *challengeModule.TopologyHandler
	flagHandler             *challengeModule.FlagHandler
	auditHandler            *systemModule.AuditHandler
	dashboardHandler        *systemModule.DashboardHandler
	riskHandler             *systemModule.RiskHandler
	adminUserHandler        *adminUserModule.Handler
	contestHandler          *contestModule.Handler
	awdHandler              *contestModule.AWDHandler
	contestChallengeHandler *contestModule.ChallengeHandler
	participationHandler    *contestModule.ParticipationHandler
}

type userRouteDeps struct {
	auditRecorder           auditlog.Recorder
	auditLogger             *zap.Logger
	challengeHandler        *challengeModule.Handler
	writeupHandler          *challengeModule.WriteupHandler
	practiceHandler         *practiceModule.Handler
	containerHandler        *containerModule.Handler
	assessmentHandler       *assessmentModule.Handler
	teacherHandler          *teacherModule.Handler
	reportHandler           *assessmentModule.ReportHandler
	contestHandler          *contestModule.Handler
	awdHandler              *contestModule.AWDHandler
	contestChallengeHandler *contestModule.ChallengeHandler
	participationHandler    *contestModule.ParticipationHandler
	submissionHandler       *contestModule.SubmissionHandler
	teamHandler             *contestModule.TeamHandler
}

func routeAudit(recorder auditlog.Recorder, logger *zap.Logger, options middleware.AuditOptions) gin.HandlerFunc {
	return middleware.Audit(recorder, options, logger)
}

func registerAdminRoutes(adminOnly *gin.RouterGroup, deps adminRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	adminOnly.POST("/images",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "image",
		}),
		deps.imageHandler.CreateImage,
	)
	adminOnly.GET("/images", deps.imageHandler.ListImages)
	adminOnly.GET("/images/:id", deps.imageHandler.GetImage)
	adminOnly.PUT("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.imageHandler.UpdateImage,
	)
	adminOnly.DELETE("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.imageHandler.DeleteImage,
	)

	adminOnly.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "challenge",
		}),
		deps.challengeHandler.CreateChallenge,
	)
	adminOnly.GET("/challenges", deps.challengeHandler.ListChallenges)
	adminOnly.GET("/challenges/:id", deps.challengeHandler.GetChallenge)
	adminOnly.PUT("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challengeHandler.UpdateChallenge,
	)
	adminOnly.DELETE("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challengeHandler.DeleteChallenge,
	)
	adminOnly.PUT("/challenges/:id/publish",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challengeHandler.PublishChallenge,
	)
	adminOnly.GET("/challenges/:id/writeup", deps.writeupHandler.GetAdmin)
	adminOnly.PUT("/challenges/:id/writeup",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.writeupHandler.Upsert,
	)
	adminOnly.DELETE("/challenges/:id/writeup",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.writeupHandler.Delete,
	)
	adminOnly.GET("/challenges/:id/topology", deps.topologyHandler.GetChallengeTopology)
	adminOnly.PUT("/challenges/:id/topology",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.topologyHandler.SaveChallengeTopology,
	)
	adminOnly.DELETE("/challenges/:id/topology",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.topologyHandler.DeleteChallengeTopology,
	)
	adminOnly.GET("/environment-templates", deps.topologyHandler.ListTemplates)
	adminOnly.POST("/environment-templates",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "environment_template",
		}),
		deps.topologyHandler.CreateTemplate,
	)
	adminOnly.GET("/environment-templates/:id", deps.topologyHandler.GetTemplate)
	adminOnly.PUT("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.topologyHandler.UpdateTemplate,
	)
	adminOnly.DELETE("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.topologyHandler.DeleteTemplate,
	)

	adminOnly.PUT("/challenges/:id/flag",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_flag",
			ResourceIDParam: "id",
		}),
		deps.flagHandler.ConfigureFlag,
	)
	adminOnly.GET("/challenges/:id/flag", deps.flagHandler.GetFlagConfig)
	adminOnly.GET("/audit-logs", deps.auditHandler.ListAuditLogs)
	adminOnly.GET("/dashboard", deps.dashboardHandler.GetDashboard)
	adminOnly.GET("/cheat-detection", deps.riskHandler.GetCheatDetection)

	adminOnly.GET("/users", deps.adminUserHandler.ListUsers)
	adminOnly.POST("/users",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user",
		}),
		deps.adminUserHandler.CreateUser,
	)
	adminOnly.PUT("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.adminUserHandler.UpdateUser,
	)
	adminOnly.DELETE("/users/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}),
		deps.adminUserHandler.DeleteUser,
	)
	adminOnly.POST("/users/import",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user_import",
		}),
		deps.adminUserHandler.ImportUsers,
	)

	adminOnly.POST("/contests",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "contest",
		}),
		deps.contestHandler.CreateContest,
	)
	adminOnly.PUT("/contests/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest",
			ResourceIDParam: "id",
		}),
		deps.contestHandler.UpdateContest,
	)
	adminOnly.GET("/contests", deps.contestHandler.ListContests)
	adminOnly.POST("/contests/:id/freeze",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contestHandler.FreezeScoreboard,
	)
	adminOnly.POST("/contests/:id/unfreeze",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contestHandler.UnfreezeScoreboard,
	)
	adminOnly.GET("/contests/:id/challenges", deps.contestChallengeHandler.ListAdminChallenges)
	adminOnly.POST("/contests/:id/challenges",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_challenge",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contestChallengeHandler.AddChallenge,
	)
	adminOnly.PUT("/contests/:id/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contestChallengeHandler.UpdatePoints,
	)
	adminOnly.DELETE("/contests/:id/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contestChallengeHandler.RemoveChallenge,
	)
	adminOnly.GET("/contests/:id/registrations", deps.participationHandler.ListRegistrations)
	adminOnly.PUT("/contests/:id/registrations/:rid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_registration",
			ResourceIDParam: "rid",
			DetailBuilder:   middleware.DetailFromParams("id", "rid"),
		}),
		deps.participationHandler.ReviewRegistration,
	)
	adminOnly.GET("/contests/:id/announcements", deps.participationHandler.ListAnnouncements)
	adminOnly.POST("/contests/:id/announcements",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_announcement",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.participationHandler.CreateAnnouncement,
	)
	adminOnly.DELETE("/contests/:id/announcements/:aid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_announcement",
			ResourceIDParam: "aid",
			DetailBuilder:   middleware.DetailFromParams("id", "aid"),
		}),
		deps.participationHandler.DeleteAnnouncement,
	)
	adminOnly.GET("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		deps.awdHandler.ListRounds,
	)
	adminOnly.GET("/contests/:id/scoreboard/live", deps.contestHandler.GetLiveScoreboard)
	adminOnly.POST("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_round",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.awdHandler.CreateRound,
	)
	adminOnly.POST("/contests/:id/awd/current-round/check",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.awdHandler.RunCurrentRoundChecks,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.awdHandler.RunRoundChecks,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/services",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.awdHandler.ListServices,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/services/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_service_check",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.awdHandler.UpsertServiceCheck,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.awdHandler.ListAttackLogs,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_attack_log",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.awdHandler.CreateAttackLog,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/summary",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		deps.awdHandler.GetRoundSummary,
	)
}

func registerUserRoutes(apiV1, protected, teacherOrAbove *gin.RouterGroup, deps userRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	contestGroup := apiV1.Group("/contests")
	contestGroup.GET("", deps.contestHandler.ListContests)
	contestGroup.GET("/:id", middleware.ParseInt64Param("id"), deps.contestHandler.GetContest)
	contestGroup.GET("/:id/scoreboard", deps.contestHandler.GetScoreboard)
	contestGroup.GET("/:id/announcements", deps.participationHandler.ListAnnouncements)

	protected.POST("/contests/:id/register",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_registration",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.participationHandler.RegisterContest,
	)
	protected.GET("/contests/:id/challenges", deps.contestChallengeHandler.ListChallenges)
	protected.GET("/contests/:id/my-progress", deps.participationHandler.GetMyProgress)
	protected.POST("/contests/:id/challenges/:cid/submissions",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "contest_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.submissionHandler.SubmitFlag,
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
		deps.awdHandler.SubmitAttack,
	)
	protected.GET("/contests/:id/teams", deps.teamHandler.ListTeams)
	protected.GET("/contests/:id/my-team", deps.teamHandler.GetMyTeam)
	protected.POST("/contests/:id/teams",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "team",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.teamHandler.CreateTeam,
	)
	protected.POST("/contests/:id/teams/:tid/join",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.teamHandler.JoinTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/leave",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.teamHandler.LeaveTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}),
		deps.teamHandler.DismissTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/members/:uid",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "uid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid", "uid"),
		}),
		deps.teamHandler.KickMember,
	)

	protected.GET("/challenges", deps.challengeHandler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionRead,
			ResourceType:    "challenge_detail",
			ResourceIDParam: "id",
		}),
		deps.challengeHandler.GetPublishedChallenge,
	)
	protected.GET("/challenges/:id/writeup", deps.writeupHandler.GetPublished)
	protected.POST("/challenges/:id/instances",
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.practiceHandler.StartChallenge,
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
		deps.practiceHandler.StartContestChallenge,
	)
	protected.POST("/challenges/:id/submit",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}),
		deps.practiceHandler.SubmitFlag,
	)
	protected.POST("/challenges/:id/hints/:level/unlock",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "challenge_hint_unlock",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id", "level"),
		}),
		deps.practiceHandler.UnlockHint,
	)
	protected.GET("/instances", deps.practiceHandler.ListUserInstances)
	protected.GET("/instances/:id", deps.practiceHandler.GetInstance)
	protected.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.containerHandler.DestroyInstance,
	)
	protected.POST("/instances/:id/extend",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.containerHandler.ExtendInstance,
	)
	protected.POST("/instances/:id/access", deps.containerHandler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", deps.containerHandler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", deps.containerHandler.ProxyInstance)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", deps.practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", deps.practiceHandler.GetTimeline)
	usersGroup.GET("/me/skill-profile", deps.assessmentHandler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", deps.assessmentHandler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(model.RoleTeacher), deps.assessmentHandler.GetStudentSkillProfile)

	teacherOrAbove.GET("/classes", deps.teacherHandler.ListClasses)
	teacherOrAbove.GET("/classes/:name/students", deps.teacherHandler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", deps.teacherHandler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", deps.teacherHandler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", deps.teacherHandler.GetClassReview)
	teacherOrAbove.GET("/instances", deps.containerHandler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.containerHandler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", deps.teacherHandler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", deps.assessmentHandler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", deps.teacherHandler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", deps.teacherHandler.GetStudentTimeline)

	protected.POST("/reports/personal", deps.reportHandler.CreatePersonalReport)
	protected.GET("/reports/:id", deps.reportHandler.GetReportStatus)
	protected.GET("/reports/:id/download", deps.reportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(model.RoleTeacher), deps.reportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", deps.reportHandler.CreateClassReport)
}
