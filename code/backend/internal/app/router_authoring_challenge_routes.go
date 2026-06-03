package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type authoringChallengeRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	challenge     *composition.ChallengeModule
}

func registerAuthoringChallengeRoutes(adminAuthoring *gin.RouterGroup, deps authoringChallengeRouteDeps) {
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
