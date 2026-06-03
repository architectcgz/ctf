package app

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/middleware"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type userSelfServiceRouteDeps struct {
	assessment *composition.AssessmentModule
	practice   *composition.PracticeModule
}

func registerUserSelfServiceRoutes(protected *gin.RouterGroup, deps userSelfServiceRouteDeps) {
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
