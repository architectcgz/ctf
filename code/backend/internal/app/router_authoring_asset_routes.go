package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type authoringAssetRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	challenge     *composition.ChallengeModule
}

func registerAuthoringAssetRoutes(adminAuthoring *gin.RouterGroup, deps authoringAssetRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	adminAuthoring.POST("/images",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "image",
		}),
		deps.challenge.ImageHandler.CreateImage,
	)
	adminAuthoring.GET("/images", deps.challenge.ImageHandler.ListImages)
	adminAuthoring.GET("/images/:id", deps.challenge.ImageHandler.GetImage)
	adminAuthoring.PUT("/images/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.UpdateImage,
	)
	adminAuthoring.DELETE("/images/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}),
		deps.challenge.ImageHandler.DeleteImage,
	)

	adminAuthoring.GET("/environment-templates", deps.challenge.TopologyHandler.ListTemplates)
	adminAuthoring.POST("/environment-templates",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "environment_template",
		}),
		deps.challenge.TopologyHandler.CreateTemplate,
	)
	adminAuthoring.GET("/environment-templates/:id", deps.challenge.TopologyHandler.GetTemplate)
	adminAuthoring.PUT("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.UpdateTemplate,
	)
	adminAuthoring.DELETE("/environment-templates/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}),
		deps.challenge.TopologyHandler.DeleteTemplate,
	)
}
