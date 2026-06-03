package app

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

func registerAdminContestAWDRoutes(
	contestByID *gin.RouterGroup,
	deps adminContestRouteDeps,
	audit contestRouteAuditFunc,
	awdReadinessAudit contestAWDReadinessAuditFunc,
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
