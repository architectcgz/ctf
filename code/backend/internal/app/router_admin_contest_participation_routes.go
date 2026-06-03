package app

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

func registerAdminContestParticipationRoutes(
	contestByID *gin.RouterGroup,
	deps adminContestRouteDeps,
	audit contestRouteAuditFunc,
) {
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
}
