package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

type contestRouteAuditFunc func(middleware.AuditOptions) gin.HandlerFunc
type contestAWDReadinessAuditFunc func() gin.HandlerFunc

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
	contestByID := contests.Group("/:id")
	contestByID.Use(middleware.ParseInt64Param("id"))

	registerAdminContestCoreRoutes(contests, contestByID, deps, audit, awdReadinessAudit)
	registerAdminContestChallengeRoutes(contestByID, deps, audit)
	registerAdminContestParticipationRoutes(contestByID, deps, audit)
	registerAdminContestAWDRoutes(contestByID, deps, audit, awdReadinessAudit)
}
