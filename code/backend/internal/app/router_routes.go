package app

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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
	auditRecorder    auditlog.Recorder
	auditLogger      *zap.Logger
	assessment       *composition.AssessmentModule
	challenge        *composition.ChallengeModule
	contest          *composition.ContestModule
	instance         *composition.InstanceModule
	practice         *composition.PracticeModule
	teachingAnalysis *composition.TeachingAnalysisModule
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
			response.FromError(c, err)
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
	registerAuthoringChallengeRoutes(adminAuthoring, authoringChallengeRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		challenge:     deps.challenge,
	})
	registerAuthoringAssetRoutes(adminAuthoring, authoringAssetRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		challenge:     deps.challenge,
	})
	registerAuthoringAWDRoutes(adminAuthoring, authoringAWDRouteDeps{
		auditRecorder: deps.auditRecorder,
		auditLogger:   deps.auditLogger,
		challenge:     deps.challenge,
	})
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
		auditRecorder:    deps.auditRecorder,
		auditLogger:      deps.auditLogger,
		assessment:       deps.assessment,
		challenge:        deps.challenge,
		instance:         deps.instance,
		practice:         deps.practice,
		teachingAnalysis: deps.teachingAnalysis,
	})
}
