package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type adminContestRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	contest       *composition.ContestModule
	practice      *composition.PracticeModule
}

func registerAdminContestRoutes(adminOnly *gin.RouterGroup, deps adminContestRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}
	awdReadinessAudit := func() gin.HandlerFunc {
		return middleware.AWDReadinessAudit(deps.auditRecorder, deps.auditLogger)
	}

	contests := adminOnly.Group("/contests")
	contests.POST("",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "contest",
		}),
		deps.contest.Handler.CreateContest,
	)
	contests.GET("", deps.contest.Handler.ListContests)

	contestByID := contests.Group("/:id")
	contestByID.Use(middleware.ParseInt64Param("id"))
	contestByID.GET("", deps.contest.Handler.GetContest)
	contestByID.PUT("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest",
			ResourceIDParam: "id",
		}),
		awdReadinessAudit(),
		deps.contest.Handler.UpdateContest,
	)
	contestByID.POST("/freeze",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contest.Handler.FreezeScoreboard,
	)
	contestByID.POST("/unfreeze",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.contest.Handler.UnfreezeScoreboard,
	)
	contestByID.POST("/export",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "contest_export",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.assessment.ReportHandler.CreateContestExport,
	)
	contestByID.GET("/challenges", deps.contest.ChallengeHandler.ListAdminChallenges)
	contestByID.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_challenge",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ChallengeHandler.AddChallenge,
	)
	contestByID.PUT("/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.UpdatePoints,
	)
	contestByID.DELETE("/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.RemoveChallenge,
	)
	contestByID.GET("/registrations", deps.contest.ParticipationHandler.ListRegistrations)
	contestByID.PUT("/registrations/:rid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest_registration",
			ResourceIDParam: "rid",
			DetailBuilder:   middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.ParticipationHandler.ReviewRegistration,
	)
	contestByID.GET("/announcements", deps.contest.ParticipationHandler.ListAnnouncements)
	contestByID.POST("/announcements",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_announcement",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ParticipationHandler.CreateAnnouncement,
	)
	contestByID.DELETE("/announcements/:aid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "contest_announcement",
			ResourceIDParam: "aid",
			DetailBuilder:   middleware.DetailFromParams("id", "aid"),
		}),
		deps.contest.ParticipationHandler.DeleteAnnouncement,
	)
	contestByID.GET("/scoreboard/live", deps.contest.Handler.GetLiveScoreboard)

	registerAdminContestAWDRoutes(contestByID, deps, audit, awdReadinessAudit)
}

func registerAdminContestAWDRoutes(
	contestByID *gin.RouterGroup,
	deps adminContestRouteDeps,
	audit func(middleware.AuditOptions) gin.HandlerFunc,
	awdReadinessAudit func() gin.HandlerFunc,
) {
	awd := contestByID.Group("/awd")
	awd.GET("/rounds", deps.contest.AWDHandler.ListRounds)
	awd.GET("/readiness", deps.contest.AWDHandler.GetReadiness)
	awd.GET("/services", deps.contest.AWDHandler.ListContestAWDServices)
	awd.GET("/instances", deps.practice.Handler.GetAdminContestAWDInstanceOrchestration)
	awd.POST("/instances",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_awd_instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.practice.Handler.StartAdminContestAWDInstance,
	)
	awd.POST("/instances/prewarm",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_awd_instance_prewarm",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.practice.Handler.PrewarmAdminContestAWDInstances,
	)

	awdTeams := awd.Group("/teams/:team_id")
	awdTeams.Use(middleware.ParseInt64Param("team_id"))
	awdTeams.PUT("/retirement",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "contest_awd_team_retirement",
			DetailBuilder: middleware.DetailFromParams("id", "team_id"),
		}),
		deps.practice.Handler.SetAdminContestAWDTeamRetired,
	)

	awdTeamServices := awdTeams.Group("/services/:sid")
	awdTeamServices.Use(middleware.ParseInt64Param("sid"))
	awdTeamServices.PUT("/disabled",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "contest_awd_service_disable",
			DetailBuilder: middleware.DetailFromParams("id", "team_id", "sid"),
		}),
		deps.practice.Handler.SetAdminContestAWDTeamServiceDisabled,
	)
	awdTeamServices.PUT("/suppression",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "contest_awd_desired_reconcile_suppression",
			DetailBuilder: middleware.DetailFromParams("id", "team_id", "sid"),
		}),
		deps.practice.Handler.SetAdminContestAWDDesiredReconcileSuppressed,
	)

	awdServices := awd.Group("/services")
	awdServices.POST("",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_awd_service",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.AWDHandler.CreateContestAWDService,
	)

	awdServiceByID := awdServices.Group("/:sid")
	awdServiceByID.Use(middleware.ParseInt64Param("sid"))
	awdServiceByID.PUT("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest_awd_service",
			ResourceIDParam: "sid",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		deps.contest.AWDHandler.UpdateContestAWDService,
	)
	awdServiceByID.DELETE("",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "contest_awd_service",
			ResourceIDParam: "sid",
			DetailBuilder:   middleware.DetailFromParams("id", "sid"),
		}),
		deps.contest.AWDHandler.DeleteContestAWDService,
	)

	awd.POST("/rounds",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "awd_round",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		awdReadinessAudit(),
		deps.contest.AWDHandler.CreateRound,
	)
	awd.POST("/current-round/check",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		awdReadinessAudit(),
		deps.contest.AWDHandler.RunCurrentRoundChecks,
	)
	awd.POST("/checker-preview",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "awd_checker_preview",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.AWDHandler.PreviewChecker,
	)

	awdRounds := awd.Group("/rounds/:rid")
	awdRounds.Use(middleware.ParseInt64Param("rid"))
	awdRounds.POST("/check",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.RunRoundChecks,
	)
	awdRounds.GET("/services", deps.contest.AWDHandler.ListServices)
	awdRounds.POST("/services/check",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionUpdate,
			ResourceType:  "awd_service_check",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.UpsertServiceCheck,
	)
	awdRounds.GET("/attacks", deps.contest.AWDHandler.ListAttackLogs)
	awdRounds.GET("/traffic/summary", deps.contest.AWDHandler.GetTrafficSummary)
	awdRounds.GET("/traffic/events", deps.contest.AWDHandler.ListTrafficEvents)
	awdRounds.POST("/attacks",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "awd_attack_log",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}),
		deps.contest.AWDHandler.CreateAttackLog,
	)
	awdRounds.GET("/summary", deps.contest.AWDHandler.GetRoundSummary)
}
