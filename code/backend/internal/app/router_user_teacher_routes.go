package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/middleware"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type teacherRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
	teachingQuery *composition.TeachingQueryModule
}

func registerTeacherRoutes(protected, teacherOrAbove *gin.RouterGroup, deps teacherRouteDeps) {
	audit := func(options middleware.AuditOptions) gin.HandlerFunc {
		return routeAudit(deps.auditRecorder, deps.auditLogger, options)
	}

	protected.GET("/users/:id/skill-profile", middleware.RequireRole(identitycontracts.RoleTeacher), deps.assessment.Handler.GetStudentSkillProfile)

	teacherOrAbove.GET("/overview", deps.teachingQuery.Handler.GetOverview)
	teacherOrAbove.GET("/classes", deps.teachingQuery.Handler.ListClasses)
	teacherOrAbove.GET("/students", deps.teachingQuery.Handler.ListStudents)
	teacherOrAbove.GET("/classes/:name/students", deps.teachingQuery.Handler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", deps.teachingQuery.Handler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", deps.teachingQuery.Handler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", deps.teachingQuery.Handler.GetClassReview)
	teacherOrAbove.GET("/instances", deps.instance.Handler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}),
		deps.instance.Handler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", deps.teachingQuery.Handler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", deps.assessment.Handler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", deps.teachingQuery.Handler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", deps.teachingQuery.Handler.GetStudentTimeline)
	teacherOrAbove.GET("/students/:id/evidence", deps.teachingQuery.Handler.GetStudentEvidence)
	teacherOrAbove.GET("/students/:id/attack-sessions", deps.teachingQuery.Handler.GetStudentAttackSessions)
	teacherOrAbove.GET("/students/:id/review-archive", deps.assessment.ReportHandler.GetStudentReviewArchive)
	teacherOrAbove.POST("/students/:id/review-archive/export",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionAdminOp,
			ResourceType:    "review_archive_export",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}),
		deps.assessment.ReportHandler.CreateStudentReviewArchive,
	)
	teacherOrAbove.GET("/awd/reviews", deps.assessment.TeacherAWDReviewHandler.ListReviews)
	teacherOrAbove.GET("/awd/reviews/:id",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.GetReview,
	)
	teacherOrAbove.POST("/awd/reviews/:id/export/archive",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.ExportArchive,
	)
	teacherOrAbove.POST("/awd/reviews/:id/export/report",
		middleware.ParseInt64Param("id"),
		deps.assessment.TeacherAWDReviewHandler.ExportReport,
	)
	teacherOrAbove.GET("/manual-review-submissions", deps.practice.Handler.ListTeacherManualReviewSubmissions)
	teacherOrAbove.GET("/manual-review-submissions/:id", deps.practice.Handler.GetTeacherManualReviewSubmission)
	teacherOrAbove.PUT("/manual-review-submissions/:id/review",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "manual_review_submission",
			ResourceIDParam: "id",
		}),
		deps.practice.Handler.ReviewManualReviewSubmission,
	)
	teacherOrAbove.POST("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RecommendCommunity,
	)
	teacherOrAbove.DELETE("/community-writeups/:id/recommend",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_recommendation",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.UnrecommendCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/hide",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_visibility",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.HideCommunity,
	)
	teacherOrAbove.POST("/community-writeups/:id/restore",
		audit(middleware.AuditOptions{
			Action:          auditlog.ActionUpdate,
			ResourceType:    "community_writeup_visibility",
			ResourceIDParam: "id",
		}),
		deps.challenge.WriteupHandler.RestoreCommunity,
	)
	teacherOrAbove.GET("/writeup-submissions", deps.challenge.WriteupHandler.ListTeacherSubmissions)
	teacherOrAbove.GET("/writeup-submissions/:id", deps.challenge.WriteupHandler.GetTeacherSubmission)
	teacherOrAbove.POST("/reports/class", deps.assessment.ReportHandler.CreateClassReport)
}
