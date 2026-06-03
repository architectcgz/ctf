package app

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
)

func registerAdminContestChallengeRoutes(
	contestByID *gin.RouterGroup,
	deps adminContestRouteDeps,
	audit contestRouteAuditFunc,
) {
	contestByID.GET("/challenges", deps.contest.ChallengeHandler.ListAdminChallenges)
	contestByID.POST("/challenges",
		audit(middleware.AuditOptions{
			Action:        auditlog.ActionCreate,
			ResourceType:  "contest_challenge",
			DetailBuilder: middleware.DetailFromParams("id"),
		}),
		deps.contest.ChallengeHandler.AddChallenge,
	)
	contestByID.PUT("/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.UpdatePoints,
	)
	contestByID.DELETE("/challenges/:cid",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}),
		deps.contest.ChallengeHandler.RemoveChallenge,
	)
}
