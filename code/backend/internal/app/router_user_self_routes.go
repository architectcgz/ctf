package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type userSelfRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	contest       *composition.ContestModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
}

func registerUserSelfRoutes(apiV1, protected *gin.RouterGroup, deps userSelfRouteDeps) {
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

	protected.POST("/reports/personal", deps.assessment.ReportHandler.CreatePersonalReport)
	protected.GET("/reports/:id", deps.assessment.ReportHandler.GetReportStatus)
	protected.GET("/reports/:id/download", deps.assessment.ReportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(identitycontracts.RoleTeacher), deps.assessment.ReportHandler.CreateClassReport)
}
