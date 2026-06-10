package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type userContestRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	contest       *composition.ContestModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
}

func registerUserContestRoutes(apiV1, protected *gin.RouterGroup, deps userContestRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	contestGroup := apiV1.Group("/contests")
	contestGroup.GET("", deps.contest.Handler.ListContests)
	contestGroup.GET("/:id", middleware.ParseInt64Param("id"), deps.contest.Handler.GetContest)
	contestGroup.GET("/:id/scoreboard", deps.contest.Handler.GetScoreboard)
	contestGroup.GET("/:id/announcements", deps.contest.ParticipationHandler.ListAnnouncements)
	contestGroup.GET("/:id/announcements/sync", deps.contest.ParticipationHandler.SyncAnnouncements)

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
}
