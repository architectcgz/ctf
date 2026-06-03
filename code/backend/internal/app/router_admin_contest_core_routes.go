package app

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

func registerAdminContestCoreRoutes(
	contests, contestByID *gin.RouterGroup,
	deps adminContestRouteDeps,
	audit contestRouteAuditFunc,
	awdReadinessAudit contestAWDReadinessAuditFunc,
) {
	contests.POST("",
		audit(middleware.AuditOptions{
			Action:       auditlog.ActionCreate,
			ResourceType: "contest",
		}),
		deps.contest.Handler.CreateContest,
	)
	contests.GET("", deps.contest.Handler.ListContests)
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
}
