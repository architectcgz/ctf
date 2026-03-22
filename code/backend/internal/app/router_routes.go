package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	adminUserModule "ctf-platform/internal/module/adminuser"
)

type adminRouteDeps struct {
	adminUserHandler *adminUserModule.Handler
	auditRecorder    auditlog.Recorder
	auditLogger      *zap.Logger
	challenge        *composition.ChallengeModule
	contest          *composition.ContestModule
	system           *composition.SystemModule
}

type userRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	container     *composition.ContainerModule
	contest       *composition.ContestModule
	practice      *composition.PracticeModule
	teacher       *composition.TeacherModule
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
		deps.challenge.ImageHandler.CreateImage,
	)
	adminOnly.GET("/images", deps.challenge.ImageHandler.ListImages)
	adminOnly.GET("/images/:id", deps.challenge.ImageHandler.GetImage)
	adminOnly.PUT("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.UpdateImage,
	)
	adminOnly.DELETE("/images/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.DeleteImage,
	)

	adminOnly.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "challenge",
		}),
		deps.challenge.Handler.CreateChallenge,
	)
	adminOnly.GET("/challenges", deps.challenge.Handler.ListChallenges)
	adminOnly.GET("/challenges/:id", deps.challenge.Handler.GetChallenge)
	adminOnly.PUT("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.UpdateChallenge,
	)
	adminOnly.DELETE("/challenges/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.DeleteChallenge,
	)
	adminOnly.PUT("/challenges/:id/publish",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}),
		deps.challenge.Handler.PublishChallenge,
	)
	adminOnly.GET("/challenges/:id/writeup", deps.challenge.WriteupHandler.GetAdmin)
	adminOnly.PUT("/challenges/:id/writeup",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Upsert,
	)
	adminOnly.DELETE("/challenges/:id/writeup",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.Delete,
	)
	adminOnly.GET("/challenges/:id/topology", deps.challenge.TopologyHandler.GetChallengeTopology)
	adminOnly.PUT("/challenges/:id/topology",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.SaveChallengeTopology,
	)
	adminOnly.DELETE("/challenges/:id/topology",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteChallengeTopology,
	)
	adminOnly.GET("/environment-templates", deps.challenge.TopologyHandler.ListTemplates)
	adminOnly.POST("/environment-templates",
		audit(middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "environment_template",
		}),
		deps.challenge.TopologyHandler.CreateTemplate,
	)
	adminOnly.GET("/environment-templates/:id", deps.challenge.TopologyHandler.GetTemplate)
	adminOnly.PUT("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.UpdateTemplate,
	)
	adminOnly.DELETE("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteTemplate,
	)

	adminOnly.PUT("/challenges/:id/flag",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_flag",
			ResourceIDParam: "id",
		}),
		deps.challenge.FlagHandler.ConfigureFlag,
	)
	adminOnly.GET("/challenges/:id/flag", deps.challenge.FlagHandler.GetFlagConfig)
	adminOnly.GET("/audit-logs", deps.system.AuditHandler.ListAuditLogs)
	adminOnly.GET("/dashboard", deps.system.DashboardHandler.GetDashboard)
	adminOnly.GET("/cheat-detection", deps.system.RiskHandler.GetCheatDetection)

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
		deps.contest.Handler.CreateContest,
	)
	adminOnly.PUT("/contests/:id",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest",
			ResourceIDParam: "id",
		}),
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
	adminOnly.GET("/contests/:id/scoreboard/live", deps.contest.Handler.GetLiveScoreboard)
	adminOnly.POST("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_round",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.AWDHandler.CreateRound,
	)
	adminOnly.POST("/contests/:id/awd/current-round/check",
		middleware.ParseInt64Param("id"),
		audit(middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.AWDHandler.RunCurrentRoundChecks,
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
	protected.POST("/challenges/:id/hints/:level/unlock",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "challenge_hint_unlock",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id", "level"),
		}),
		deps.practice.Handler.UnlockHint,
	)
	protected.GET("/instances", deps.practice.Handler.ListUserInstances)
	protected.GET("/instances/:id", deps.practice.Handler.GetInstance)
	protected.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.container.Handler.DestroyInstance,
	)
	protected.POST("/instances/:id/extend",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.container.Handler.ExtendInstance,
	)
	protected.POST("/instances/:id/access", deps.container.Handler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", deps.container.Handler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", deps.container.Handler.ProxyInstance)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", deps.practice.Handler.GetProgress)
	usersGroup.GET("/me/timeline", deps.practice.Handler.GetTimeline)
	usersGroup.GET("/me/skill-profile", deps.assessment.Handler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", deps.assessment.Handler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(model.RoleTeacher), deps.assessment.Handler.GetStudentSkillProfile)

	teacherOrAbove.GET("/classes", deps.teacher.Handler.ListClasses)
	teacherOrAbove.GET("/classes/:name/students", deps.teacher.Handler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", deps.teacher.Handler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", deps.teacher.Handler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", deps.teacher.Handler.GetClassReview)
	teacherOrAbove.GET("/instances", deps.container.Handler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.container.Handler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", deps.teacher.Handler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", deps.assessment.Handler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", deps.teacher.Handler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", deps.teacher.Handler.GetStudentTimeline)

	protected.POST("/reports/personal", deps.assessment.ReportHandler.CreatePersonalReport)
	protected.GET("/reports/:id", deps.assessment.ReportHandler.GetReportStatus)
	protected.GET("/reports/:id/download", deps.assessment.ReportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(model.RoleTeacher), deps.assessment.ReportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", deps.assessment.ReportHandler.CreateClassReport)
}
