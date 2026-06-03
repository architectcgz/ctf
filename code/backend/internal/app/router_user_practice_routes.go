package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type userPracticeRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	challenge     *composition.ChallengeModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
}

func registerUserPracticeRoutes(apiV1, protected *gin.RouterGroup, deps userPracticeRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

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
}
