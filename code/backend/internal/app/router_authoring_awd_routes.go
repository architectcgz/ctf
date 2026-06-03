package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type authoringAWDRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	challenge     *composition.ChallengeModule
}

func registerAuthoringAWDRoutes(adminAuthoring *gin.RouterGroup, deps authoringAWDRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

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
}
