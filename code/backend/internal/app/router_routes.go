package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	"ctf-platform/internal/middleware"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type adminRouteDeps struct {
	identityHandler *identityhttp.Handler
	auditRecorder   auditlog.Recorder
	auditLogger     *zap.Logger
	assessment      *composition.AssessmentModule
	challenge       *composition.ChallengeModule
	contest         *composition.ContestModule
	ops             *composition.OpsModule
	practice        *composition.PracticeModule
	tokenService    authcontracts.TokenService
}

type userRouteDeps struct {
	auditRecorder auditlog.Recorder
	auditLogger   *zap.Logger
	assessment    *composition.AssessmentModule
	challenge     *composition.ChallengeModule
	contest       *composition.ContestModule
	instance      *composition.InstanceModule
	practice      *composition.PracticeModule
	teachingQuery *composition.TeachingQueryModule
}

type challengeLookup interface {
	FindByID(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error)
}

func routeAudit(recorder auditlog.Recorder, logger *zap.Logger, options middleware.AuditOptions) gin.HandlerFunc {
	return middleware.Audit(recorder, options, logger)
}

func challengeOwnerGuard(catalog challengeLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := authctx.MustCurrentUser(c)
		if currentUser.Role == identitycontracts.RoleAdmin {
			c.Next()
			return
		}

		challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.InvalidParams(c, "无效的ID")
			c.Abort()
			return
		}

		challenge, err := catalog.FindByID(c.Request.Context(), challengeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, challengecontracts.ErrChallengeNotFound)
			} else {
				response.FromError(c, err)
			}
			c.Abort()
			return
		}
		if challenge.CreatedBy == nil || *challenge.CreatedBy != currentUser.UserID {
			response.Error(c, apperror.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func registerTeacherAuthoringRoutes(adminAuthoring *gin.RouterGroup, deps adminRouteDeps) {
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

func registerAdminRoutes(adminOnly *gin.RouterGroup, deps adminRouteDeps) {
	registerAdminOpsRoutes(adminOnly, adminOpsRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		ops:           deps.ops,
	})
	registerAdminIdentityRoutes(adminOnly, adminIdentityRouteDeps{
		auditRecorder:   deps.auditRecorder,
		auditLogger:     deps.auditLogger,
		identityHandler: deps.identityHandler,
		tokenService:    deps.tokenService,
	})

	registerAdminContestRoutes(adminOnly, adminContestRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		assessment:    deps.assessment,
		contest:       deps.contest,
		practice:      deps.practice,
	})
}

func registerUserRoutes(apiV1, protected, teacherOrAbove *gin.RouterGroup, deps userRouteDeps) {
	registerUserSelfRoutes(apiV1, protected, userSelfRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		assessment:    deps.assessment,
		challenge:     deps.challenge,
		contest:       deps.contest,
		instance:      deps.instance,
		practice:      deps.practice,
	})
	registerTeacherRoutes(protected, teacherOrAbove, teacherRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		assessment:    deps.assessment,
		challenge:     deps.challenge,
		instance:      deps.instance,
		practice:      deps.practice,
		teachingQuery: deps.teachingQuery,
	})
}
